package module

import "context"

// Definition is a function-backed application module.
type Definition struct {
 ID string
 Requires []string
 OnStart func(context.Context)error
 OnStop func(context.Context)error
}
func(d Definition)Name()string{return d.ID}
func(d Definition)Dependencies()[]string{return append([]string(nil),d.Requires...)}
func(d Definition)Start(ctx context.Context)error{if d.OnStart!=nil{return d.OnStart(ctx)};return nil}
func(d Definition)Stop(ctx context.Context)error{if d.OnStop!=nil{return d.OnStop(ctx)};return nil}
