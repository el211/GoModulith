package module

import "context"

// Info describes a module's package, ownership, exported contracts and dependencies.
// ImportPath is the public Go package import path, not an internal implementation path.
type Info struct {
 Name string `json:"name"`
 ImportPath string `json:"import_path,omitempty"`
 Description string `json:"description,omitempty"`
 Version string `json:"version,omitempty"`
 Owners []string `json:"owners,omitempty"`
 DependsOn []string `json:"depends_on,omitempty"`
 Exports []string `json:"exports,omitempty"`
}

// Provider is an optional interface for modules that publish package metadata.
type Provider interface { PackageInfo() Info }

// Definition is a function-backed application module. ID and Requires are
// lifecycle identifiers; Metadata holds its package-level descriptive fields.
type Definition struct {
 ID string
 Requires []string
 Metadata Info
 OnStart func(context.Context)error
 OnStop func(context.Context)error
}
func(d Definition)Name()string{return d.ID}
func(d Definition)Dependencies()[]string{return append([]string(nil),d.Requires...)}
func(d Definition)PackageInfo()Info{
 info:=d.Metadata
 info.Name=d.ID
 info.DependsOn=append([]string(nil),d.Requires...)
 info.Owners=append([]string(nil),info.Owners...)
 info.Exports=append([]string(nil),info.Exports...)
 return info
}
func(d Definition)Start(ctx context.Context)error{if d.OnStart!=nil{return d.OnStart(ctx)};return nil}
func(d Definition)Stop(ctx context.Context)error{if d.OnStop!=nil{return d.OnStop(ctx)};return nil}
