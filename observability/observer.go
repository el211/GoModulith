package observability

import("context";"time")
// Observer receives module lifecycle and event-dispatch measurements.
type Observer interface{
 ModuleStarted(context.Context,string,time.Duration,error)
 ModuleStopped(context.Context,string,time.Duration,error)
 EventPublished(context.Context,string,time.Duration,error)
}
// NopObserver can be embedded to opt in selectively to instrumentation.
type NopObserver struct{}
func(NopObserver)ModuleStarted(context.Context,string,time.Duration,error){}
func(NopObserver)ModuleStopped(context.Context,string,time.Duration,error){}
func(NopObserver)EventPublished(context.Context,string,time.Duration,error){}
