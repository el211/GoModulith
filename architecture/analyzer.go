package architecture

import (
 "errors"
 "fmt"
 "go/parser"
 "go/token"
 "io/fs"
 "os"
 "path"
 "path/filepath"
 "sort"
 "strconv"
 "strings"
)

// Config identifies the Go module and application-module layout.
type Config struct {
 Root string // repository root containing go.mod
 ModulePath string // e.g. example.com/shop; inferred from go.mod if empty
 ModulesDir string // defaults to modules
 AllowedDependencies map[string][]string // optional allowlist, per source module
}
// Violation describes a forbidden cross-module import.
type Violation struct{ From,To,Import,Reason string }
func(v Violation)Error()string{return fmt.Sprintf("%s -> %s (%s): %s",v.From,v.To,v.Import,v.Reason)}
// Report contains discovered modules and architectural violations.
type Report struct{Modules []string;Dependencies map[string][]string;Violations []Violation}
func(r Report)Verify()error{
 errs:=make([]error,0,len(r.Violations))
 for _,v:=range r.Violations{errs=append(errs,v)}
 return errors.Join(errs...)
}
// Analyze inspects all .go files in each module (including tests), records
// imports between modules, rejects internal access and checks optional allowlists.
func Analyze(c Config)(Report,error){
 root:=c.Root;if root==""{root="."}
 dir:=c.ModulesDir;if dir==""{dir="modules"}
 if c.ModulePath==""{
  data,err:=os.ReadFile(filepath.Join(root,"go.mod"));if err!=nil{return Report{},err}
  for _,line:=range strings.Split(string(data),"\n"){s:=strings.Fields(strings.TrimSpace(line));if len(s)>=2&&s[0]=="module"{c.ModulePath=s[1];break}}
 }
 if c.ModulePath==""{return Report{},errors.New("architecture: missing module path")}
 entries,err:=os.ReadDir(filepath.Join(root,dir));if err!=nil{return Report{},err}
 r:=Report{Dependencies:map[string][]string{}}
 for _,e:=range entries{if e.IsDir()&&!strings.HasPrefix(e.Name(),".")&& !strings.HasPrefix(e.Name(),"_"){r.Modules=append(r.Modules,e.Name())}}
 sort.Strings(r.Modules)
 exists:=map[string]bool{};for _,n:=range r.Modules{exists[n]=true;r.Dependencies[n]=[]string{}}
 prefix:=path.Join(c.ModulePath,filepath.ToSlash(dir))+"/"
 edges:=map[string]map[string]bool{}
 for _,n:=range r.Modules{
  edges[n]=map[string]bool{}
  base:=filepath.Join(root,dir,n)
  err:=filepath.WalkDir(base,func(filename string,d fs.DirEntry,walkErr error)error{
   if walkErr!=nil{return walkErr}
   if d.IsDir(){if filename!=base&&(d.Name()=="vendor"||strings.HasPrefix(d.Name(),".")){return filepath.SkipDir};return nil}
   if !strings.HasSuffix(filename,".go"){return nil}
   file,err:=parser.ParseFile(token.NewFileSet(),filename,nil,parser.ImportsOnly);if err!=nil{return fmt.Errorf("%s: %w",filename,err)}
   for _,spec:=range file.Imports{
    imp,err:=strconv.Unquote(spec.Path.Value);if err!=nil{return err}
    if !strings.HasPrefix(imp,prefix){continue}
    parts:=strings.Split(strings.TrimPrefix(imp,prefix),"/");target:=parts[0]
    if !exists[target]||target==n{continue}
    edges[n][target]=true
    reason:=""
    if len(parts)>1&&parts[1]=="internal"{reason="cross-module internal package access"}
    if allowed,configured:=c.AllowedDependencies[n];configured{
     ok:=false;for _,a:=range allowed{if a==target{ok=true;break}}
     if !ok{reason="dependency is not allowed"}
    }
    if reason!=""{r.Violations=append(r.Violations,Violation{n,target,imp,reason})}
   }
   return nil
  });if err!=nil{return r,err}
 }
 for _,n:=range r.Modules{for dep:=range edges[n]{r.Dependencies[n]=append(r.Dependencies[n],dep)};sort.Strings(r.Dependencies[n])}
 states:=map[string]int{};var visit func(string,[]string)
 seen:=map[string]bool{}
 visit=func(n string,stack []string){
  states[n]=1
  for _,dep:=range r.Dependencies[n]{
   if states[dep]==1{
    key:=n+"->"+dep;if !seen[key]{seen[key]=true;r.Violations=append(r.Violations,Violation{n,dep,"","dependency cycle"})}
   } else if states[dep]==0 {visit(dep,append(stack,n))}
  }
  states[n]=2
 }
 for _,n:=range r.Modules{if states[n]==0{visit(n,nil)}}
 sort.Slice(r.Violations,func(i,j int)bool{a,b:=r.Violations[i],r.Violations[j];return a.From+a.To+a.Import+a.Reason<b.From+b.To+b.Import+b.Reason})
 return r,nil
}
