package architecture_test
import("os";"path/filepath";"strings";"testing";"github.com/el211/GoModulith/architecture")
func TestInternalBoundary(t *testing.T){
 root:=t.TempDir()
 write:=func(name,body string){p:=filepath.Join(root,name);if err:=os.MkdirAll(filepath.Dir(p),0755);err!=nil{t.Fatal(err)};if err:=os.WriteFile(p,[]byte(body),0644);err!=nil{t.Fatal(err)}}
 write("go.mod","module example.com/shop\n\ngo 1.23\n")
 write("modules/orders/orders.go","package orders\nimport _ \"example.com/shop/modules/users/internal/store\"\n")
 write("modules/users/users.go","package users\n")
 report,err:=architecture.Analyze(architecture.Config{Root:root});if err!=nil{t.Fatal(err)}
 if len(report.Violations)!=1||!strings.Contains(report.Violations[0].Reason,"internal"){t.Fatalf("unexpected violations: %+v",report.Violations)}
}
