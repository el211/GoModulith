package architecture_test
import("os";"path/filepath";"testing";"github.com/el211/GoModulith/architecture")
func TestIncrementalCacheInvalidation(t *testing.T){
 root:=t.TempDir();dir:=filepath.Join(root,"modules","orders")
 if err:=os.MkdirAll(dir,0755);err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(filepath.Join(root,"go.mod"),[]byte("module example.com/shop\n"),0644);err!=nil{t.Fatal(err)}
 file:=filepath.Join(dir,"orders.go")
 if err:=os.WriteFile(file,[]byte("package orders\n"),0644);err!=nil{t.Fatal(err)}
 cache:=architecture.NewCache();cfg:=architecture.Config{Root:root}
 _,first,err:=architecture.AnalyzeIncremental(cfg,cache);if err!=nil{t.Fatal(err)}
 _,second,err:=architecture.AnalyzeIncremental(cfg,cache);if err!=nil{t.Fatal(err)}
 if first.ParsedFiles!=1||second.ReusedFiles!=1{t.Fatalf("first=%+v second=%+v",first,second)}
 if err:=os.WriteFile(file,[]byte("package orders\n// changed\n"),0644);err!=nil{t.Fatal(err)}
 _,third,err:=architecture.AnalyzeIncremental(cfg,cache);if err!=nil{t.Fatal(err)}
 if third.ParsedFiles!=1{t.Fatalf("third=%+v",third)}
}
