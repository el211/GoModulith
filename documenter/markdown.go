package documenter
import("fmt";"strings";"github.com/el211/GoModulith/architecture")
// Mermaid emits a module dependency graph with safely quoted node labels.
func Mermaid(r architecture.Report)string{
 var b strings.Builder;b.WriteString("graph TD\n")
 for i,n:=range r.Modules{fmt.Fprintf(&b,"  M%d[%q]\n",i,n)}
 for i,n:=range r.Modules{for _,d:=range r.Dependencies[n]{for j,m:=range r.Modules{if d==m{fmt.Fprintf(&b,"  M%d --> M%d\n",i,j)}}}}
 return b.String()
}
// Markdown generates a simple package inventory and Mermaid diagram.
func Markdown(r architecture.Report)string{
 var b strings.Builder;b.WriteString("# Module architecture\n\n")
 for _,n:=range r.Modules{fmt.Fprintf(&b,"- **%s**",n);if deps:=r.Dependencies[n];len(deps)>0{fmt.Fprintf(&b," depends on: %s",strings.Join(deps,", "))};b.WriteString("\n")}
 b.WriteString("\n```mermaid\n");b.WriteString(Mermaid(r));b.WriteString("```\n")
 return b.String()
}
