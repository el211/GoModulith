// Package modulith coordinates independent modules in a modular monolith.
//
// A module declares its name, dependencies, and lifecycle through the Module
// interface. Modules are started in dependency order and stopped in reverse
// order. Import boundaries are checked separately with the architecture package.
package modulith

import (
 "context"
 "errors"
 "fmt"
 "sort"
 "sync"
 "time"
 "github.com/el211/GoModulith/observability"
)

// Module is an independently owned application capability.
type Module interface {
 Name() string
 Dependencies() []string
 Start(context.Context) error
 Stop(context.Context) error
}

// App manages the lifecycle and dependency graph of registered modules.
type App struct {
 name string
 mu sync.Mutex
 modules map[string]Module
 started []string
 running bool
 observer observability.Observer
}

// Option customizes an App.
type Option func(*App)

// WithName assigns a descriptive application name.
func WithName(name string) Option { return func(a *App) { a.name = name } }

// WithObserver instruments lifecycle transitions without requiring an SDK.
func WithObserver(observer observability.Observer) Option {return func(a *App){a.observer=observer}}

// New creates an empty application.
func New(options ...Option) *App {
 a := &App{name:"application",modules:make(map[string]Module)}
 for _, option := range options { option(a) }
 return a
}

// Name returns the application name.
func (a *App) Name() string { return a.name }

// Register adds modules before Run; duplicate and empty names are rejected.
func (a *App) Register(modules ...Module) error {
 a.mu.Lock(); defer a.mu.Unlock()
 if a.running { return errors.New("modulith: cannot register while running") }
 pending:=map[string]bool{}
 for _, m:=range modules {
  if m==nil { return errors.New("modulith: nil module") }
  n:=m.Name()
  if n=="" { return errors.New("modulith: empty module name") }
  if _,ok:=a.modules[n];ok || pending[n] { return fmt.Errorf("modulith: duplicate module %q",n) }
  pending[n]=true
 }
 for _,m:=range modules { a.modules[m.Name()]=m }
 return nil
}

// Order returns a deterministic dependency-first topological ordering.
func (a *App) Order() ([]string,error) {
 a.mu.Lock(); defer a.mu.Unlock()
 return order(a.modules)
}

func order(modules map[string]Module)([]string,error){
 state:=map[string]uint8{}
 names:=make([]string,0,len(modules))
 for n:=range modules { names=append(names,n) }
 sort.Strings(names)
 result:=make([]string,0,len(names))
 var visit func(string) error
 visit=func(n string) error {
  if state[n]==1 { return fmt.Errorf("modulith: dependency cycle at %q",n) }
  if state[n]==2 { return nil }
  m,ok:=modules[n]; if !ok { return fmt.Errorf("modulith: missing dependency %q",n) }
  state[n]=1
  deps:=append([]string(nil),m.Dependencies()...);sort.Strings(deps)
  for _,d:=range deps { if d==n { return fmt.Errorf("modulith: module %q depends on itself",n) };if err:=visit(d);err!=nil{return fmt.Errorf("%s: %w",n,err)} }
  state[n]=2;result=append(result,n);return nil
 }
 for _,n:=range names { if err:=visit(n);err!=nil{return nil,err} }
 return result,nil
}

// Verify checks dependency existence and cycles without starting modules.
func (a *App) Verify() error { _,err:=a.Order();return err }

// Modules returns the modules in dependency order.
func (a *App) Modules() ([]Module,error) {
 a.mu.Lock();defer a.mu.Unlock()
 names,err:=order(a.modules);if err!=nil{return nil,err}
 result:=make([]Module,0,len(names))
 for _,n:=range names { result=append(result,a.modules[n]) }
 return result,nil
}

// Run starts every module, rolling back successfully started modules on failure.
// Call Shutdown to stop the application. Run never blocks on context cancellation.
func (a *App) Run(ctx context.Context) error {
 a.mu.Lock();defer a.mu.Unlock()
 if a.running { return errors.New("modulith: already running") }
 names,err:=order(a.modules);if err!=nil{return err}
 started:=make([]string,0,len(names))
 for _,n:=range names {
  if err:=ctx.Err();err!=nil { a.rollback(ctx,started);return err }
  begin:=time.Now()
  err:=a.modules[n].Start(ctx)
  if a.observer!=nil {a.observer.ModuleStarted(ctx,n,time.Since(begin),err)}
  if err!=nil { a.rollback(context.WithoutCancel(ctx),started);return fmt.Errorf("modulith: start %s: %w",n,err) }
  started=append(started,n)
 }
 a.started=started;a.running=true;return nil
}
func(a *App)rollback(ctx context.Context,started []string){
 for i:=len(started)-1;i>=0;i-- { _=a.modules[started[i]].Stop(ctx) }
}

// Shutdown stops all running modules in reverse order and joins errors.
func (a *App) Shutdown(ctx context.Context)error{
 a.mu.Lock();defer a.mu.Unlock()
 if !a.running{return nil}
 var errs []error
 for i:=len(a.started)-1;i>=0;i-- { n:=a.started[i];begin:=time.Now();err:=a.modules[n].Stop(ctx);if a.observer!=nil{a.observer.ModuleStopped(ctx,n,time.Since(begin),err)};if err!=nil{errs=append(errs,fmt.Errorf("%s: %w",n,err))} }
 a.running=false;a.started=nil
 return errors.Join(errs...)
}
