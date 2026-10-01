package mongooutbox

import(
 "context"
 "errors"
 "time"
 "github.com/el211/GoModulith/durable"
 "github.com/el211/GoModulith/durable/reliable"
 "go.mongodb.org/mongo-driver/bson"
 "go.mongodb.org/mongo-driver/mongo"
 "go.mongodb.org/mongo-driver/mongo/options"
)
type document struct{
 ID string `bson:"_id"`
 Topic string `bson:"topic"`
 Payload []byte `bson:"payload"`
 CreatedAt time.Time `bson:"created_at"`
 Attempts int `bson:"attempts"`
 AvailableAt time.Time `bson:"available_at"`
 LeaseUntil time.Time `bson:"lease_until"`
 Owner string `bson:"owner"`
 State string `bson:"state"`
 LastError string `bson:"last_error"`
}
// Store is a MongoDB-backed reliable outbox.
type Store struct{ Collection *mongo.Collection }
var ErrLease=errors.New("mongooutbox: lease not owned or message missing")
func(s Store)check()error{if s.Collection==nil{return errors.New("mongooutbox: nil collection")};return nil}
// EnsureIndexes prepares the claim index.
func(s Store)EnsureIndexes(ctx context.Context)error{
 if err:=s.check();err!=nil{return err}
 _,err:=s.Collection.Indexes().CreateOne(ctx,mongo.IndexModel{Keys:bson.D{{Key:"state",Value:1},{Key:"available_at",Value:1},{Key:"lease_until",Value:1}}});return err
}
// EnqueueSession inserts using the passed context; if it is a session context
// with an active transaction the enqueue joins that business transaction.
func(s Store)EnqueueSession(ctx mongo.SessionContext,e durable.Event)error{return s.Enqueue(ctx,e)}
func(s Store)Enqueue(ctx context.Context,e durable.Event)error{
 if err:=s.check();err!=nil{return err};if e.ID==""||e.Topic==""{return errors.New("mongooutbox: ID and topic required")}
 now:=e.CreatedAt;if now.IsZero(){now=time.Now().UTC()}
 _,err:=s.Collection.InsertOne(ctx,document{ID:e.ID,Topic:e.Topic,Payload:e.Payload,CreatedAt:now,AvailableAt:now,LeaseUntil:time.Unix(0,0).UTC(),State:"pending"});return err
}
// Claim uses FindOneAndUpdate for atomic lease acquisition.
func(s Store)Claim(ctx context.Context,worker string,now time.Time,lease time.Duration,limit int)([]reliable.Message,error){
 if err:=s.check();err!=nil{return nil,err};if worker==""||lease<=0||limit<=0{return nil,errors.New("mongooutbox: invalid claim arguments")}
 out:=make([]reliable.Message,0,limit)
 filter:=bson.M{"state":"pending","available_at":bson.M{"$lte":now},"lease_until":bson.M{"$lte":now}}
 update:=bson.M{"$set":bson.M{"owner":worker,"lease_until":now.Add(lease)}}
 for i:=0;i<limit;i++{
  var d document
  err:=s.Collection.FindOneAndUpdate(ctx,filter,update,options.FindOneAndUpdate().SetSort(bson.D{{Key:"available_at",Value:1},{Key:"_id",Value:1}}).SetReturnDocument(options.After)).Decode(&d)
  if errors.Is(err,mongo.ErrNoDocuments){break};if err!=nil{return out,err}
  out=append(out,reliable.Message{Event:durable.Event{ID:d.ID,Topic:d.Topic,Payload:d.Payload,CreatedAt:d.CreatedAt},Attempts:d.Attempts,AvailableAt:d.AvailableAt,LeaseUntil:d.LeaseUntil,LastError:d.LastError})
 }
 return out,nil
}
func(s Store)update(ctx context.Context,id,worker string,v any)error{
 if err:=s.check();err!=nil{return err}
 result,err:=s.Collection.UpdateOne(ctx,bson.M{"_id":id,"owner":worker,"state":"pending"},bson.M{"$set":v});if err!=nil{return err};if result.MatchedCount!=1{return ErrLease};return nil
}
func(s Store)Ack(ctx context.Context,id,worker string)error{
 if err:=s.check();err!=nil{return err};r,err:=s.Collection.DeleteOne(ctx,bson.M{"_id":id,"owner":worker,"state":"pending"});if err!=nil{return err};if r.DeletedCount!=1{return ErrLease};return nil
}
func(s Store)Retry(ctx context.Context,id,worker string,attempt int,next time.Time,reason string)error{
 return s.update(ctx,id,worker,bson.M{"attempts":attempt,"available_at":next,"last_error":reason,"owner":"","lease_until":time.Unix(0,0).UTC()})
}
func(s Store)Dead(ctx context.Context,id,worker string,attempt int,reason string)error{
 return s.update(ctx,id,worker,bson.M{"attempts":attempt,"last_error":reason,"state":"dead","owner":"","lease_until":time.Unix(0,0).UTC()})
}
var _ reliable.Store=Store{}

// DeadLetters returns a bounded list of exhausted events for inspection.
func(s Store)DeadLetters(ctx context.Context,limit int)([]reliable.DeadLetter,error){
 if err:=s.check();err!=nil{return nil,err};if limit<=0{return nil,nil}
 cur,err:=s.Collection.Find(ctx,bson.M{"state":"dead"},options.Find().SetSort(bson.D{{Key:"created_at",Value:1},{Key:"_id",Value:1}}).SetLimit(int64(limit)))
 if err!=nil{return nil,err};defer cur.Close(ctx)
 out:=[]reliable.DeadLetter{}
 for cur.Next(ctx){var d document;if err:=cur.Decode(&d);err!=nil{return nil,err}
  out=append(out,reliable.DeadLetter{Event:durable.Event{ID:d.ID,Topic:d.Topic,Payload:d.Payload,CreatedAt:d.CreatedAt},Attempts:d.Attempts,Reason:d.LastError})
 }
 return out,cur.Err()
}
// RequeueDead resets an exhausted event for a new delivery attempt.
func(s Store)RequeueDead(ctx context.Context,id string)error{
 if err:=s.check();err!=nil{return err}
 res,err:=s.Collection.UpdateOne(ctx,bson.M{"_id":id,"state":"dead"},bson.M{"$set":bson.M{"state":"pending","attempts":0,"last_error":"","owner":"","lease_until":time.Unix(0,0).UTC(),"available_at":time.Now().UTC()}})
 if err!=nil{return err};if res.MatchedCount!=1{return ErrLease};return nil
}
var _ reliable.DeadLetterStore=Store{}
