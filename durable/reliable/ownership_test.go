package reliable_test

import (
 "context"
 "strings"
 "testing"
 "time"
 "github.com/el211/GoModulith/durable"
 "github.com/el211/GoModulith/durable/reliable"
)
type ownerStore struct { claims []string; acknowledged []string; jobs []reliable.Message }
func(s *ownerStore) Enqueue(context.Context,durable.Event)error{return nil}
func(s *ownerStore) Claim(_ context.Context,owner string,_ time.Time,_ time.Duration,_ int)([]reliable.Message,error){s.claims=append(s.claims,owner);return s.jobs,nil}
func(s *ownerStore) Ack(_ context.Context,_ string,owner string)error{s.acknowledged=append(s.acknowledged,owner);return nil}
func(s *ownerStore) Retry(context.Context,string,string,int,time.Time,string)error{return nil}
func(s *ownerStore) Dead(context.Context,string,string,int,string)error{return nil}
func TestDrainUsesDistinctLeaseOwners(t *testing.T){
 store:=&ownerStore{jobs:[]reliable.Message{{Event:durable.Event{ID:"event",Topic:"orders"}}}}
 d:=reliable.Dispatcher{Store:store,Publisher:publisher{},WorkerID:"worker"}
 for i:=0;i<2;i++{n,err:=d.Drain(context.Background());if err!=nil||n!=1{t.Fatalf("drain %d: n=%d err=%v",i,n,err)}}
 if len(store.claims)!=2||store.claims[0]==store.claims[1]{t.Fatalf("owners not distinct: %v",store.claims)}
 for i,owner:=range store.claims{if !strings.HasPrefix(owner,"worker:")||store.acknowledged[i]!=owner{t.Fatalf("claim and ack ownership mismatch: claims=%v ack=%v",store.claims,store.acknowledged)}}
}
