package reliable

import (
 "context"
 "errors"
 "fmt"
 "math"
 "time"
 "github.com/el211/GoModulith/durable"
)

// Message adds delivery metadata to the persistent event envelope.
type Message struct {
 durable.Event
 Attempts int
 AvailableAt time.Time
 LeaseUntil time.Time
 LastError string
}
// Store atomically claims work for a worker and persists its outcomes.
// Claim must exclude active leases, and Ack/Retry/Dead must verify ownership.
type Store interface {
 Enqueue(context.Context, durable.Event) error
 Claim(context.Context,string,time.Time,time.Duration,int)([]Message,error)
 Ack(context.Context,string,string)error
 Retry(context.Context,string,string,int,time.Time,string)error
 Dead(context.Context,string,string,int,string)error
}
// Publisher sends the serialized event.
type Publisher interface { PublishEvent(context.Context,durable.Event)error }
// Policy configures attempts and delays. MaxAttempts includes the first try.
type Policy struct {MaxAttempts int; BaseDelay time.Duration; MaxDelay time.Duration; Lease time.Duration; Batch int}
func(p Policy)normalize()Policy{
 if p.MaxAttempts<=0{p.MaxAttempts=5};if p.BaseDelay<=0{p.BaseDelay=time.Second}
 if p.MaxDelay<=0{p.MaxDelay=time.Minute};if p.Lease<=0{p.Lease=30*time.Second};if p.Batch<=0{p.Batch=100}
 return p
}
// Delay returns capped exponential backoff; attempt 1 uses BaseDelay.
// No jitter is included; callers may layer jitter when scheduling jobs.
func(p Policy)Delay(attempt int)time.Duration{
 p=p.normalize();if attempt<=1{return p.BaseDelay}
 delay:=float64(p.BaseDelay)*math.Pow(2,float64(attempt-1))
 if delay>=float64(p.MaxDelay){return p.MaxDelay}
 return time.Duration(delay)
}
// Dispatcher owns one worker ID and processes a bounded claimed batch.
type Dispatcher struct{Store Store;Publisher Publisher;WorkerID string;Policy Policy;Now func()time.Time}
// Drain attempts every claimed message; errors are joined so a failed
// delivery does not prevent other claimed messages from being attempted.
func(d Dispatcher)Drain(ctx context.Context)(int,error){
 if d.Store==nil||d.Publisher==nil||d.WorkerID==""{return 0,errors.New("reliable: store, publisher and worker ID are required")}
 p:=d.Policy.normalize();now:=time.Now;if d.Now!=nil{now=d.Now}
 jobs,err:=d.Store.Claim(ctx,d.WorkerID,now().UTC(),p.Lease,p.Batch);if err!=nil{return 0,err}
 done:=0;var errs []error
 for _,job:=range jobs{
  if err:=ctx.Err();err!=nil{return done,errors.Join(append(errs,err)...)}
  attempt:=job.Attempts+1
  pubErr:=d.Publisher.PublishEvent(ctx,job.Event)
  if pubErr==nil{
   if err:=d.Store.Ack(ctx,job.ID,d.WorkerID);err!=nil{errs=append(errs,fmt.Errorf("ack %s: %w",job.ID,err))}else{done++}
   continue
  }
  var stateErr error
  if attempt>=p.MaxAttempts{stateErr=d.Store.Dead(ctx,job.ID,d.WorkerID,attempt,pubErr.Error())}else{
   stateErr=d.Store.Retry(ctx,job.ID,d.WorkerID,attempt,now().Add(p.Delay(attempt)),pubErr.Error())
  }
  errs=append(errs,fmt.Errorf("publish %s: %w",job.ID,pubErr))
  if stateErr!=nil{errs=append(errs,fmt.Errorf("record %s: %w",job.ID,stateErr))}
 }
 return done,errors.Join(errs...)
}

// DeadLetter represents an exhausted event retained for administrative review.
type DeadLetter struct { Event durable.Event; Attempts int; Reason string }

// DeadLetterStore is an optional administrative extension implemented by the
// SQL and MongoDB adapters. RequeueDead clears errors and attempts but does not
// guarantee exactly-once processing.
type DeadLetterStore interface {
 DeadLetters(context.Context,int)([]DeadLetter,error)
 RequeueDead(context.Context,string)error
}
