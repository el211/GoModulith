package architecture

import (
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
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
 "sync"
)

// Cache remembers package import lists by file content digest. Repeated
// analyses only parse changed files; deleted files are pruned automatically.
// A Cache is safe for concurrent use by tooling.
type Cache struct {
 mu sync.Mutex
 files map[string]cachedImports
}
type cachedImports struct { Digest string; Imports []string }
func NewCache()*Cache{return &Cache{files:map[string]cachedImports{}}}

// Stats reports how much work was reused on one incremental scan.
type Stats struct{ParsedFiles,ReusedFiles int}

// AnalyzeIncremental performs per-source-file incremental parsing while
// re-evaluating all architecture rules against the resulting complete graph.
// Never reuse the previous Report when Config rules are modified.
func AnalyzeIncremental(c Config,cache *Cache)(Report,Stats,error){
 var stats Stats
 if cache==nil{cache=NewCache()}
 cache.mu.Lock();defer cache.mu.Unlock()
 if cache.files==nil{cache.files=map[string]cachedImports{}}
 root:=c.Root;if root==""{root="."}
 dir:=c.ModulesDir;if dir==""{dir="modules"}
 if c.ModulePath==""{
  data,err:=os.ReadFile(filepath.Join(root,"go.mod"));if err!=nil{return Report{},stats,err}
  for _,line:=range strings.Split(string(data),"\n"){fields:=strings.Fields(line);if len(fields)>1&&fields[0]=="module"{c.ModulePath=fields[1];break}}
 }
 if c.ModulePath==""{return Report{},stats,errors.New("architecture: missing module path")}
 entries,err:=os.ReadDir(filepath.Join(root,dir));if err!=nil{return Report{},stats,err}
 r:=Report{Dependencies:map[string][]string{}}
 exists:=map[string]bool{}
 for _,entry:=range entries{if entry.IsDir()&&!strings.HasPrefix(entry.Name(),".")&&!strings.HasPrefix(entry.Name(),"_"){r.Modules=append(r.Modules,entry.Name());exists[entry.Name()]=true;r.Dependencies[entry.Name()]=[]string{}}}
 sort.Strings(r.Modules)
 seen:=map[string]bool{};edges:=map[string]map[string]bool{}
 prefix:=path.Join(c.ModulePath,filepath.ToSlash(dir))+"/"
 for _,name:=range r.Modules{
  edges[name]=map[string]bool{}
  base:=filepath.Join(root,dir,name)
  err:=filepath.WalkDir(base,func(file string,d fs.DirEntry,walkErr error)error{
   if walkErr!=nil{return walkErr}
   if d.IsDir(){if file!=base&&(d.Name()=="vendor"||strings.HasPrefix(d.Name(),".")){return filepath.SkipDir};return nil}
   if !strings.HasSuffix(file,".go"){return nil}
   canonical,err:=filepath.Abs(file);if err!=nil{return err};seen[canonical]=true
   content,err:=os.ReadFile(file);if err!=nil{return err}
   digest:=sha256.Sum256(content);hash:=hex.EncodeToString(digest[:])
   item,ok:=cache.files[canonical]
   if !ok||item.Digest!=hash{
    parsed,err:=parser.ParseFile(token.NewFileSet(),file,content,parser.ImportsOnly);if err!=nil{return fmt.Errorf("%s: %w",file,err)}
    item=cachedImports{Digest:hash,Imports:[]string{}}
    for _,spec:=range parsed.Imports{v,err:=strconv.Unquote(spec.Path.Value);if err!=nil{return err};item.Imports=append(item.Imports,v)}
    cache.files[canonical]=item;stats.ParsedFiles++
   }else{stats.ReusedFiles++}
   for _,imp:=range item.Imports{
    if !strings.HasPrefix(imp,prefix){continue}
    parts:=strings.Split(strings.TrimPrefix(imp,prefix),"/");dest:=parts[0]
    if !exists[dest]||dest==name{continue}
    edges[name][dest]=true;reason:=""
    if len(parts)>1&&parts[1]=="internal"{reason="cross-module internal package access"}
    if allowed,configured:=c.AllowedDependencies[name];configured{
     permitted:=false;for _,dep:=range allowed{if dep==dest{permitted=true;break}}
     if !permitted{reason="dependency is not allowed"}
    }
    if reason!=""{r.Violations=append(r.Violations,Violation{name,dest,imp,reason})}
   }
   return nil
  });if err!=nil{return r,stats,err}
 }
 for key:=range cache.files{if !seen[key]{delete(cache.files,key)}}
 for _,name:=range r.Modules{for dep:=range edges[name]{r.Dependencies[name]=append(r.Dependencies[name],dep)};sort.Strings(r.Dependencies[name])}
 state:=map[string]uint8{};var visit func(string)
 visit=func(name string){state[name]=1;for _,dep:=range r.Dependencies[name]{if state[dep]==1{r.Violations=append(r.Violations,Violation{name,dep,"","dependency cycle"})}else if state[dep]==0{visit(dep)}};state[name]=2}
 for _,name:=range r.Modules{if state[name]==0{visit(name)}}
 sort.Slice(r.Violations,func(i,j int)bool{a,b:=r.Violations[i],r.Violations[j];return a.From+a.To+a.Import+a.Reason<b.From+b.To+b.Import+b.Reason})
 return r,stats,nil
}

// MarshalCache serializes the parser cache for use in incremental CI runs.
func(c *Cache)MarshalCache()([]byte,error){c.mu.Lock();defer c.mu.Unlock();return json.Marshal(c.files)}
// UnmarshalCache restores a previously serialized parser cache. Digests are
// rechecked against source files on the next scan before reuse.
func(c *Cache)UnmarshalCache(data []byte)error{
 c.mu.Lock();defer c.mu.Unlock()
 restored:=map[string]cachedImports{};if err:=json.Unmarshal(data,&restored);err!=nil{return err};c.files=restored;return nil
}
