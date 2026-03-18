// Package validator performs semantic validation on a parsed PLF Document.
package validator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/antnet1094/plf/pkg/types"
)

var templateVarRe = regexp.MustCompile(`\{\{\s*(\w+)\s*\}\}`)

// Validate returns all issues found in the document.
func Validate(doc *types.Document) []types.ValidationIssue {
	var issues []types.ValidationIssue
	issues = append(issues, checkRequired(doc)...)
	issues = append(issues, checkRole(doc)...)
	issues = append(issues, checkContext(doc)...)
	issues = append(issues, checkTools(doc)...)
	issues = append(issues, checkRules(doc)...)
	issues = append(issues, checkFallback(doc)...)
	issues = append(issues, checkChain(doc)...)
	issues = append(issues, checkOutput(doc)...)
	issues = append(issues, checkCrossSection(doc)...)
	return issues
}

// HasErrors returns true if any issue is severity=error.
func HasErrors(issues []types.ValidationIssue) bool {
	for _, i := range issues {
		if i.Severity == "error" {
			return true
		}
	}
	return false
}

func checkRequired(doc *types.Document) []types.ValidationIssue {
	var issues []types.ValidationIssue
	if doc.Role == "" {
		issues = append(issues, err("role", "@role is required and cannot be empty"))
	}
	if len(doc.Context) == 0 {
		issues = append(issues, err("context", "@context is required — define at least one knowledge boundary entry"))
	}
	if doc.TaskTemplate == "" {
		issues = append(issues, err("task", "@task is required — cannot render a prompt without a task"))
	}
	return issues
}

func checkRole(doc *types.Document) []types.ValidationIssue {
	var issues []types.ValidationIssue
	if len(doc.Role) > 2000 {
		issues = append(issues, warn("role", fmt.Sprintf("@role is %d chars — consider condensing to <500", len(doc.Role))))
	}
	return issues
}

func checkContext(doc *types.Document) []types.ValidationIssue {
	var issues []types.ValidationIssue
	if len(doc.Context) > 50 {
		issues = append(issues, warn("context",
			fmt.Sprintf("@context has %d entries — large contexts increase token usage and reduce focus", len(doc.Context))))
	}
	return issues
}

func checkTools(doc *types.Document) []types.ValidationIssue {
	var issues []types.ValidationIssue
	names := make(map[string]bool)
	for _, t := range doc.Tools {
		if t.Name == "" {
			issues = append(issues, err("tools", "tool name cannot be empty"))
		}
		if names[t.Name] {
			issues = append(issues, err("tools", fmt.Sprintf("duplicate tool name: %s", t.Name)))
		}
		names[t.Name] = true
	}
	return issues
}

func checkRules(doc *types.Document) []types.ValidationIssue {
	var issues []types.ValidationIssue
	
	nevers := make(map[string]bool)
	always := make(map[string]bool)

	for _, r := range doc.Rules {
		if r.Type == types.RuleNever {
			nevers[strings.ToLower(r.Value)] = true
		}
		if r.Type == types.RuleAlways {
			always[strings.ToLower(r.Value)] = true
		}
		
		if r.Type == types.RuleMax || r.Type == types.RuleMin {
			if _, e := strconv.Atoi(strings.TrimSpace(r.Value)); e != nil {
				issues = append(issues, err("rules", fmt.Sprintf("%s_%s must have a numeric value", r.Type, r.Subject)))
			}
		}
	}

	// Check for direct contradictions
	for val := range nevers {
		if always[val] {
			issues = append(issues, err("rules", fmt.Sprintf("contradiction: rule '%s' is marked both ALWAYS and NEVER", val)))
		}
	}

	return issues
}

func checkFallback(doc *types.Document) []types.ValidationIssue {
	var issues []types.ValidationIssue
	fb := doc.Fallback
	if len(fb.Signals) == 0 {
		issues = append(issues, warn("fallback", "no uncertainty signals defined"))
	}
	if fb.DefaultAction == "" {
		issues = append(issues, warn("fallback", "no default fallback action defined"))
	}
	return issues
}

func checkChain(doc *types.Document) []types.ValidationIssue {
	var issues []types.ValidationIssue
	for i, step := range doc.Chain {
		if step.Index != i+1 {
			issues = append(issues, warn("chain", fmt.Sprintf("step %d index out of order, expected %d", step.Index, i+1)))
		}
		if step.Question == "" {
			issues = append(issues, err("chain", fmt.Sprintf("step %d question is empty", step.Index)))
		}
	}
	return issues
}

func checkOutput(doc *types.Document) []types.ValidationIssue {
	var issues []types.ValidationIssue
	if doc.Output.Format == "" {
		issues = append(issues, warn("output", "no output format defined"))
	} else {
		validFormats := map[string]bool{
			"numbered_steps": true,
			"json":           true,
			"markdown":       true,
			"plain":          true,
			"delegation":     true,
		}
		if !validFormats[doc.Output.Format] {
			issues = append(issues, err("output", fmt.Sprintf("invalid format: %s", doc.Output.Format)))
		}
	}
	
	validTypes := map[string]bool{
		"string": true, "integer": true, "number": true, "boolean": true, "array": true, "object": true,
	}
	typeRe := regexp.MustCompile(`\(\s*(\w+)\s*\)`)
	for _, f := range doc.Output.Fields {
		if m := typeRe.FindStringSubmatch(f); m != nil {
			typ := strings.ToLower(m[1])
			if !validTypes[typ] {
				issues = append(issues, err("output", fmt.Sprintf("field '%s' uses invalid JSON schema type '%s'", f, m[1])))
			}
		}
	}

	return issues
}

func checkCrossSection(doc *types.Document) []types.ValidationIssue {
	var issues []types.ValidationIssue
	
	// Check if @task variables are in @context or elsewhere (just a warning)
	taskVars := templateVarRe.FindAllStringSubmatch(doc.TaskTemplate, -1)
	if len(taskVars) == 0 {
		issues = append(issues, info("task", "@task has no template variables"))
	}

	// Check if @chain references @fallback but fallback is empty
	if doc.Fallback.DefaultAction == "" {
		for _, step := range doc.Chain {
			if strings.Contains(strings.ToLower(step.OnFail), "fallback") {
				issues = append(issues, err("chain", fmt.Sprintf("step %d references fallback but no default action is defined in @fallback", step.Index)))
			}
		}
	}

	// Cross-reference tools used in rules
	validTools := make(map[string]bool)
	for _, t := range doc.Tools {
		validTools[strings.ToLower(t.Name)] = true
	}
	
	// A heuristic: if a rule mentions a tool inside backticks (e.g. `alert_ops`) next to the word "tool" or "herramienta"
	toolRefRe := regexp.MustCompile("`([^`]+)`")
	for _, r := range doc.Rules {
		lowerRule := strings.ToLower(r.Value)
		if strings.Contains(lowerRule, "tool") || strings.Contains(lowerRule, "herramienta") {
			matches := toolRefRe.FindAllStringSubmatch(r.Value, -1)
			for _, m := range matches {
				toolName := strings.ToLower(m[1])
				if !validTools[toolName] && len(toolName) > 2 {
					issues = append(issues, warn("rules", fmt.Sprintf("rule references potential tool `%s` but it is not defined in @tools", m[1])))
				}
			}
		}
	}

	return issues
}

func err(section, msg string) types.ValidationIssue {
	return types.ValidationIssue{Section: section, Message: msg, Severity: "error"}
}

func warn(section, msg string) types.ValidationIssue {
	return types.ValidationIssue{Section: section, Message: msg, Severity: "warning"}
}

func info(section, msg string) types.ValidationIssue {
	return types.ValidationIssue{Section: section, Message: msg, Severity: "info"}
}

