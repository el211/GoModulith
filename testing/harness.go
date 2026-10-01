package testing

import("context";"github.com/el211/GoModulith";"github.com/el211/GoModulith/module")
// Harness starts one module and optional dependency stubs in topological order.
type Harness struct{App *modulith.App}
func New(target modulith.Module,dependencies ...modulith.Module)(*Harness,error){
 app:=modulith.New()
 all:=append(append([]modulith.Module{},dependencies...),target)
 if err:=app.Register(all...);err!=nil{return nil,err}
 if err:=app.Verify();err!=nil{return nil,err}
 return &Harness{App:app},nil
}
// Stub defines a dependency that does not require external resources.
func Stub(name string)modulith.Module{return module.Definition{ID:name}}
func(h *Harness)Start(ctx context.Context)error{return h.App.Run(ctx)}
func(h *Harness)Stop(ctx context.Context)error{return h.App.Shutdown(ctx)}
