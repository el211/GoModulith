package otel
import (
 "context"
 "time"
 "github.com/el211/GoModulith/observability"
 "go.opentelemetry.io/otel/attribute"
 "go.opentelemetry.io/otel/metric"
)
// Observer measures lifecycle and dispatch duration and counts errors.
type Observer struct{duration metric.Float64Histogram; failures metric.Int64Counter}
// New creates instruments from an application-provided Meter.
func New(meter metric.Meter)(*Observer,error){
 dur,err:=meter.Float64Histogram("gomodulith.operation.duration",metric.WithUnit("s"));if err!=nil{return nil,err}
 fail,err:=meter.Int64Counter("gomodulith.operation.failures");if err!=nil{return nil,err}
 return &Observer{duration:dur,failures:fail},nil
}
func(o *Observer)record(ctx context.Context,kind,name string,d time.Duration,err error){
 attrs:=metric.WithAttributes(attribute.String("operation",kind),attribute.String("module_or_event",name))
 o.duration.Record(ctx,d.Seconds(),attrs);if err!=nil{o.failures.Add(ctx,1,attrs)}
}
func(o *Observer)ModuleStarted(ctx context.Context,n string,d time.Duration,e error){o.record(ctx,"module.start",n,d,e)}
func(o *Observer)ModuleStopped(ctx context.Context,n string,d time.Duration,e error){o.record(ctx,"module.stop",n,d,e)}
func(o *Observer)EventPublished(ctx context.Context,n string,d time.Duration,e error){o.record(ctx,"event.publish",n,d,e)}
var _ observability.Observer=(*Observer)(nil)
