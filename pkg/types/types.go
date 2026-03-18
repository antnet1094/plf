// Package types defines all core data structures for PLF documents.
package types

// Known section names
const (
	SectionMeta     = "meta"
	SectionRole     = "role"
	SectionContext  = "context"
	SectionTools    = "tools"
	SectionRules    = "rules"
	SectionFallback = "fallback"
	SectionChain    = "chain"
	SectionTask     = "task"
	SectionOutput   = "output"
)

// RequiredSections lists sections that must be present in every PLF document.
var RequiredSections = []string{
	SectionRole,
	SectionContext,
	SectionRules,
	SectionFallback,
	SectionTask,
	SectionOutput,
}

// ToolParameter defines a single parameter for a tool.
type ToolParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // string, integer, boolean, array, object
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// ToolDefinition represents a function or capability the agent can use.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Webhook     string          `json:"webhook,omitempty"`
	Parameters  []ToolParameter `json:"parameters,omitempty"`
}

// Rule directive types
const (
	RuleNever  = "NEVER"
	RuleAlways = "ALWAYS"
	RuleIf     = "IF"
	RuleMax    = "MAX"
	RuleMin    = "MIN"
)

// Render target formats
const (
	FormatRaw   = "raw"
	FormatCore  = "core"
	FormatNexus = "nexus"
	FormatLocal = "local"
)

// MetaConfig holds document-level metadata.
type MetaConfig struct {
	Version     string
	Lang        string
	Description string
	Author      string
	Target      string // preferred render target: core, nexus, local, raw
}

// ContextEntry is a single verified knowledge boundary entry.
// Key may contain spaces (e.g. "PostgreSQL 16").
type ContextEntry struct {
	Key   string
	Value string
}

// Rule is a behavioral constraint directive.
type Rule struct {
	Type    string // NEVER | ALWAYS | IF | MAX | MIN
	Subject string // what the rule applies to
	Value   string // constraint value or condition
	Raw     string // original unparsed line (for display)
}

// FallbackConfig defines the uncertainty handling protocol.
// This replaces the fictional "threshold: 0.95" concept with
// observable behavioral triggers.
type FallbackConfig struct {
	// Signals are linguistic patterns that indicate the model is uncertain.
	// When the model detects these in its own forming response, it must stop.
	Signals []string

	// DefaultAction is the exact response to emit when uncertain.
	DefaultAction string

	// UnknownAction is the response when the user asks about something
	// outside the @context knowledge boundary.
	UnknownAction string

	// ConflictAction is the response when rules conflict with each other.
	ConflictAction string

	// Escalate names the human/system to escalate to.
	Escalate string
}

// ChainStep is a blocking reasoning checkpoint.
// If OnFail is set and the condition evaluates false, the step
// directs the model to invoke the fallback instead of continuing.
type ChainStep struct {
	Index    int
	Question string
	OnFail   string // "fallback" | "restrict" | "warn" | custom message
}

// OutputConfig defines strict response format constraints.
type OutputConfig struct {
	Format       string            // numbered_steps | json | markdown | plain | delegation
	MaxWords     int               // 0 = unlimited
	MaxItems     int               // for list-type outputs, 0 = unlimited
	Language     string            // response language (overrides @meta.lang)
	IncludeChain bool              // whether to show chain reasoning in output
	Fields       []string          // for json format: required fields
	Extra        map[string]string // format-specific extra keys
}

// Document is the fully parsed PLF document.
type Document struct {
	Meta         MetaConfig
	Role         string
	Context      []ContextEntry
	Tools        []ToolDefinition
	Rules        []Rule
	Fallback     FallbackConfig
	Chain        []ChainStep
	TaskTemplate string
	Output       OutputConfig

	// Custom holds any user-defined sections not in the standard set.
	Custom map[string][]string

	// SourceFile is the path to the .plf file, if parsed from disk.
	SourceFile string
}

// ValidationIssue describes a single problem found during validation.
type ValidationIssue struct {
	Section  string
	Message  string
	Severity string // "error" | "warning" | "info"
}

func (v ValidationIssue) String() string {
	return "[" + v.Severity + "] @" + v.Section + ": " + v.Message
}

// RenderResult is the structured output of rendering a Document.
type RenderResult struct {
	// System is the system/instruction portion (everything except @task).
	System string

	// User is the rendered @task with variables substituted.
	User string

	// Full is System + User concatenated (used for FormatRaw).
	Full string

	// Vars holds the variable values used during rendering.
	Vars map[string]string

	// UnresolvedVars lists template variables that had no value provided.
	UnresolvedVars []string
}

// CoreResponse matches the Core message format.
type CoreResponse struct {
	Messages []CoreMessage `json:"messages"`
}

type CoreMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// NexusResponse matches the Nexus API format.
type NexusResponse struct {
	System   string         `json:"system"`
	Messages []NexusMessage `json:"messages"`
}

type NexusMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LocalResponse is compatible with the local API chat endpoint.
type LocalResponse struct {
	Model    string         `json:"model,omitempty"`
	Messages []LocalMessage `json:"messages"`
	Stream   bool           `json:"stream"`
}

type LocalMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DynamicContextResolver is a callback used to fetch dynamic context natively during rendering.
// It receives the URI/command specified in the @context section (e.g., "MCP: file://logs/app", uri="file://logs/app").
type DynamicContextResolver func(uri string) (string, error)

// RenderOptions controls the rendering process.
type RenderOptions struct {
	// Vars maps template variable names to their values.
	// e.g. {"mensaje_usuario": "el servicio no inicia"}
	Vars map[string]string

	// Format selects the output structure.
	Format string

	// Resolver allows fetching dynamic context blocks (e.g., MCP calls) at render time.
	// If nil, dynamic context blocks will be skipped or rendered plainly.
	Resolver DynamicContextResolver

	// Minify compresses the prompt by removing ASCII art, blank lines, and verbose instructions.
	Minify bool
}
