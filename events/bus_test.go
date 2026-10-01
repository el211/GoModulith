package events_test
import("context";"testing";"github.com/el211/GoModulith/events")
type placed struct{ID int}
func TestTypedUnsubscribe(t *testing.T){
 bus:=events.New();n:=0
 cancel,err:=events.SubscribeTyped(bus,func(_ context.Context,e placed)error{n+=e.ID;return nil});if err!=nil{t.Fatal(err)}
 if err:=bus.Publish(context.Background(),placed{2});err!=nil{t.Fatal(err)}
 cancel();if err:=bus.Publish(context.Background(),placed{2});err!=nil{t.Fatal(err)}
 if n!=2{t.Fatalf("got %d",n)}
}
