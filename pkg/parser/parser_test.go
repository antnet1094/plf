package parser_test

import (
	"testing"

	"github.com/antnet1094/plf/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseFile_ValidPLF tests parsing a valid PLF file from disk
func TestParseFile_ValidPLF(t *testing.T) {
	t.Skip("Skipping file test - run manually with correct path")
	// doc, err := parser.ParseFile("../../../examples/restaurant_bot.plf")
	// require.NoError(t, err)
	// require.NotNil(t, doc)
	// assert.Equal(t, "1.0", doc.Meta.Version)
}

// TestParseString_ValidPLF tests parsing PLF from string
func TestParseString_ValidPLF(t *testing.T) {
	content := `
@meta
  version: 1.0
  lang: en

@role
  You are a helpful assistant.

@context
  Service A: port 8080
  Service B: port 9090

@rules
  NEVER: restart without confirmation
  ALWAYS: log all actions

@fallback
  signals: I think, probably, maybe
  default: I don't have verified information.

@chain
  1. Is the service in @context? → if not: fallback
  2. Is the action safe? → if not: warn

@task
  {{ user_message }}

@output
  format: numbered_steps
  max_words: 100
`

	doc, err := parser.ParseString(content)
	
	require.NoError(t, err)
	require.NotNil(t, doc)
	
	assert.Equal(t, "1.0", doc.Meta.Version)
	assert.Equal(t, "en", doc.Meta.Lang)
	assert.NotEmpty(t, doc.Role)
	assert.Len(t, doc.Context, 2)
	assert.Len(t, doc.Rules, 2)
	assert.Len(t, doc.Fallback.Signals, 3)
	assert.Len(t, doc.Chain, 2)
}

// TestParseString_EmptyDocument tests parsing empty document
func TestParseString_EmptyDocument(t *testing.T) {
	content := ""
	
	doc, err := parser.ParseString(content)
	
	assert.NoError(t, err)
	assert.NotNil(t, doc)
}

// TestParseString_OnlyComments tests parsing document with only comments
func TestParseString_OnlyComments(t *testing.T) {
	content := `
# This is a comment
# Another comment
`
	
	doc, err := parser.ParseString(content)
	
	assert.NoError(t, err)
	assert.NotNil(t, doc)
}

// TestParseString_InvalidSectionName tests parsing with invalid section name
func TestParseString_InvalidSectionName(t *testing.T) {
	content := `
@
  content here
`
	
	_, err := parser.ParseString(content)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty section name")
}

// TestParseString_MultipleSections tests parsing with multiple sections
func TestParseString_MultipleSections(t *testing.T) {
	content := `
@meta
  version: 1.0

@role
  Role description

@context
  Item 1: value 1
  Item 2: value 2

@rules
  NEVER: do something
  ALWAYS: do something else

@fallback
  signals: signal1, signal2
  default: default action

@chain
  1. Check condition → if not: fallback

@task
  Do something

@output
  format: plain
`

	doc, err := parser.ParseString(content)
	
	require.NoError(t, err)
	require.NotNil(t, doc)
	
	assert.Equal(t, "1.0", doc.Meta.Version)
	assert.NotEmpty(t, doc.Role)
	assert.Len(t, doc.Context, 2)
	assert.Len(t, doc.Rules, 2)
	assert.Len(t, doc.Fallback.Signals, 2)
	assert.Len(t, doc.Chain, 1)
}

// TestParseString_ContextWithSpaces tests parsing context entries with spaces in keys
func TestParseString_ContextWithSpaces(t *testing.T) {
	content := `
@context
  PostgreSQL 16: port 5432, logs /var/log/postgresql/
  Redis 7.2: port 6379, config /etc/redis/redis.conf
`

	doc, err := parser.ParseString(content)
	
	require.NoError(t, err)
	require.NotNil(t, doc)
	
	assert.Len(t, doc.Context, 2)
	assert.Equal(t, "PostgreSQL 16", doc.Context[0].Key)
	assert.Equal(t, "port 5432, logs /var/log/postgresql/", doc.Context[0].Value)
	assert.Equal(t, "Redis 7.2", doc.Context[1].Key)
	assert.Equal(t, "port 6379, config /etc/redis/redis.conf", doc.Context[1].Value)
}

// TestParseString_RuleTypes tests parsing different rule types
func TestParseString_RuleTypes(t *testing.T) {
	content := `
@rules
  NEVER: do something bad
  ALWAYS: do something good
  IF condition is met: take action
  MAX ITEMS: 10
  MIN ITEMS: 2
`

	doc, err := parser.ParseString(content)
	
	require.NoError(t, err)
	require.NotNil(t, doc)
	
	assert.Len(t, doc.Rules, 5)
	assert.Equal(t, "NEVER", doc.Rules[0].Type)
	assert.Equal(t, "ALWAYS", doc.Rules[1].Type)
	assert.Equal(t, "IF", doc.Rules[2].Type)
	assert.Equal(t, "MAX", doc.Rules[3].Type)
	assert.Equal(t, "MIN", doc.Rules[4].Type)
}

// TestParseString_FallbackFields tests parsing all fallback fields
func TestParseString_FallbackFields(t *testing.T) {
	content := `
@fallback
  signals: I think, probably, maybe
  default: I don't know
  unknown: That's not in my knowledge
  conflict: There's a conflict
  escalate: admin-on-call
`

	doc, err := parser.ParseString(content)
	
	require.NoError(t, err)
	require.NotNil(t, doc)
	
	assert.Len(t, doc.Fallback.Signals, 3)
	assert.Equal(t, "I don't know", doc.Fallback.DefaultAction)
	assert.Equal(t, "That's not in my knowledge", doc.Fallback.UnknownAction)
	assert.Equal(t, "There's a conflict", doc.Fallback.ConflictAction)
	assert.Equal(t, "admin-on-call", doc.Fallback.Escalate)
}

// TestParseString_ChainSteps tests parsing chain steps with conditions
func TestParseString_ChainSteps(t *testing.T) {
	content := `
@chain
  1. Is condition met? → if not: fallback
  2. Is action safe? → if not: warn
  3. Are rules satisfied? → if not: restrict
  4. Is output formatted? → if not: adjust
`

	doc, err := parser.ParseString(content)
	
	require.NoError(t, err)
	require.NotNil(t, doc)
	
	assert.Len(t, doc.Chain, 4)
	assert.Equal(t, 1, doc.Chain[0].Index)
	assert.Contains(t, doc.Chain[0].Question, "condition met")
	assert.Equal(t, "fallback", doc.Chain[0].OnFail)
}

// TestParseString_OutputFields tests parsing output configuration
func TestParseString_OutputFields(t *testing.T) {
	content := `
@output
  format: numbered_steps
  max_words: 180
  max_items: 5
  language: es
  include_chain: false
  fields: field1, field2, field3
`

	doc, err := parser.ParseString(content)
	
	require.NoError(t, err)
	require.NotNil(t, doc)
	
	assert.Equal(t, "numbered_steps", doc.Output.Format)
	assert.Equal(t, 180, doc.Output.MaxWords)
	assert.Equal(t, 5, doc.Output.MaxItems)
	assert.Equal(t, "es", doc.Output.Language)
	assert.False(t, doc.Output.IncludeChain)
	assert.Len(t, doc.Output.Fields, 3)
}

// TestParseString_WithVariables tests parsing with template variables
func TestParseString_WithVariables(t *testing.T) {
	content := `
@task
  {{ user_message }}
  Context: {{ context_info }}
  User: {{ user_name }}
`

	doc, err := parser.ParseString(content)
	
	require.NoError(t, err)
	require.NotNil(t, doc)
	
	assert.Contains(t, doc.TaskTemplate, "{{ user_message }}")
	assert.Contains(t, doc.TaskTemplate, "{{ context_info }}")
	assert.Contains(t, doc.TaskTemplate, "{{ user_name }}")
}

// TestParseString_WithCustomSection tests parsing with custom section
func TestParseString_WithCustomSection(t *testing.T) {
	content := `
@meta
  version: 1.0

@custom_section
  item1: value1
  item2: value2

@task
  Do something
`

	doc, err := parser.ParseString(content)
	
	require.NoError(t, err)
	require.NotNil(t, doc)
	
	assert.NotNil(t, doc.Custom)
	assert.Contains(t, doc.Custom, "custom_section")
	assert.Len(t, doc.Custom["custom_section"], 2)
}

// TestParseString_ComplexPLF tests parsing a complex PLF document
func TestParseString_ComplexPLF(t *testing.T) {
	content := `
@meta
  version: 1.0
  lang: es
  description: Complex test agent
  author: test-team
  target: nexus

@role
  Eres un asistente experto.
  Solo usas información verificada.
  Respondes en español técnico.

@context
  Service A: port 8080, health /health
  Service B: port 9090, health /status
  Commands: start, stop, restart, status

@rules
  NEVER: restart production without confirmation
  NEVER: execute unknown commands
  ALWAYS: verify service exists
  ALWAYS: log all actions
  IF production: require explicit confirmation
  MAX COMMANDS: 3
  MIN VERIFICATION: 1

@fallback
  signals: creo que, probablemente, quizás, no estoy seguro
  default: No tengo información verificada.
  unknown: Ese servicio no está en mi contexto.
  conflict: Hay un conflicto de reglas.
  escalate: on-call-admin

@chain
  1. ¿El servicio está en @context? → si no: fallback
  2. ¿El comando es seguro? → si no: evaluar riesgo
  3. ¿Se cumplen las @rules? → si no: aplicar restricción
  4. ¿La respuesta es verificada? → si no: pedir más info

@task
  {{ user_message }}
  Tenant: {{ tenant_id }}
  Context: {{ previous_context }}

@output
  format: numbered_steps
  max_words: 200
  max_items: 5
  language: es
  include_chain: false
`

	doc, err := parser.ParseString(content)
	
	require.NoError(t, err)
	require.NotNil(t, doc)
	
	// Verify metadata
	assert.Equal(t, "1.0", doc.Meta.Version)
	assert.Equal(t, "es", doc.Meta.Lang)
	assert.Equal(t, "Complex test agent", doc.Meta.Description)
	assert.Equal(t, "test-team", doc.Meta.Author)
	assert.Equal(t, "nexus", doc.Meta.Target)
	
	// Verify role
	assert.NotEmpty(t, doc.Role)
	assert.Contains(t, doc.Role, "asistente experto")
	
	// Verify context
	assert.Greater(t, len(doc.Context), 0)
	
	// Verify rules
	assert.Greater(t, len(doc.Rules), 0)
	
	// Verify fallback
	assert.Len(t, doc.Fallback.Signals, 4)
	assert.NotEmpty(t, doc.Fallback.DefaultAction)
	
	// Verify chain
	assert.Len(t, doc.Chain, 4)
	
	// Verify task
	assert.Contains(t, doc.TaskTemplate, "{{ user_message }}")
	
	// Verify output
	assert.Equal(t, "numbered_steps", doc.Output.Format)
	assert.Equal(t, 200, doc.Output.MaxWords)
}

// TestParseString_PreserveFormatting tests that formatting is preserved
func TestParseString_PreserveFormatting(t *testing.T) {
	content := `
@role
  Line 1
  Line 2
  Line 3
`

	doc, err := parser.ParseString(content)
	
	require.NoError(t, err)
	require.NotNil(t, doc)
	
	assert.Contains(t, doc.Role, "Line 1")
	assert.Contains(t, doc.Role, "Line 2")
	assert.Contains(t, doc.Role, "Line 3")
}

// TestParseString_ColonInValue tests parsing values with colons
func TestParseString_ColonInValue(t *testing.T) {
	content := `
@context
  URL: http://localhost:8080/health
  Time: 10:30:45
`

	doc, err := parser.ParseString(content)
	
	require.NoError(t, err)
	require.NotNil(t, doc)
	
	assert.Equal(t, "URL", doc.Context[0].Key)
	assert.Equal(t, "http://localhost:8080/health", doc.Context[0].Value)
	assert.Equal(t, "Time", doc.Context[1].Key)
	assert.Equal(t, "10:30:45", doc.Context[1].Value)
}

// TestParseString_MixedContent tests parsing mixed content types
func TestParseString_MixedContent(t *testing.T) {
	content := `
@context
  Simple key: simple value
  Complex key with spaces: value with: colons: and spaces
  URL: https://example.com:8080/path

@rules
  NEVER: do something
  ALWAYS: do something else
  IF condition: action
`

	doc, err := parser.ParseString(content)
	
	require.NoError(t, err)
	require.NotNil(t, doc)
	
	assert.Len(t, doc.Context, 3)
	assert.Len(t, doc.Rules, 3)
}

