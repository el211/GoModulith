package memory

import ("context";"errors";"sync";"github.com/el211/GoModulith/durable")
// Store is a thread-safe, insertion-ordered volatile outbox.
type Store struct{mu sync.Mutex;order []string;records map[string]durable.Event}
func New()*Store{return &Store{records:make(map[string]durable.Event)}}
func(s *Store)Enqueue(ctx context.Context,e durable.Event)error{
 if err:=ctx.Err();err!=nil{return err};if e.ID==""{return errors.New("memory: empty event id")}
 s.mu.Lock();defer s.mu.Unlock();if _,ok:=s.records[e.ID];ok{return errors.New("memory: duplicate event id")}
 if s.records==nil{s.records=make(map[string]durable.Event)}
 e.Payload=append([]byte(nil),e.Payload...);s.records[e.ID]=e;s.order=append(s.order,e.ID);return nil
}
func(s *Store)Pending(ctx context.Context,max int)([]durable.Event,error){
 if err:=ctx.Err();err!=nil{return nil,err}
 s.mu.Lock();defer s.mu.Unlock();if max<=0{return nil,nil}
 out:=make([]durable.Event,0)
 for _,id:=range s.order {if len(out)>=max{break};e,ok:=s.records[id];if ok{e.Payload=append([]byte(nil),e.Payload...);out=append(out,e)}}
 return out,nil
}
func(s *Store)Acknowledge(ctx context.Context,id string)error{
 if err:=ctx.Err();err!=nil{return err};s.mu.Lock();defer s.mu.Unlock()
 delete(s.records,id);for i,x:=range s.order {if x==id{s.order=append(s.order[:i],s.order[i+1:]...);break}}
 return nil
}
var _ durable.Store=(*Store)(nil)
