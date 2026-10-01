package durable

import ("context";"errors";"time")

// Event is a transport-neutral serialized envelope.
type Event struct {
 ID string
 Topic string
 Payload []byte
 CreatedAt time.Time
}
// Store is an outbox persistence port. Enqueue should participate in the same
// transaction as the application's business mutation when possible.
type Store interface{
 Enqueue(context.Context,Event)error
 Pending(context.Context,int)([]Event,error)
 Acknowledge(context.Context,string)error
}
// Publisher publishes a stored message to a broker or in-process bus.
type Publisher interface{PublishEvent(context.Context,Event)error}
// Dispatcher retries unacknowledged messages on subsequent Drain calls.
type Dispatcher struct{ Store Store; Publisher Publisher; BatchSize int }
func(d Dispatcher)Drain(ctx context.Context)(int,error){
 if d.Store==nil||d.Publisher==nil{return 0,errors.New("durable: store and publisher required")}
 size:=d.BatchSize;if size<=0{size=100}
 batch,err:=d.Store.Pending(ctx,size);if err!=nil{return 0,err}
 done:=0
 for _,e:=range batch{
  if err:=ctx.Err();err!=nil{return done,err}
  if err:=d.Publisher.PublishEvent(ctx,e);err!=nil{return done,err}
  if err:=d.Store.Acknowledge(ctx,e.ID);err!=nil{return done,err}
  done++
 }
 return done,nil
}
