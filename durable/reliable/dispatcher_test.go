package reliable_test
import("context";"errors";"testing";"time";"github.com/el211/GoModulith/durable";"github.com/el211/GoModulith/durable/reliable")
type fake struct{jobs []reliable.Message;acks,retries,deads int}
func(f *fake)Enqueue(context.Context,durable.Event)error{return nil}
func(f *fake)Claim(context.Context,string,time.Time,time.Duration,int)([]reliable.Message,error){return f.jobs,nil}
func(f *fake)Ack(context.Context,string,string)error{f.acks++;return nil}
func(f *fake)Retry(context.Context,string,string,int,time.Time,string)error{f.retries++;return nil}
func(f *fake)Dead(context.Context,string,string,int,string)error{f.deads++;return nil}
type publisher struct{err error}
func(p publisher)PublishEvent(context.Context,durable.Event)error{return p.err}
func TestDeadLetterOnExhaustion(t *testing.T){
 store:=&fake{jobs:[]reliable.Message{{Event:durable.Event{ID:"1"},Attempts:2}}}
 d:=reliable.Dispatcher{Store:store,Publisher:publisher{err:errors.New("unavailable")},WorkerID:"worker",Policy:reliable.Policy{MaxAttempts:3}}
 _,err:=d.Drain(context.Background());if err==nil||store.deads!=1||store.retries!=0{t.Fatalf("err=%v dead=%d retry=%d",err,store.deads,store.retries)}
}
func TestBackoffCap(t *testing.T){
 p:=reliable.Policy{BaseDelay:time.Second,MaxDelay:5*time.Second}
 if p.Delay(1)!=time.Second||p.Delay(3)!=4*time.Second||p.Delay(10)!=5*time.Second{t.Fatal("unexpected backoff")}
}
