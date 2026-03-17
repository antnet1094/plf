# PLF Specification v1.0

**Prompt Language Format** — Estándar estructurado para agentes LLM sin alucinaciones.

---

## Motivación

La mayoría de las alucinaciones en LLMs no provienen de los pesos del modelo sino de prompts mal estructurados:

- Contexto implícito que el modelo debe "adivinar"
- Reglas contradictorias no detectadas antes del envío
- Sin protocolo explícito para manejar incertidumbre
- Fronteras de conocimiento no definidas
- Razonamiento sin puntos de control obligatorios

PLF resuelve esto con un **documento estructurado en secciones** que define exactamente qué sabe el agente, qué puede hacer, cuándo debe detenerse y cómo debe responder.

---

## Estructura del documento

Un archivo `.plf` es texto plano con secciones marcadas por `@nombre_sección`.

```
@meta
  ...

@role
  ...

@context
  ...

@rules
  ...

@fallback
  ...

@chain
  ...

@task
  ...

@output
  ...
```

### Reglas generales de sintaxis

| Elemento | Sintaxis | Descripción |
|---|---|---|
| Sección | `@nombre` | Línea que inicia una sección |
| Par clave-valor | `clave: valor` | La clave no puede tener espacios |
| Entrada de contexto | `Nombre con espacios: valor` | Primer `:` como separador |
| Lista | `- item` | Ítem de lista |
| Paso de cadena | `1. texto → si no: acción` | Paso numerado con bifurcación |
| Comentario | `# texto` | Ignorado por el parser |
| Variable | `{{ nombre }}` | Sustituida en el render |

---

## Secciones

### `@meta` — Metadatos del documento

```
@meta
  version: 1.0
  lang: es
  description: Descripción breve del agente
  author: equipo-ops
  target: nexus
```

| Campo | Valores válidos | Descripción |
|---|---|---|
| `version` | `1.0` | Versión del formato PLF |
| `lang` | código BCP-47 (`es`, `en`) | Idioma por defecto |
| `description` | texto libre | Descripción del agente |
| `author` | texto libre | Autor o equipo |
| `target` | `raw` `core` `nexus` `local` | API de destino |

---

### `@role` — Identidad y propósito del agente

Define **quién** es el agente, **qué puede hacer** y **qué no puede hacer**. Es texto libre.

```
@role
  Eres un técnico de sistemas senior especializado en Ubuntu 24.04.
  Tu único propósito es diagnosticar problemas de los servicios en @context.
  Respondes en español técnico preciso. No improvisas ni asumes.
```

**Buenas prácticas:**
- Incluir el dominio específico (no "eres un asistente útil")
- Mencionar explícitamente las limitaciones de rol
- Mencionar el idioma y tono esperado

---

### `@context` — Frontera de conocimiento verificado

Define el **límite exacto del conocimiento** del agente. El renderer lo envuelve con instrucciones que indican que esta es la **única** fuente verificada.

```
@context
  PostgreSQL 16: puerto 5432, logs /var/log/postgresql/
  Redis 7.2: puerto 6379, config /etc/redis/redis.conf
  Comandos seguros: systemctl status|restart, tail -f, journalctl -u
```

**Por qué funciona:** En vez de esperar que el modelo "sepa" cuándo está en territorio incierto, le definimos explícitamente qué es territorio seguro. Todo lo que no esté aquí debe activar `@fallback`.

**Formato:**
- Usar `Nombre descriptivo: detalles` para entradas con clave
- La clave puede tener espacios (`PostgreSQL 16`)
- Preferir entradas específicas sobre categorías generales

---

### `@rules` — Restricciones de comportamiento

Define restricciones **no negociables** que ningún mensaje de usuario puede anular.

```
@rules
  NEVER: rm -rf sin confirmar ruta completa
  NEVER: kill -9 sin mostrar PID
  ALWAYS: advertir sobre riesgos antes de ejecutar
  ALWAYS: verificar que el comando esté en @context
  IF reinicio en producción: pedir confirmación explícita
  MAX COMMANDS: 3
```

**Directivas soportadas:**

| Directiva | Sintaxis | Descripción |
|---|---|---|
| `NEVER` | `NEVER: <qué>` | Acción prohibida absoluta |
| `ALWAYS` | `ALWAYS: <qué>` | Comportamiento obligatorio |
| `IF` | `IF <condición>: <acción>` | Regla condicional |
| `MAX` | `MAX <SUJETO>: <número>` | Límite numérico superior |
| `MIN` | `MIN <SUJETO>: <número>` | Límite numérico inferior |

**Por qué es más efectivo que indicaciones en el prompt:** Las reglas en `@rules` se renderizan con énfasis visual explícito como "HARD RULES — NON-NEGOTIABLE" y se separan del texto de instrucción general, reduciendo la probabilidad de que sean ignoradas.

---

### `@fallback` — Protocolo de incertidumbre

Reemplaza el concepto incorrecto de `threshold: 0.95` con **señales conductuales observables**.

```
@fallback
  signals: creo que, probablemente, podría ser, quizás, no estoy seguro
  default: No tengo información verificada. Requiere diagnóstico con acceso root.
  unknown: Este servicio no está en mi contexto verificado. No puedo dar una solución segura.
  conflict: Hay un conflicto. Escalo al administrador.
  escalate: sysadmin-on-call
```

| Campo | Descripción |
|---|---|
| `signals` | Frases que indican incertidumbre del modelo (separadas por coma) |
| `default` | Respuesta exacta cuando el modelo detecta incertidumbre |
| `unknown` | Respuesta cuando el tema no está en `@context` |
| `conflict` | Respuesta cuando las reglas se contradicen |
| `escalate` | A quién/qué escalar en última instancia |

**Por qué funciona mejor que una probabilidad:** Los modelos no exponen probabilidades calibradas al prompt. En cambio, sí producen lenguaje hedónico ("creo que", "quizás") cuando están inseguros. `@fallback.signals` aprovecha ese comportamiento real.

---

### `@chain` — Cadena de razonamiento con puntos de bloqueo

Define pasos de razonamiento **con consecuencias estructurales** si fallan. Diferente de "piensa paso a paso" porque cada paso puede redirigir al fallback.

```
@chain
  1. ¿El comando solicitado está en @context? → si no: fallback
  2. ¿Cumple todas las @rules? → si no: explicar restricción
  3. ¿La respuesta es determinista con la información disponible? → si no: pedir más información
  4. ¿Se incluyeron advertencias de riesgo donde corresponde? → si no: agregar advertencias
```

**Formato de paso:**
```
N. <pregunta binaria> → si no: <acción>
```

La acción puede ser:
- `fallback` — invocar el protocolo `@fallback`
- `restrict` — negar la solicitud con explicación
- `warn` — advertir pero continuar
- Texto libre — respuesta específica para ese fallo

**Por qué es más efectivo que CoT simple:** El razonamiento está **vinculado** a acciones. Un "no" en el paso 1 no es solo una nota mental — es una instrucción que redirige el flujo del agente.

---

### `@task` — Plantilla de la tarea

El prompt real que se renderiza con variables. Puede ser texto libre con `{{ variables }}`.

```
@task
  {{ mensaje_usuario }}
```

```
@task
  tenant_id: {{ tenant_id }}
  mensaje entrante: {{ mensaje }}
  contexto previo: {{ historial_resumido }}
```

Las variables se sustituyen en el render. Las variables no resueltas se mantienen como `{{ nombre }}` y se reportan en el resultado.

---

### `@output` — Formato de respuesta

Especifica la estructura exacta que debe tener la respuesta del agente.

```
@output
  format: numbered_steps
  max_words: 180
  max_items: 5
  language: es
  include_chain: false
  fields: agent_destino, confidence, razon
```

| Campo | Valores | Descripción |
|---|---|---|
| `format` | `numbered_steps` `json` `markdown` `plain` `delegation` | Estructura de la respuesta |
| `max_words` | número | Límite de palabras |
| `max_items` | número | Límite de ítems en listas |
| `language` | código BCP-47 | Idioma de la respuesta |
| `include_chain` | `true` `false` | Mostrar cadena de razonamiento |
| `fields` | lista separada por comas | Campos requeridos (para formato `json`) |

---

## Formatos de render

### `raw` (defecto)

Todo en un único string. Útil para modelos vía API directa o stdin.

### `core`

```json
{
  "messages": [
    {"role": "system", "content": "<todo excepto @task>"},
    {"role": "user", "content": "<@task renderizado>"}
  ]
}
```

### `nexus`

```json
{
  "system": "<todo excepto @task>",
  "messages": [
    {"role": "user", "content": "<@task renderizado>"}
  ]
}
```

### `local`

Compatible con el endpoint `/api/chat` de Local.

---

## Uso con el CLI

```bash
# Validar un archivo PLF
plf validate sysadmin.plf

# Renderizar con variables
plf render sysadmin.plf --var mensaje_usuario="El servicio PostgreSQL no inicia"

# Renderizar para la API de Nexus y guardar JSON
plf render sysadmin.plf \
  --var mensaje_usuario="error en nginx" \
  --format nexus \
  --json \
  --output prompt.json

# Inspeccionar la estructura parseada
plf inspect sysadmin.plf

# Ejecutar linting completo
plf lint sysadmin.plf
```

---

## Integración en Go

```go
import (
    "github.com/antnet1094/plf/pkg/parser"
    "github.com/antnet1094/plf/pkg/validator"
    "github.com/antnet1094/plf/pkg/renderer"
    "github.com/antnet1094/plf/pkg/types"
)

// Parsear
doc, err := parser.ParseFile("sysadmin.plf")

// Validar
issues := validator.Validate(doc)
if validator.HasErrors(issues) {
    // manejar errores
}

// Renderizar
result, err := renderer.Render(doc, types.RenderOptions{
    Vars:   map[string]string{"mensaje_usuario": userInput},
    Format: types.FormatNexus,
})

// Usar en llamada a API
payload := renderer.ToNexus(result)
```

---

## Diseño filosófico

### Lo que PLF NO hace

PLF no inventa nuevas capacidades del LLM. No hay:
- Probabilidades calibradas de confianza (los modelos no las exponen)
- Ejecución garantizada del chain (el modelo puede ignorarlo)
- Seguridad criptográfica de las reglas

### Lo que PLF SÍ hace

PLF maximiza la probabilidad de comportamiento correcto al:

1. **Reducir la ambigüedad** — cada sección tiene un propósito exacto
2. **Externalizar la incertidumbre** — el modelo no decide cuándo no saber, el autor del prompt lo define
3. **Crear fronteras explícitas** — `@context` hace visible el límite del conocimiento
4. **Vincular razonamiento con acción** — `@chain` convierte CoT en flujo con consecuencias
5. **Detectar contradicciones antes del envío** — el validador atrapa problemas en tiempo de autoría

---

## Extensiones

PLF soporta secciones personalizadas con `@nombre_custom`. El parser las almacena en `Document.Custom` como `map[string][]string`. El renderer las ignora por defecto pero pueden procesarse manualmente.

---

*Especificación PLF v1.0 — Marzo 2026*

