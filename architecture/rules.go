package architecture

import (
 "fmt"
 "strings"
)

// Rule defines a richer declarative constraint on a completed import graph.
type Rule interface { Check(Report) []Violation }
// NoDependency forbids a directed module-to-module edge.
type NoDependency struct{From,To string}
func(rule NoDependency)Check(r Report)[]Violation{
 for _,dep:=range r.Dependencies[rule.From]{if dep==rule.To{return []Violation{{From:rule.From,To:rule.To,Reason:"explicitly forbidden dependency"}}}}
 return nil
}
// LayerRule maps modules to ordered layers. Higher layers may import lower
// layers, but importing a higher layer or a peer (unless SameLayer is true)
// is forbidden. Unmapped modules are ignored.
type LayerRule struct{Layers [][]string;SameLayer bool}
func(rule LayerRule)Check(r Report)[]Violation{
 layer:=map[string]int{}
 for i,names:=range rule.Layers{for _,n:=range names{layer[n]=i}}
 var violations []Violation
 for from,deps:=range r.Dependencies{
  source,ok:=layer[from];if !ok{continue}
  for _,to:=range deps{
   target,ok:=layer[to];if !ok{continue}
   if target>source || (target==source&&!rule.SameLayer){violations=append(violations,Violation{From:from,To:to,Reason:"layer dependency violation"})}
  }
 }
 return violations
}
// NamingRule requires module names to begin with the given prefix.
type NamingRule struct{Prefix string}
func(rule NamingRule)Check(r Report)[]Violation{
 var violations []Violation
 for _,name:=range r.Modules{if !strings.HasPrefix(name,rule.Prefix){violations=append(violations,Violation{From:name,Reason:fmt.Sprintf("module must begin with %q",rule.Prefix)})}}
 return violations
}
// VerifyRules checks source-level restrictions alongside custom policies.
func VerifyRules(r Report,rules ...Rule)error{
 for _,rule:=range rules{r.Violations=append(r.Violations,rule.Check(r)...)}
 return r.Verify()
}
