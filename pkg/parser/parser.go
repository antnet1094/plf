// Package parser implements the PLF document parser.
//
// Grammar (informal):
//
//	document    := section*
//	section     := header body
//	header      := '@' name newline
//	body        := line*
//	line        := comment | keyvalue | numbered | listitem | textline
//	comment     := '#' rest-of-line
//	keyvalue    := identifier ':' rest-of-line   (identifier = no spaces)
//	numbered    := digit+ '.' rest-of-line
//	listitem    := '-' rest-of-line
//	textline    := non-empty non-comment rest-of-line
package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/antnet1094/plf/pkg/types"
)

// rawSection holds the unparsed lines belonging to a section.
type rawSection struct {
	name  string
	lines []string
}

// Parse reads a .plf file from disk and returns a Document.
func ParseFile(path string) (*types.Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer f.Close()

	doc, err := Parse(f)
	if err != nil {
		return nil, err
	}
	doc.SourceFile = path
	return doc, nil
}

// Parse reads PLF content from an io.Reader and returns a Document.
func Parse(r io.Reader) (*types.Document, error) {
	sections, err := lex(r)
	if err != nil {
		return nil, err
	}
	return assemble(sections)
}

// ParseString parses PLF content from a string.
func ParseString(content string) (*types.Document, error) {
	return Parse(strings.NewReader(content))
}

// ─── Lexer ────────────────────────────────────────────────────────────────────

// lex splits the input into raw sections by scanning for @header lines.
func lex(r io.Reader) ([]rawSection, error) {
	scanner := bufio.NewScanner(r)
	var sections []rawSection
	var current *rawSection

	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip top-level comments and blank lines before first section
		if current == nil {
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
		}

		// Section header: starts with '@'
		if strings.HasPrefix(trimmed, "@") {
			name := strings.TrimPrefix(trimmed, "@")
			// Remove inline comments
			if idx := strings.Index(name, "#"); idx != -1 {
				name = name[:idx]
			}
			name = strings.ToLower(strings.TrimSpace(name))
			if name == "" {
				return nil, fmt.Errorf("line %d: empty section name after '@'", lineNo)
			}
			sections = append(sections, rawSection{name: name})
			current = &sections[len(sections)-1]
			continue
		}

		// Body line: append to current section
		if current != nil {
			current.lines = append(current.lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}
	return sections, nil
}

// ─── Assembler ────────────────────────────────────────────────────────────────

// assemble maps raw sections into a typed Document.
func assemble(sections []rawSection) (*types.Document, error) {
	doc := &types.Document{
		Custom: make(map[string][]string),
		Meta: types.MetaConfig{
			Version: "1.0",
			Lang:    "es",
		},
	}

	for _, s := range sections {
		lines := cleanLines(s.lines)
		switch s.name {
		case types.SectionMeta:
			doc.Meta = parseMeta(lines, doc.Meta)
		case types.SectionRole:
			doc.Role = parseText(lines)
		case types.SectionContext:
			doc.Context = parseContext(lines)
		case types.SectionTools:
			doc.Tools = parseTools(lines)
		case types.SectionRules:
			doc.Rules = parseRules(lines)
		case types.SectionFallback:
			doc.Fallback = parseFallback(lines)
		case types.SectionChain:
			doc.Chain = parseChain(lines)
		case types.SectionTask:
			doc.TaskTemplate = parseText(lines)
		case types.SectionOutput:
			doc.Output = parseOutput(lines)
		default:
			doc.Custom[s.name] = lines
		}
	}

	return doc, nil
}

// ─── Section-specific parsers ─────────────────────────────────────────────────

// parseMeta parses key:value pairs, seeding defaults from base.
func parseMeta(lines []string, base types.MetaConfig) types.MetaConfig {
	kv := parseKeyValues(lines)
	if v, ok := kv["version"]; ok {
		base.Version = v
	}
	if v, ok := kv["lang"]; ok {
		base.Lang = v
	}
	if v, ok := kv["description"]; ok {
		base.Description = v
	}
	if v, ok := kv["author"]; ok {
		base.Author = v
	}
	if v, ok := kv["target"]; ok {
		base.Target = strings.ToLower(v)
	}
	return base
}

// parseContext handles lines like "key: value" where the key may contain spaces.
// e.g.  "PostgreSQL 16: puerto 5432, logs /var/log/postgresql/"
// We split on the FIRST colon only.
func parseContext(lines []string) []types.ContextEntry {
	var entries []types.ContextEntry
	for _, line := range lines {
		if isComment(line) || strings.TrimSpace(line) == "" {
			continue
		}
		idx := strings.Index(line, ":")
		if idx == -1 {
			// No colon — store as a value-only entry
			entries = append(entries, types.ContextEntry{Value: strings.TrimSpace(line)})
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		entries = append(entries, types.ContextEntry{Key: key, Value: val})
	}
	return entries
}

// parseTools handles lines like "tool_name: tool description"
func parseTools(lines []string) []types.ToolDefinition {
	var tools []types.ToolDefinition
	for _, line := range lines {
		if isComment(line) || strings.TrimSpace(line) == "" {
			continue
		}
		idx := strings.Index(line, ":")
		if idx == -1 {
			// No colon — just the name
			tools = append(tools, types.ToolDefinition{Name: strings.TrimSpace(line)})
			continue
		}
		name := strings.TrimSpace(line[:idx])
		desc := strings.TrimSpace(line[idx+1:])
		tools = append(tools, types.ToolDefinition{Name: name, Description: desc})
	}
	return tools
}

// parseRules recognises directive-style lines:
//
//	NEVER: <subject>
//	ALWAYS: <subject>
//	IF <condition>: <action>
//	MAX_<SUBJECT>: <n>
//	MIN_<SUBJECT>: <n>
func parseRules(lines []string) []types.Rule {
	var rules []types.Rule
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isComment(trimmed) || trimmed == "" {
			continue
		}

		var r types.Rule
		r.Raw = trimmed
		upper := strings.ToUpper(trimmed)

		switch {
		case strings.HasPrefix(upper, "NEVER:"):
			r.Type = types.RuleNever
			r.Value = strings.TrimSpace(trimmed[len("NEVER:"):])

		case strings.HasPrefix(upper, "ALWAYS:"):
			r.Type = types.RuleAlways
			r.Value = strings.TrimSpace(trimmed[len("ALWAYS:"):])

		case strings.HasPrefix(upper, "IF "):
			r.Type = types.RuleIf
			rest := trimmed[3:]
			if idx := strings.Index(rest, ":"); idx != -1 {
				r.Subject = strings.TrimSpace(rest[:idx])
				r.Value = strings.TrimSpace(rest[idx+1:])
			} else {
				r.Value = strings.TrimSpace(rest)
			}

		case strings.HasPrefix(upper, "MAX_") || strings.HasPrefix(upper, "MAX "):
			r.Type = types.RuleMax
			rest := trimmed[4:]
			if idx := strings.Index(rest, ":"); idx != -1 {
				r.Subject = strings.TrimSpace(rest[:idx])
				r.Value = strings.TrimSpace(rest[idx+1:])
			} else {
				r.Value = strings.TrimSpace(rest)
			}

		case strings.HasPrefix(upper, "MIN_") || strings.HasPrefix(upper, "MIN "):
			r.Type = types.RuleMin
			rest := trimmed[4:]
			if idx := strings.Index(rest, ":"); idx != -1 {
				r.Subject = strings.TrimSpace(rest[:idx])
				r.Value = strings.TrimSpace(rest[idx+1:])
			} else {
				r.Value = strings.TrimSpace(rest)
			}

		default:
			// Free-text rule — store as ALWAYS for semantic purposes
			r.Type = types.RuleAlways
			r.Value = trimmed
		}
		rules = append(rules, r)
	}
	return rules
}

// parseFallback parses the @fallback section.
// signals: comma-separated list of linguistic uncertainty markers.
// default: what to say when uncertain.
// unknown: what to say when context boundary is hit.
// conflict: what to say on rule conflicts.
// escalate: who to escalate to.
func parseFallback(lines []string) types.FallbackConfig {
	kv := parseKeyValues(lines)
	var fb types.FallbackConfig

	if v, ok := kv["signals"]; ok {
		for _, s := range strings.Split(v, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				fb.Signals = append(fb.Signals, s)
			}
		}
	}
	fb.DefaultAction = kv["default"]
	fb.UnknownAction = kv["unknown"]
	fb.ConflictAction = kv["conflict"]
	fb.Escalate = kv["escalate"]
	return fb
}

// parseChain parses numbered reasoning steps.
//
//	1. ¿El comando está en @context? → si no: fallback
//	2. ¿Cumple @rules? → si no: restrict
func parseChain(lines []string) []types.ChainStep {
	var steps []types.ChainStep
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isComment(trimmed) || trimmed == "" {
			continue
		}

		// Find leading number
		dotIdx := strings.Index(trimmed, ".")
		if dotIdx == -1 {
			continue
		}
		numStr := strings.TrimSpace(trimmed[:dotIdx])
		n, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}

		rest := strings.TrimSpace(trimmed[dotIdx+1:])
		step := types.ChainStep{Index: n}

		// Check for "→ si no: <action>" or "-> si no: <action>"
		for _, sep := range []string{"→", "->"} {
			if idx := strings.Index(rest, sep); idx != -1 {
				step.Question = strings.TrimSpace(rest[:idx])
				after := strings.TrimSpace(rest[idx+len(sep):])
				// Remove "si no:" or "if not:" prefix
				for _, prefix := range []string{"si no:", "if not:", "on fail:", "else:"} {
					if strings.HasPrefix(strings.ToLower(after), prefix) {
						after = strings.TrimSpace(after[len(prefix):])
						break
					}
				}
				step.OnFail = after
				break
			}
		}
		if step.Question == "" {
			step.Question = rest
		}

		steps = append(steps, step)
	}
	return steps
}

// parseOutput parses the @output section into an OutputConfig.
func parseOutput(lines []string) types.OutputConfig {
	kv := parseKeyValues(lines)
	out := types.OutputConfig{
		Extra: make(map[string]string),
	}

	if v, ok := kv["format"]; ok {
		out.Format = strings.ToLower(strings.TrimSpace(v))
	}
	if v, ok := kv["max_words"]; ok {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			out.MaxWords = n
		}
	}
	if v, ok := kv["max_items"]; ok {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			out.MaxItems = n
		}
	}
	if v, ok := kv["language"]; ok {
		out.Language = strings.ToLower(strings.TrimSpace(v))
	}
	if v, ok := kv["include_chain"]; ok {
		out.IncludeChain = isTruthy(v)
	}
	if v, ok := kv["fields"]; ok {
		for _, f := range strings.Split(v, ",") {
			f = strings.TrimSpace(f)
			if f != "" {
				out.Fields = append(out.Fields, f)
			}
		}
	}

	// Any remaining keys go into Extra
	known := map[string]bool{
		"format": true, "max_words": true, "max_items": true,
		"language": true, "include_chain": true, "fields": true,
	}
	for k, v := range kv {
		if !known[k] {
			out.Extra[k] = v
		}
	}
	return out
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// cleanLines removes blank lines and comment-only lines but preserves
// indentation of substantive lines. It does NOT trim individual lines
// so parsers can inspect indentation if needed.
func cleanLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		stripped := strings.TrimSpace(l)
		if stripped == "" || strings.HasPrefix(stripped, "#") {
			continue
		}
		out = append(out, stripped) // normalise to trimmed form
	}
	return out
}

// parseText joins lines into a single string preserving newlines.
func parseText(lines []string) string {
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// parseKeyValues extracts key:value pairs where the key is a single token
// (no spaces). Keys are lower-cased. Later duplicates overwrite earlier ones.
func parseKeyValues(lines []string) map[string]string {
	m := make(map[string]string)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isComment(trimmed) || trimmed == "" {
			continue
		}
		// Key must be contiguous non-space chars before the first colon
		idx := strings.Index(trimmed, ":")
		if idx == -1 {
			continue
		}
		rawKey := trimmed[:idx]
		// Only accept as KV if the key part has no spaces
		hasSpace := false
		for _, r := range rawKey {
			if unicode.IsSpace(r) {
				hasSpace = true
				break
			}
		}
		if hasSpace {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(rawKey))
		val := strings.TrimSpace(trimmed[idx+1:])
		// Strip inline comments from value
		if ci := strings.Index(val, " #"); ci != -1 {
			val = strings.TrimSpace(val[:ci])
		}
		if key != "" {
			m[key] = val
		}
	}
	return m
}

func isComment(s string) bool {
	return strings.HasPrefix(strings.TrimSpace(s), "#")
}

func isTruthy(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "yes" || s == "1" || s == "si" || s == "sí"
}

