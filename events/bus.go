package events

import (
 "context"
 "errors"
 "reflect"
 "sync"
)

// Handler receives a published event.
type Handler func(context.Context,any)error
type subscriber struct { id uint64; fn Handler }
// Bus dispatches events by their concrete Go type.
type Bus struct { mu sync.RWMutex; next uint64; handlers map[reflect.Type][]subscriber }
func New()*Bus{return &Bus{handlers:make(map[reflect.Type][]subscriber)}}

// Subscribe registers a handler for the concrete type represented by prototype.
// Its returned function safely removes only this registration.
func(b *Bus)Subscribe(prototype any,h Handler)(func(),error){
 if prototype==nil||h==nil{return nil,errors.New("events: prototype and handler required")}
 t:=reflect.TypeOf(prototype)
 b.mu.Lock();if b.handlers==nil{b.handlers=make(map[reflect.Type][]subscriber)}
 b.next++;id:=b.next;b.handlers[t]=append(b.handlers[t],subscriber{id,h});b.mu.Unlock()
 return func(){b.mu.Lock();defer b.mu.Unlock();old:=b.handlers[t];out:=make([]subscriber,0,len(old));for _,s:=range old{if s.id!=id{out=append(out,s)}};if len(out)==0{delete(b.handlers,t)}else{b.handlers[t]=out}},nil
}

// Publish synchronously dispatches all handlers and joins their errors.
func(b *Bus)Publish(ctx context.Context,event any)error{
 if event==nil{return errors.New("events: nil event")}
 b.mu.RLock();listeners:=append([]subscriber(nil),b.handlers[reflect.TypeOf(event)]...);b.mu.RUnlock()
 var errs []error
 for _,s:=range listeners{if err:=ctx.Err();err!=nil{return errors.Join(append(errs,err)...)};if err:=s.fn(ctx,event);err!=nil{errs=append(errs,err)}}
 return errors.Join(errs...)
}

// PublishAsync delivers asynchronously; the buffered result channel receives
// exactly one result and closes. The caller controls cancellation via ctx.
func(b *Bus)PublishAsync(ctx context.Context,event any)<-chan error{
 result:=make(chan error,1)
 go func(){defer close(result);result<-b.Publish(ctx,event)}()
 return result
}

// SubscribeTyped registers a statically typed handler without handler casts.
func SubscribeTyped[T any](b *Bus,h func(context.Context,T)error)(func(),error){
 if h==nil{return nil,errors.New("events: nil handler")}
 var t T
 // A nonnil pointer prototype is required when T is a pointer.
 typ:=reflect.TypeOf((*T)(nil)).Elem()
 b.mu.Lock();if b.handlers==nil{b.handlers=make(map[reflect.Type][]subscriber)}
 b.next++;id:=b.next
 b.handlers[typ]=append(b.handlers[typ],subscriber{id,func(ctx context.Context,v any)error{return h(ctx,v.(T))}})
 b.mu.Unlock()
 _=t
 return func(){b.mu.Lock();defer b.mu.Unlock();old:=b.handlers[typ];out:=make([]subscriber,0,len(old));for _,s:=range old{if s.id!=id{out=append(out,s)}};if len(out)==0{delete(b.handlers,typ)}else{b.handlers[typ]=out}},nil
}
