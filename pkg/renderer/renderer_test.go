package renderer_test

import (
	"testing"

	"github.com/antnet1094/plf/pkg/parser"
	"github.com/antnet1094/plf/pkg/renderer"
	"github.com/antnet1094/plf/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRender_BasicPLF tests rendering a basic PLF document
func TestRender_BasicPLF(t *testing.T) {
	content := `
@meta
  version: 1.0
  lang: en

@role
  You are a helpful assistant.

@context
  Service A: port 8080

@rules
  NEVER: do something bad

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

	result, err := renderer.Render(doc, types.RenderOptions{
		Vars:   map[string]string{"message": "Hello"},
		Format: types.FormatRaw,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.System)
	assert.NotEmpty(t, result.User)
	assert.Contains(t, result.User, "Hello")
}

// TestRender_WithVariables tests rendering with template variables
func TestRender_WithVariables(t *testing.T) {
	content := `
@meta
  version: 1.0

@role
  You are {{ role_name }}.

@context
  Service A: {{ service_port }}

@rules
  NEVER: bad thing

@fallback
  signals: I think
  default: I don't know

@chain
  1. Check → if not: fallback

@task
  {{ user_message }}

@output
  format: plain
`

	doc, err := parser.ParseString(content)
	require.NoError(t, err)

	result, err := renderer.Render(doc, types.RenderOptions{
		Vars: map[string]string{
			"role_name":     "an expert assistant",
			"service_port":  "9090",
			"user_message":  "Help me",
		},
		Format: types.FormatRaw,
	})

	require.NoError(t, err)
	assert.Contains(t, result.System, "an expert assistant")
	assert.Contains(t, result.System, "9090")
	assert.Contains(t, result.User, "Help me")
}

// TestRender_UnresolvedVariables tests rendering with unresolved variables
func TestRender_UnresolvedVariables(t *testing.T) {
	content := `
@meta
  version: 1.0

@role
  You are {{ role_name }}.

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
  {{ user_message }}

@output
  format: plain
`

	doc, err := parser.ParseString(content)
	require.NoError(t, err)

	result, err := renderer.Render(doc, types.RenderOptions{
		Vars:   map[string]string{"role_name": "assistant"},
		Format: types.FormatRaw,
	})

	require.NoError(t, err)
	assert.Len(t, result.UnresolvedVars, 1)
	assert.Contains(t, result.UnresolvedVars, "user_message")
}

// TestRender_AnthropicFormat tests rendering for Anthropic API
func TestRender_AnthropicFormat(t *testing.T) {
	content := `
@meta
  version: 1.0
  target: anthropic

@role
  You are a helpful assistant.

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

	result, err := renderer.Render(doc, types.RenderOptions{
		Vars:   map[string]string{"message": "Hello"},
		Format: types.FormatAnthropic,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, result.System)
	assert.NotEmpty(t, result.User)
}

// TestRender_OpenAIFormat tests rendering for OpenAI API
func TestRender_OpenAIFormat(t *testing.T) {
	content := `
@meta
  version: 1.0
  target: openai

@role
  You are a helpful assistant.

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

	result, err := renderer.Render(doc, types.RenderOptions{
		Vars:   map[string]string{"message": "Hello"},
		Format: types.FormatOpenAI,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, result.System)
	assert.NotEmpty(t, result.User)
}

// TestRender_OllamaFormat tests rendering for Ollama API
func TestRender_OllamaFormat(t *testing.T) {
	content := `
@meta
  version: 1.0
  target: ollama

@role
  You are a helpful assistant.

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

	result, err := renderer.Render(doc, types.RenderOptions{
		Vars:   map[string]string{"message": "Hello"},
		Format: types.FormatOllama,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, result.System)
	assert.NotEmpty(t, result.User)
}

// TestRender_RawFormat tests rendering in raw format
func TestRender_RawFormat(t *testing.T) {
	content := `
@meta
  version: 1.0

@role
  You are a helpful assistant.

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

	result, err := renderer.Render(doc, types.RenderOptions{
		Vars:   map[string]string{"message": "Hello"},
		Format: types.FormatRaw,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, result.Full)
	assert.Contains(t, result.Full, "Hello")
}

// TestRender_EmptyVars tests rendering with no variables provided
func TestRender_EmptyVars(t *testing.T) {
	content := `
@meta
  version: 1.0

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
  format: plain
`

	doc, err := parser.ParseString(content)
	require.NoError(t, err)

	result, err := renderer.Render(doc, types.RenderOptions{
		Vars:   map[string]string{},
		Format: types.FormatRaw,
	})

	require.NoError(t, err)
	assert.Len(t, result.UnresolvedVars, 1)
	assert.Contains(t, result.UnresolvedVars, "message")
}

// TestRender_ComplexPLF tests rendering a complex PLF document
func TestRender_ComplexPLF(t *testing.T) {
	content := `
@meta
  version: 1.0
  lang: es
  description: Complex agent
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
  Tenant: {{ tenant_id }}

@output
  format: numbered_steps
  max_words: 180
  language: es
`

	doc, err := parser.ParseString(content)
	require.NoError(t, err)

	result, err := renderer.Render(doc, types.RenderOptions{
		Vars: map[string]string{
			"mensaje_usuario": "PostgreSQL no inicia",
			"tenant_id":       "tenant-123",
		},
		Format: types.FormatAnthropic,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, result.System)
	assert.Contains(t, result.User, "PostgreSQL no inicia")
	assert.Contains(t, result.User, "tenant-123")
}

// TestRender_PreserveFormatting tests that formatting is preserved
func TestRender_PreserveFormatting(t *testing.T) {
	content := `
@meta
  version: 1.0

@role
  Line 1
  Line 2
  Line 3

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

	result, err := renderer.Render(doc, types.RenderOptions{
		Vars:   map[string]string{"message": "Hello"},
		Format: types.FormatRaw,
	})

	require.NoError(t, err)
	assert.Contains(t, result.System, "Line 1")
	assert.Contains(t, result.System, "Line 2")
	assert.Contains(t, result.System, "Line 3")
}

// TestRender_ToAnthropicHelper tests the ToAnthropic helper function
func TestRender_ToAnthropicHelper(t *testing.T) {
	content := `
@meta
  version: 1.0
  target: anthropic

@role
  You are a helpful assistant.

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

	result, err := renderer.Render(doc, types.RenderOptions{
		Vars:   map[string]string{"message": "Hello"},
		Format: types.FormatAnthropic,
	})

	require.NoError(t, err)
	
	// Test ToAnthropic helper
	anthropicMsg := renderer.ToAnthropic(result)
	assert.NotNil(t, anthropicMsg)
	assert.NotEmpty(t, anthropicMsg.System)
	assert.NotEmpty(t, anthropicMsg.Messages)
	assert.Equal(t, "user", anthropicMsg.Messages[0].Role)
}

// TestRender_ToOpenAIHelper tests the ToOpenAI helper function
func TestRender_ToOpenAIHelper(t *testing.T) {
	content := `
@meta
  version: 1.0
  target: openai

@role
  You are a helpful assistant.

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

	result, err := renderer.Render(doc, types.RenderOptions{
		Vars:   map[string]string{"message": "Hello"},
		Format: types.FormatOpenAI,
	})

	require.NoError(t, err)
	
	// Test ToOpenAI helper
	openaiMsg := renderer.ToOpenAI(result)
	assert.NotNil(t, openaiMsg)
	assert.NotEmpty(t, openaiMsg.Messages)
	assert.Equal(t, 2, len(openaiMsg.Messages))
	assert.Equal(t, "system", openaiMsg.Messages[0].Role)
	assert.Equal(t, "user", openaiMsg.Messages[1].Role)
}

