# PLF — Prompt Language Format

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![Go Reference](https://pkg.go.dev/badge/github.com/antnet1094/plf.svg)](https://pkg.go.dev/github.com/antnet1094/plf)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**PLF es el sustituto estructurado de los archivos `.md` para ingeniería de prompts. Reduce alucinaciones mediante fronteras de conocimiento explícitas, protocolos de incertidumbre conductuales y resolución dinámica de contexto (MCP).**

---

## 🎯 ¿Por qué PLF en lugar de `.md`?

Los archivos Markdown (`.md`) son ideales para humanos, pero **pésimos para LLMs** debido a su ambigüedad estructural. PLF transforma tus instrucciones en un motor determinista:

| Característica | Prompt en `.md` | Prompt en PLF |
|---|---|---|
| **Estructura** | Texto libre, inconsistente | Secciones `@` rígidas |
| **Validación** | Ninguna (error en runtime) | Validador semántico pre-envío |
| **Contexto** | Mezclado con instrucciones | Frontera `@context` aislada |
| **Incertidumbre** | El modelo inventa si no sabe | Protocolo `@fallback` explícito |
| **Variables** | Reemplazo de texto frágil | Template engine con tipado fuerte |

---

## Ecosistema Políglota: Motor Único, Múltiples Lenguajes

PLF está escrito en **Go** para máxima velocidad, pero está diseñado para ser el **Estándar Universal** de prompts. No necesitas reescribir la lógica; puedes consumir el motor de Go desde cualquier lenguaje mediante el **Bridge** (`bridge.go`).

### 🐍 Uso desde Python (IA/Data Science)
Ideal para integrar PLF en **LangChain**, **LlamaIndex** o **Hugging Face**.

```python
from bindings.python.plf import render_plf

# Sustituye tus archivos .md por .plf en Python:
prompt = render_plf("agente.plf", {"user_input": "..."})
```

### 🌐 Uso desde JavaScript/Node.js (Web Apps)
Perfecto para aplicaciones con **Next.js** o **SaaS** de IA concurrentes.

---

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
go get github.com/antnet1094/plf@v1.0.0
```

---

## Características Enterprise (Fase 2)

### 1. Resolución Dinámica de Contexto (MCP)
PLF ya no es estático. Puedes definir entradas en `@context` con el prefijo `MCP:` o `DYNAMIC:`. El renderer inyectará datos reales (ej. logs, métricas de DB) en vivo.

### 2. Minificador de Tokens (`-minify`)
Comprime el prompt a su densidad entrópica máxima, eliminando ASCII art y redundancia, ahorrando hasta un 20% de tokens.

### 3. Framework de Evaluación (`plf eval`)
Ejecuta suites de pruebas automatizadas contra modelos locales para medir la tasa de acierto y detectar regresiones.

---

## Ejemplo de un Agente PLF (v1.0)

```yaml
@meta
  version: 1.0
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
  fields: target(string), reason(string)
```

---

## Compilación del Bridge (Cross-Language)

Para generar la biblioteca compartida para otros lenguajes (requiere `gcc/mingw`):

```bash
# Windows
go build -o libplf.dll -buildmode=c-shared bridge.go

# Linux
go build -o libplf.so -buildmode=c-shared bridge.go
```

---

## 🛠️ Herramientas de Desarrollador (DX)

### Visual Studio Code
Para una mejor experiencia escribiendo archivos `.plf`, instala el soporte oficial de resaltado de sintaxis:
👉 **[Instalar desde el VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=AntNetworks.vscode-plf)**

---

## Licencia

MIT — libre para uso comercial.
