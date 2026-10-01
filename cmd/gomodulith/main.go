package main

import(
 "encoding/json"
 "flag"
 "fmt"
 "os"
 "github.com/el211/GoModulith/architecture"
 "github.com/el211/GoModulith/documenter"
)
func main(){os.Exit(run(os.Args[1:]))}
func run(args []string)int{
 if len(args)==0{fmt.Fprintln(os.Stderr,"usage: gomodulith <verify|graph|inspect|docs> [-root .] [-dir modules]");return 2}
 cmd:=args[0];flags:=flag.NewFlagSet(cmd,flag.ContinueOnError)
 root:=flags.String("root",".","repository root");dir:=flags.String("dir","modules","directory of modules")
 if err:=flags.Parse(args[1:]);err!=nil{return 2}
 report,err:=architecture.Analyze(architecture.Config{Root:*root,ModulesDir:*dir})
 if err!=nil{fmt.Fprintln(os.Stderr,err);return 1}
 switch cmd{
 case "verify":
  if err:=report.Verify();err!=nil{fmt.Fprintln(os.Stderr,err);return 1};fmt.Printf("verified %d modules\n",len(report.Modules))
 case "graph":fmt.Print(documenter.Mermaid(report))
 case "docs":fmt.Print(documenter.Markdown(report))
 case "inspect":if err:=json.NewEncoder(os.Stdout).Encode(report);err!=nil{fmt.Fprintln(os.Stderr,err);return 1}
 default:fmt.Fprintln(os.Stderr,"unknown command:",cmd);return 2
 }
 return 0
}
