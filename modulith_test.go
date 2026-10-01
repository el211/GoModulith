package modulith_test

import("context";"reflect";"testing";"github.com/el211/GoModulith";"github.com/el211/GoModulith/module")
func TestOrderAndShutdown(t *testing.T){
 var calls []string;mk:=func(id string,deps ...string)module.Definition{return module.Definition{ID:id,Requires:deps,OnStart:func(context.Context)error{calls=append(calls,"start "+id);return nil},OnStop:func(context.Context)error{calls=append(calls,"stop "+id);return nil}}}
 app:=modulith.New();if err:=app.Register(mk("payments","orders"),mk("orders","users"),mk("users"));err!=nil{t.Fatal(err)}
 if err:=app.Run(context.Background());err!=nil{t.Fatal(err)}
 if err:=app.Shutdown(context.Background());err!=nil{t.Fatal(err)}
 want:=[]string{"start users","start orders","start payments","stop payments","stop orders","stop users"}
 if !reflect.DeepEqual(calls,want){t.Fatalf("got %v want %v",calls,want)}
}
func TestCycleAndUnknown(t *testing.T){
 a:=modulith.New();_ =a.Register(module.Definition{ID:"a",Requires:[]string{"b"}},module.Definition{ID:"b",Requires:[]string{"a"}})
 if err:=a.Verify();err==nil{t.Fatal("expected cycle")}
 b:=modulith.New();_ =b.Register(module.Definition{ID:"a",Requires:[]string{"missing"}})
 if err:=b.Verify();err==nil{t.Fatal("expected missing dependency")}
}
