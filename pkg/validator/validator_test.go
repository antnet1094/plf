package validator_test

import (
	"testing"

	"github.com/antnet1094/plf/pkg/parser"
	"github.com/antnet1094/plf/pkg/types"
	"github.com/antnet1094/plf/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidate_ValidPLF tests validation of a valid PLF document
func TestValidate_ValidPLF(t *testing.T) {
	content := `
@meta
  version: 1.0
  lang: es

@role
  You are a helpful assistant.

@context
  Service A: port 8080

@rules
  NEVER: do something bad
  ALWAYS: do something good

@fallback
  signals: I think, probably
  default: I don't know

@chain
  1. Is it safe? → if not: fallback

@task
  {{ message }}

@output
  format: plain
`

	doc, err := parser.ParseString(content)
	require.NoError(t, err)

	issues := validator.Validate(doc)

	// Should have no errors (may have warnings)
	hasErrors := false
	for _, issue := range issues {
		if issue.Severity == "error" {
			hasErrors = true
			break
		}
	}

	assert.False(t, hasErrors, "Valid PLF should have no errors")
}

// TestValidate_MissingRequiredSections tests validation with missing sections
func TestValidate_MissingRequiredSections(t *testing.T) {
	content := `
@meta
  version: 1.0

@role
  You are an assistant.

@context
  Service A: port 8080
`

	doc, err := parser.ParseString(content)
	require.NoError(t, err)

	issues := validator.Validate(doc)

	// Should have errors for missing sections
	assert.True(t, len(issues) > 0, "Should detect missing sections")
	
	hasSectionError := false
	for _, issue := range issues {
		if issue.Section == "document" && issue.Severity == "error" {
			hasSectionError = true
			break
		}
	}
	assert.True(t, hasSectionError, "Should report missing required sections")
}

// TestValidate_EmptyRole tests validation with empty role
func TestValidate_EmptyRole(t *testing.T) {
	content := `
@role

@context
  Service A: port 8080

@rules
  NEVER: bad thing

@fallback
  signals: I think
  default: I don't know

@chain
  1. Check → if not: fallback

@task
  {{ message }}

@output
  format: plain
`

	doc, err := parser.ParseString(content)
	require.NoError(t, err)

	issues := validator.Validate(doc)

	// Should have warning or error for empty role
	hasRoleIssue := false
	for _, issue := range issues {
		if issue.Section == "role" {
			hasRoleIssue = true
			break
		}
	}
	assert.True(t, hasRoleIssue, "Should detect empty role")
}

// TestValidate_EmptyContext tests validation with empty context
func TestValidate_EmptyContext(t *testing.T) {
	content := `
@role
  You are an assistant.

@context

@rules
  NEVER: bad thing

@fallback
  signals: I think
  default: I don't know

@chain
  1. Check → if not: fallback

@task
  {{ message }}

@output
  format: plain
`

	doc, err := parser.ParseString(content)
	require.NoError(t, err)

	issues := validator.Validate(doc)

	// Should have warning or error for empty context
	hasContextIssue := false
	for _, issue := range issues {
		if issue.Section == "context" {
			hasContextIssue = true
			break
		}
	}
	assert.True(t, hasContextIssue, "Should detect empty context")
}

// TestValidate_NoFallbackSignals tests validation with no fallback signals
func TestValidate_NoFallbackSignals(t *testing.T) {
	content := `
@role
  You are an assistant.

@context
  Service A: port 8080

@rules
  NEVER: bad thing

@fallback
  default: I don't know

@chain
  1. Check → if not: fallback

@task
  {{ message }}

@output
  format: plain
`

	doc, err := parser.ParseString(content)
	require.NoError(t, err)

	issues := validator.Validate(doc)

	// Should have error for missing signals
	hasSignalsIssue := false
	for _, issue := range issues {
		if issue.Section == "fallback" {
			hasSignalsIssue = true
			break
		}
	}
	assert.True(t, hasSignalsIssue, "Should detect missing fallback signals")
}

// TestValidate_NoFallbackDefault tests validation with no fallback default
func TestValidate_NoFallbackDefault(t *testing.T) {
	content := `
@role
  You are an assistant.

@context
  Service A: port 8080

@rules
  NEVER: bad thing

@fallback
  signals: I think, probably

@chain
  1. Check → if not: fallback

@task
  {{ message }}

@output
  format: plain
`

	doc, err := parser.ParseString(content)
	require.NoError(t, err)

	issues := validator.Validate(doc)

	// Should have error for missing default action
	hasDefaultIssue := false
	for _, issue := range issues {
		if issue.Section == "fallback" && issue.Message != "" {
			hasDefaultIssue = true
			break
		}
	}
	assert.True(t, hasDefaultIssue, "Should detect missing fallback default")
}

// TestValidate_EmptyChain tests validation with empty chain
func TestValidate_EmptyChain(t *testing.T) {
	content := `
@role
  You are an assistant.

@context
  Service A: port 8080

@rules
  NEVER: bad thing

@fallback
  signals: I think
  default: I don't know

@chain

@task
  {{ message }}

@output
  format: plain
`

	doc, err := parser.ParseString(content)
	require.NoError(t, err)

	issues := validator.Validate(doc)

	// Should have warning or error for empty chain
	hasChainIssue := false
	for _, issue := range issues {
		if issue.Section == "chain" {
			hasChainIssue = true
			break
		}
	}
	assert.True(t, hasChainIssue, "Should detect empty chain")
}

// TestValidate_NoOutputFormat tests validation with no output format
func TestValidate_NoOutputFormat(t *testing.T) {
	content := `
@role
  You are an assistant.

@context
  Service A: port 8080

@rules
  NEVER: bad thing

@fallback
  signals: I think
  default: I don't know

@chain
  1. Check → if not: fallback

@task
  {{ message }}

@output
`

	doc, err := parser.ParseString(content)
	require.NoError(t, err)

	issues := validator.Validate(doc)

	// Should have warning or error for missing output format
	hasOutputIssue := false
	for _, issue := range issues {
		if issue.Section == "output" {
			hasOutputIssue = true
			break
		}
	}
	assert.True(t, hasOutputIssue, "Should detect missing output format")
}

// TestValidate_InvalidOutputFormat tests validation with invalid output format
func TestValidate_InvalidOutputFormat(t *testing.T) {
	content := `
@role
  You are an assistant.

@context
  Service A: port 8080

@rules
  NEVER: bad thing

@fallback
  signals: I think
  default: I don't know

@chain
  1. Check → if not: fallback

@task
  {{ message }}

@output
  format: invalid_format
`

	doc, err := parser.ParseString(content)
	require.NoError(t, err)

	issues := validator.Validate(doc)

	// Should have error for invalid format
	hasFormatIssue := false
	for _, issue := range issues {
		if issue.Section == "output" && issue.Severity == "error" {
			hasFormatIssue = true
			break
		}
	}
	assert.True(t, hasFormatIssue, "Should detect invalid output format")
}

// TestValidate_ContradictoryRules tests validation with contradictory rules
func TestValidate_ContradictoryRules(t *testing.T) {
	content := `
@role
  You are an assistant.

@context
  Service A: port 8080

@rules
  ALWAYS: do something
  NEVER: do something

@fallback
  signals: I think
  default: I don't know

@chain
  1. Check → if not: fallback

@task
  {{ message }}

@output
  format: plain
`

	doc, err := parser.ParseString(content)
	require.NoError(t, err)

	issues := validator.Validate(doc)

	// Should have warning for contradictory rules
	hasContradiction := false
	for _, issue := range issues {
		if issue.Section == "rules" {
			hasContradiction = true
			break
		}
	}
	assert.True(t, hasContradiction, "Should detect contradictory rules")
}

// TestValidate_UnresolvedVariables tests validation with unresolved variables
func TestValidate_UnresolvedVariables(t *testing.T) {
	content := `
@role
  You are an assistant.

@context
  Service A: port 8080

@rules
  NEVER: bad thing

@fallback
  signals: I think
  default: I don't know

@chain
  1. Check → if not: fallback

@task
  {{ unresolved_variable }}

@output
  format: plain
`

	doc, err := parser.ParseString(content)
	require.NoError(t, err)

	issues := validator.Validate(doc)

	// Should have warning for unresolved variables
	hasVarIssue := false
	for _, issue := range issues {
		if issue.Section == "task" {
			hasVarIssue = true
			break
		}
	}
	assert.True(t, hasVarIssue, "Should detect unresolved variables")
}

// TestValidate_CrossSectionConsistency tests cross-section consistency
func TestValidate_CrossSectionConsistency(t *testing.T) {
	content := `
@role
  Eres un asistente que NUNCA debe hacer algo malo.

@context
  Service A: port 8080

@rules
  ALWAYS: do something bad

@fallback
  signals: I think
  default: I don't know

@chain
  1. Check → if not: fallback

@task
  {{ message }}

@output
  format: plain
`

	doc, err := parser.ParseString(content)
	require.NoError(t, err)

	issues := validator.Validate(doc)

	// Should detect inconsistency between role and rules
	hasInconsistency := false
	for _, issue := range issues {
		if issue.Section == "rules" || issue.Section == "role" {
			hasInconsistency = true
			break
		}
	}
	assert.True(t, hasInconsistency, "Should detect cross-section inconsistency")
}

// TestValidate_ComplexValidPLF tests validation of a complex valid PLF
func TestValidate_ComplexValidPLF(t *testing.T) {
	content := `
@meta
  version: 1.0
  lang: es
  description: Complex valid agent
  target: anthropic

@role
  Eres un técnico de sistemas experto.
  Solo trabajas con información verificada.

@context
  PostgreSQL 16: puerto 5432, logs /var/log/postgresql/
  Redis 7.2: puerto 6379, config /etc/redis/redis.conf
  Comandos seguros: systemctl status|restart, tail -f

@rules
  NEVER: rm -rf sin confirmar
  NEVER: kill -9 sin PID
  ALWAYS: advertir riesgos
  MAX COMMANDS: 3

@fallback
  signals: creo que, probablemente, quizás
  default: No tengo información verificada.
  unknown: Ese servicio no está en mi contexto.
  escalate: admin-on-call

@chain
  1. ¿El servicio está en @context? → si no: fallback
  2. ¿El comando es seguro? → si no: evaluar riesgo
  3. ¿Se cumplen @rules? → si no: aplicar restricción

@task
  {{ mensaje_usuario }}

@output
  format: numbered_steps
  max_words: 180
  language: es
`

	doc, err := parser.ParseString(content)
	require.NoError(t, err)

	issues := validator.Validate(doc)

	// Should have no errors for valid complex PLF
	hasErrors := false
	for _, issue := range issues {
		if issue.Severity == "error" {
			hasErrors = true
			break
		}
	}

	assert.False(t, hasErrors, "Complex valid PLF should have no errors")
}

// TestHasErrors tests the HasErrors helper function
func TestHasErrors(t *testing.T) {
	issues := []types.ValidationIssue{
		{Section: "role", Message: "Empty role", Severity: "warning"},
		{Section: "context", Message: "Empty context", Severity: "error"},
		{Section: "output", Message: "Missing format", Severity: "info"},
	}

	assert.True(t, validator.HasErrors(issues), "Should detect errors")
	
	noErrorIssues := []types.ValidationIssue{
		{Section: "role", Message: "Empty role", Severity: "warning"},
		{Section: "output", Message: "Missing format", Severity: "info"},
	}

	assert.False(t, validator.HasErrors(noErrorIssues), "Should not report errors when none exist")
}

