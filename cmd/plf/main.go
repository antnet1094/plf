// plf — CLI for the Prompt Language Format
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/antnet1094/plf/pkg/parser"
	"github.com/antnet1094/plf/pkg/renderer"
	"github.com/antnet1094/plf/pkg/types"
	"github.com/antnet1094/plf/pkg/validator"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	gray   = "\033[37m"
)

func col(code, s string) string { return code + s + reset }

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	var err error
	switch os.Args[1] {
	case "validate":
		err = runValidate(os.Args[2:])
	case "render":
		err = runRender(os.Args[2:])
	case "inspect":
		err = runInspect(os.Args[2:])
	case "lint":
		err = runLint(os.Args[2:])
	case "version":
		fmt.Println("plf version 1.0.0")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, col(red, "Error: ")+err.Error())
		os.Exit(1)
	}
}

func runValidate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: plf validate <file.plf>")
	}
	doc, err := parser.ParseFile(args[0])
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}
	issues := validator.Validate(doc)
	if len(issues) == 0 {
		fmt.Printf("%s %s — no issues found\n", col(green+bold, "✓"), args[0])
		return nil
	}
	for _, issue := range issues {
		switch issue.Severity {
		case "error":
			fmt.Printf("%s @%-12s %s\n", col(red+bold, "  ERROR  "), issue.Section, issue.Message)
		case "warning":
			fmt.Printf("%s @%-12s %s\n", col(yellow, "  WARN   "), issue.Section, issue.Message)
		case "info":
			fmt.Printf("%s @%-12s %s\n", col(gray, "  INFO   "), issue.Section, issue.Message)
		}
	}
	if validator.HasErrors(issues) {
		return fmt.Errorf("\n%s has validation errors", args[0])
	}
	fmt.Printf("\n%s %s has warnings but is renderable\n", col(yellow, "⚠"), args[0])
	return nil
}

type varFlags []string
func (v *varFlags) String() string        { return strings.Join(*v, ", ") }
func (v *varFlags) Set(s string) error    { *v = append(*v, s); return nil }

func runRender(args []string) error {
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	var vars varFlags
	fs.Var(&vars, "var", "Template variable key=value (repeatable)")
	format := fs.String("format", "", "Output format: raw|core|nexus|local")
	outputFile := fs.String("output", "", "Write output to file")
	jsonOut := fs.Bool("json", false, "Output as JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return fmt.Errorf("usage: plf render <file.plf> [-var key=value] [-format FORMAT] [-json] [-output file]")
	}
	doc, err := parser.ParseFile(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}
	issues := validator.Validate(doc)
	if validator.HasErrors(issues) {
		return fmt.Errorf("cannot render: validation failed — run 'plf validate' for details")
	}
	result, err := renderer.Render(doc, types.RenderOptions{
		Vars:   parseVars(vars),
		Format: *format,
	})
	if err != nil {
		return err
	}
	if len(result.UnresolvedVars) > 0 {
		fmt.Fprintf(os.Stderr, "%s Unresolved variables: %s\n",
			col(yellow, "⚠"), strings.Join(result.UnresolvedVars, ", "))
	}
	var out []byte
	if *jsonOut {
		switch strings.ToLower(*format) {
		case types.FormatCore:
			out, _ = json.MarshalIndent(renderer.ToCore(result), "", "  ")
		case types.FormatNexus:
			out, _ = json.MarshalIndent(renderer.ToNexus(result), "", "  ")
		default:
			out, _ = json.MarshalIndent(map[string]string{"system": result.System, "user": result.User}, "", "  ")
		}
	} else {
		out = []byte(result.Full)
	}
	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, out, 0644); err != nil {
			return err
		}
		fmt.Printf("%s Written to %s\n", col(green+bold, "✓"), *outputFile)
	} else {
		fmt.Println(string(out))
	}
	return nil
}

func runInspect(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: plf inspect <file.plf>")
	}
	doc, err := parser.ParseFile(args[0])
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}
	fmt.Printf("\n%s %s\n\n", col(bold, "📄"), args[0])
	sec("META")
	kv("version", doc.Meta.Version)
	kv("lang", doc.Meta.Lang)
	if doc.Meta.Description != "" { kv("description", doc.Meta.Description) }
	if doc.Meta.Author != ""      { kv("author", doc.Meta.Author) }
	if doc.Meta.Target != ""      { kv("target", doc.Meta.Target) }
	sec("ROLE")
	fmt.Printf("  %s\n", trunc(doc.Role, 120))
	sec("CONTEXT")
	for _, e := range doc.Context {
		if e.Key != "" {
			fmt.Printf("  %s %s\n", col(cyan, e.Key+":"), e.Value)
		} else {
			fmt.Printf("  %s\n", e.Value)
		}
	}
	sec("RULES")
	for _, r := range doc.Rules {
		var t string
		switch r.Type {
		case types.RuleNever:  t = col(red+bold, "NEVER  ")
		case types.RuleAlways: t = col(green+bold, "ALWAYS ")
		case types.RuleIf:     t = col(yellow, "IF     ")
		case types.RuleMax, types.RuleMin: t = col(cyan, r.Type+"    ")
		default: t = "       "
		}
		if r.Subject != "" {
			fmt.Printf("  %s %-20s %s\n", t, r.Subject, r.Value)
		} else {
			fmt.Printf("  %s %s\n", t, r.Value)
		}
	}
	sec("FALLBACK")
	if len(doc.Fallback.Signals) > 0   { kv("signals",  strings.Join(doc.Fallback.Signals, ", ")) }
	if doc.Fallback.DefaultAction != "" { kv("default",  doc.Fallback.DefaultAction) }
	if doc.Fallback.UnknownAction != ""  { kv("unknown",  doc.Fallback.UnknownAction) }
	if doc.Fallback.Escalate != ""       { kv("escalate", doc.Fallback.Escalate) }
	sec("CHAIN")
	if len(doc.Chain) == 0 {
		fmt.Printf("  %s\n", col(gray, "(none)"))
	}
	for _, step := range doc.Chain {
		if step.OnFail != "" {
			fmt.Printf("  %d. %s\n     %s on fail: %s\n", step.Index, step.Question, col(yellow, "→"), step.OnFail)
		} else {
			fmt.Printf("  %d. %s\n", step.Index, step.Question)
		}
	}
	sec("TASK TEMPLATE")
	fmt.Printf("  %s\n", trunc(doc.TaskTemplate, 200))
	sec("OUTPUT")
	if doc.Output.Format != ""   { kv("format", doc.Output.Format) }
	if doc.Output.MaxWords > 0   { kv("max_words", fmt.Sprintf("%d", doc.Output.MaxWords)) }
	if doc.Output.Language != "" { kv("language", doc.Output.Language) }
	kv("include_chain", fmt.Sprintf("%v", doc.Output.IncludeChain))
	if len(doc.Custom) > 0 {
		sec("CUSTOM SECTIONS")
		for name, lines := range doc.Custom {
			fmt.Printf("  @%s (%d lines)\n", name, len(lines))
		}
	}
	fmt.Println()
	return nil
}

func runLint(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: plf lint <file.plf>")
	}
	doc, err := parser.ParseFile(args[0])
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}
	issues := validator.Validate(doc)
	errs, warns, infos := 0, 0, 0
	for _, i := range issues {
		switch i.Severity {
		case "error":   errs++
		case "warning": warns++
		case "info":    infos++
		}
	}
	fmt.Printf("\n%s PLF Lint: %s\n\n", col(bold, "🔍"), args[0])
	if len(issues) == 0 {
		fmt.Printf("  %s All checks passed — document is optimal\n", col(green+bold, "✓"))
	} else {
		for _, issue := range issues {
			s := issue.Section
			if s == "" { s = "global" }
			switch issue.Severity {
			case "error":
				fmt.Printf("  %s @%-12s %s\n", col(red+bold, "✗ ERROR  "), s, issue.Message)
			case "warning":
				fmt.Printf("  %s @%-12s %s\n", col(yellow, "⚠ WARN   "), s, issue.Message)
			case "info":
				fmt.Printf("  %s @%-12s %s\n", col(gray, "ℹ INFO   "), s, issue.Message)
			}
		}
	}
	fmt.Printf("\n  %s  %s  %s\n\n",
		col(red, fmt.Sprintf("%d errors", errs)),
		col(yellow, fmt.Sprintf("%d warnings", warns)),
		col(gray, fmt.Sprintf("%d suggestions", infos)),
	)
	if errs > 0 {
		return fmt.Errorf("lint failed with %d errors", errs)
	}
	return nil
}

func parseVars(pairs []string) map[string]string {
	m := make(map[string]string)
	for _, p := range pairs {
		if idx := strings.Index(p, "="); idx != -1 {
			m[p[:idx]] = p[idx+1:]
		} else {
			m[p] = ""
		}
	}
	return m
}

func sec(name string)    { fmt.Printf("\n  %s\n", col(bold, "── "+name)) }
func kv(k, v string)     { fmt.Printf("  %-16s %s\n", col(cyan, k+":"), v) }
func trunc(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ↵ ")
	if len(s) > n { return s[:n] + "…" }
	return s
}

func printUsage() {
	fmt.Print(`
PLF — Prompt Language Format toolkit

USAGE:
  plf <command> [options]

COMMANDS:
  validate <file.plf>           Validate a PLF file
  render   <file.plf> [flags]   Render to a prompt string
  inspect  <file.plf>           Show parsed structure
  lint     <file.plf>           Extended linting with suggestions
  version                       Print version

RENDER FLAGS:
  -var key=value    Template variable (repeatable)
  -format string    Output format: raw | core | nexus | local
  -json             Output as JSON
  -output string    Write to file instead of stdout

EXAMPLES:
  plf validate examples/sysadmin.plf
  plf render   examples/sysadmin.plf -var mensaje_usuario="El servicio no inicia"
  plf render   examples/sysadmin.plf -var mensaje_usuario="502" -format nexus -json -output p.json
  plf inspect  examples/whatsapp_router.plf
  plf lint     examples/restaurant_bot.plf

`)
}

