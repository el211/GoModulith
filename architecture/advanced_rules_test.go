package architecture_test
import (
 "strings"
 "testing"
 "github.com/el211/GoModulith/architecture"
)
func TestAdditionalRules(t *testing.T){
 report:=architecture.Report{Modules:[]string{"api","orders","payments"},Dependencies:map[string][]string{"api":{"orders","payments"},"orders":{},"payments":{}}}
 if err:=architecture.VerifyRules(report,architecture.MaxDependencies{Module:"api",Limit:1});err==nil||!strings.Contains(err.Error(),"maximum"){t.Fatalf("expected maximum violation, got %v",err)}
 if err:=architecture.VerifyRules(report,architecture.RequiredDependency{From:"orders",To:"payments"});err==nil||!strings.Contains(err.Error(),"required"){t.Fatalf("expected required dependency violation, got %v",err)}
 if err:=architecture.VerifyRules(report,architecture.RequiredDependency{From:"api",To:"orders"},architecture.MaxDependencies{Module:"api",Limit:2});err!=nil{t.Fatalf("expected valid rules, got %v",err)}
}
