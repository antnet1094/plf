# PLF — Prompt Language Format

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![Go Reference](https://pkg.go.dev/badge/github.com/antnet1094/plf.svg)](https://pkg.go.dev/github.com/antnet1094/plf)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Formato estructurado para agentes LLM que reduce alucinaciones mediante fronteras de conocimiento explícitas, protocolos de incertidumbre conductuales y resolución dinámica de contexto (MCP).**

---

## El problema real

Los LLMs no alucinan principalmente por sus pesos — alucinan porque los prompts son ambiguos. Concretamente:

| Causa | Frecuencia | Solución PLF |
|---|---|---|
| Contexto implícito (el modelo "adivina") | Alta | `@context` como frontera explícita |
| Sin protocolo cuando el modelo no sabe | Alta | `@fallback` con señales conductuales |
| Reglas contradictorias no detectadas | Media | Validador semántico pre-envío |
| Razonamiento sin puntos de control | Media | `@chain` con bifurcaciones |
| Output sin estructura definida | Alta | `@output` con **JSON Schema rígido** |

## Instalación

### Como herramienta CLI
```bash
git clone https://github.com/antnet1094/plf
cd plf
go build -o plf ./cmd/plf
sudo mv plf /usr/local/bin/
```

### Como biblioteca de Go
```bash
go get github.com/antnet1094/plf
```

## Uso de la CLI

```bash
# Renderizar un agente con variables y MINIFICACIÓN (ahorro de tokens)
plf render examples/sysadmin.plf \
  --var mensaje_usuario="El servicio PostgreSQL no inicia" \
  --minify

# Resolución dinámica (MCP): Inyecta data en vivo desde archivos o APIs
# En el .plf: @context: Health: MCP: file://logs.txt
plf render test_mcp.plf

# Evaluación de regresión (E2E)
plf eval examples/sysadmin.plf testsuite.json

# Validación semántica y Linting
plf validate examples/whatsapp_router.plf
plf lint examples/sysadmin.plf
```

## Estructura de un archivo `.plf` (v1.0)

```yaml
@meta
  version: 1.0
  lang: es
  target: nexus

@role
  Eres un técnico senior especializado en infra. Solo usas @context.

@context
  PostgreSQL 16: puerto 5432, logs /var/log/postgresql/
  System Metrics: MCP: file://mock_data.txt # Contexto dinámico en vivo

@tools
  reiniciar_servicio: Reinicia un servicio del sistema
    params:
      - service_name (string, required): nombre del servicio

@rules
  NEVER: rm -rf sin confirmar
  ALWAYS: advertir riesgos antes de ejecutar
  MAX COMMANDS: 3

@fallback
  signals: creo que, probablemente, quizás, no estoy seguro
  default: No tengo información verificada. Requiere diagnóstico root.

@chain
  1. ¿El servicio está en @context? → si no: fallback
  2. ¿El comando es seguro? → si no: aplicar restricción

@task
  {{ mensaje_usuario }}

@output
  format: json
  fields: target(string), reason(string: motivo del fallo)
  language: es
```

## Integración en Go (Biblioteca)

```go
package main

import (
    "fmt"
    "github.com/antnet1094/plf/pkg/parser"
    "github.com/antnet1094/plf/pkg/renderer"
    "github.com/antnet1094/plf/pkg/types"
    "github.com/antnet1094/plf/pkg/validator"
)

func main() {
    // 1. Parsear el archivo
    doc, _ := parser.ParseFile("agent.plf")

    // 2. Validar semántica
    issues := validator.Validate(doc)
    if validator.HasErrors(issues) {
        panic("Error de validación")
    }

    // 3. Renderizar con resolución dinámica y minificación
    result, _ := renderer.Render(doc, types.RenderOptions{
        Vars:   map[string]string{"mensaje_usuario": "hola"},
        Format: types.FormatNexus,
        Minify: true,
        Resolver: func(uri string) (string, error) {
            // Implementación de MCP personalizada (file, http, etc)
            return "Datos dinámicos", nil
        },
    })

    // 4. Obtener payload para API
    apiPayload := renderer.ToNexus(result)
    fmt.Println(apiPayload.System)
}
```

## Características Enterprise (Fase 2)

### 1. Resolución Dinámica de Contexto (MCP)
PLF ya no es estático. Puedes definir entradas en `@context` con el prefijo `MCP:` o `DYNAMIC:`. El renderer suspenderá la ejecución, llamará a tu `Resolver` y concatenará la data fresca antes de que el LLM la vea.

### 2. Minificador de Tokens
Soporta un flag `--minify` que comprime el prompt a su densidad entrópica máxima, eliminando ASCII art, redundancia de instrucciones y espacios innecesarios, ahorrando hasta un 20% de tokens.

### 3. Framework de Evaluación (`plf eval`)
Permite ejecutar suites de pruebas automatizadas contra modelos locales (Llama.cpp, Ollama) para medir la tasa de acierto y detectar regresiones en los cambios de prompts.

## Estructura del proyecto

```
plf/
├── cmd/plf/main.go              # CLI (validate, render, eval, lint)
├── pkg/
│   ├── types/types.go           # Tipos centrales
│   ├── parser/parser.go         # Lexer + parser inductivo
│   ├── validator/validator.go   # Validación semántica cruzada
│   ├── renderer/renderer.go     # Compilador a prompt estructurado
│   └── evaluator/evaluator.go   # Motor de tests de regresión LLM
├── examples/                    # Agentes listos para usar
├── docs/                        # Especificación completa y planes
└── README.md
```

## Licencia

MIT — libre para uso comercial, incluyendo en plataformas SaaS.
