# PLF — Prompt Language Format

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Formato estructurado para agentes LLM que reduce alucinaciones mediante fronteras de conocimiento explícitas, protocolos de incertidumbre conductuales y cadenas de razonamiento con puntos de bloqueo.**

---

## El problema real

Los LLMs no alucinan principalmente por sus pesos — alucinan porque los prompts son ambiguos. Concretamente:

| Causa | Frecuencia | Solución PLF |
|---|---|---|
| Contexto implícito (el modelo "adivina") | Alta | `@context` como frontera explícita |
| Sin protocolo cuando el modelo no sabe | Alta | `@fallback` con señales conductuales |
| Reglas contradictorias no detectadas | Media | Validador pre-envío |
| Razonamiento sin puntos de control | Media | `@chain` con bifurcaciones |
| Output sin estructura definida | Alta | `@output` con formato estricto |

## Instalación

```bash
git clone https://github.com/antnet1094/plf
cd plf
go mod download
go build -o plf ./cmd/plf
sudo mv plf /usr/local/bin/
```

## Uso rápido

```bash
# Renderizar un agente con variables
plf render examples/sysadmin.plf \
  --var mensaje_usuario="El servicio PostgreSQL no inicia desde las 3am"

# Renderizar para Nexus API (JSON listo para enviar)
plf render examples/sysadmin.plf \
  --var mensaje_usuario="nginx devuelve 502" \
  --format nexus \
  --json \
  --output prompt.json

# Validar antes de desplegar
plf validate examples/whatsapp_router.plf

# Ver estructura parseada
plf inspect examples/restaurant_bot.plf

# Linting con sugerencias
plf lint examples/sysadmin.plf
```

## Estructura de un archivo `.plf`

```
@meta
  version: 1.0
  lang: es
  target: nexus

@role
  Eres un técnico de sistemas Ubuntu 24.04. Solo trabajas con
  información verificada de @context. Español técnico preciso.

@context
  PostgreSQL 16: puerto 5432, logs /var/log/postgresql/
  Redis 7.2: puerto 6379, config /etc/redis/redis.conf
  Comandos seguros: systemctl status|restart, tail -f, journalctl

@rules
  NEVER: rm -rf sin confirmar ruta
  NEVER: kill -9 sin PID específico
  ALWAYS: advertir riesgos antes de ejecutar
  MAX COMMANDS: 3

@fallback
  signals: creo que, probablemente, quizás, no estoy seguro
  default: No tengo información verificada. Requiere diagnóstico root.
  unknown: Este servicio no está en mi contexto verificado.
  escalate: sysadmin-on-call

@chain
  1. ¿El servicio mencionado está en @context? → si no: fallback
  2. ¿El comando está en comandos seguros? → si no: evaluar riesgo
  3. ¿Se cumplen todas las @rules? → si no: aplicar restricción

@task
  {{ mensaje_usuario }}

@output
  format: numbered_steps
  max_words: 180
  language: es
```

## Integración en Go

```go
package main

import (
    "fmt"
    "github.com/antnet1094/plf/pkg/parser"
    "github.com/antnet1094/plf/pkg/renderer"
    "github.com/antnet1094/plf/pkg/types"
    "github.com/antnet1094/plf/pkg/validator"
)

func BuildPrompt(plfPath, userInput string) (*types.RenderResult, error) {
    doc, err := parser.ParseFile(plfPath)
    if err != nil {
        return nil, err
    }

    issues := validator.Validate(doc)
    if validator.HasErrors(issues) {
        return nil, fmt.Errorf("PLF validation failed: %v", issues)
    }

    return renderer.Render(doc, types.RenderOptions{
        Vars:   map[string]string{"mensaje_usuario": userInput},
        Format: types.FormatNexus,
    })
}
```

## Ejemplos incluidos

| Archivo | Descripción | Variables |
|---|---|---|
| `examples/sysadmin.plf` | Soporte técnico Ubuntu/Debian producción | `mensaje_usuario` |
| `examples/whatsapp_router.plf` | Router multi-tenant para plataforma SaaS WhatsApp | `tenant_id`, `mensaje` |
| `examples/restaurant_bot.plf` | Atención al cliente restaurante colombiano | `nombre_restaurante`, `ciudad`, `mensaje_cliente` |

## Diseño de `@fallback` vs threshold numérico

El approach original de `threshold: 0.95` es incorrecto porque los LLMs no exponen probabilidades calibradas al autor del prompt. `@fallback` usa en cambio **señales lingüísticas observables**:

```
# ❌ No funciona — el modelo no tiene acceso a esta métrica
@certainty
  threshold: 0.95

# ✅ Funciona — el modelo SÍ produce estas frases cuando está inseguro
@fallback
  signals: creo que, probablemente, quizás, podría ser
  default: No tengo información verificada para esto.
```

Cuando el modelo está formando una respuesta insegura, naturalmente produce lenguaje hedónico. `@fallback.signals` convierte ese comportamiento emergente en un trigger explícito y controlado.

## Estructura del proyecto

```
plf/
├── cmd/plf/main.go              # CLI (validate, render, inspect, lint)
├── pkg/
│   ├── types/types.go           # Tipos centrales
│   ├── parser/parser.go         # Lexer + parser por sección
│   ├── validator/validator.go   # Validación semántica y cross-section
│   └── renderer/renderer.go     # Render a prompt estructurado
├── examples/
│   ├── sysadmin.plf
│   ├── whatsapp_router.plf
│   └── restaurant_bot.plf
├── docs/
│   └── SPEC.md                  # Especificación completa
└── README.md
```

## Licencia

MIT — libre para uso comercial, incluyendo en plataformas SaaS.

