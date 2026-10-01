// Command shop demonstrates a small modular application and a typed event.
package main

import (
 "context"
 "fmt"
 "log"
 "github.com/el211/GoModulith"
 "github.com/el211/GoModulith/events"
 "github.com/el211/GoModulith/module"
)
type OrderPlaced struct{OrderID string}
func main(){
 ctx:=context.Background();app:=modulith.New(modulith.WithName("shop"));bus:=events.New()
 _,err:=events.SubscribeTyped(bus,func(_ context.Context,e OrderPlaced)error{fmt.Println("payment requested for",e.OrderID);return nil});if err!=nil{log.Fatal(err)}
 if err:=app.Register(module.Definition{ID:"users"},module.Definition{ID:"orders",Requires:[]string{"users"}},module.Definition{ID:"payments",Requires:[]string{"orders"}});err!=nil{log.Fatal(err)}
 if err:=app.Verify();err!=nil{log.Fatal(err)}
 if err:=app.Run(ctx);err!=nil{log.Fatal(err)}
 defer func(){if err:=app.Shutdown(ctx);err!=nil{log.Print(err)}}()
 if err:=bus.Publish(ctx,OrderPlaced{OrderID:"123"});err!=nil{log.Fatal(err)}
}
