# Instrucciones para Agente AI: PLF Session Logger

## Objetivo
Usar PLF (Prompt Language Format) para documentar TODA tu sesión de trabajo:
- Pensamientos (razonamiento interno)
- Acciones (comandos ejecutados, archivos modificados)
- Decisiones (por qué elegiste un camino sobre otro)
- Verificaciones (cómo confirmaste que funciona)
- Respuesta final (solución/documentación)

## Cómo Usar

### 1. Iniciar Sesión
Al comenzar una tarea, genera un archivo PLF de sesión:

```bash
plf-agent -mode generate -session <tu_id> -agent gemini -task "<descripción>"
```

Esto crea `session_<tu_id>.plf` con la estructura completa.

### 2. Durante la Sesión

Registra tus pensamientos:
```bash
plf-agent -mode log-thought -session <tu_id> "Estoy analizando el error porque..."
```

Registra acciones:
```bash
plf-agent -mode log-action -session <tu_id> "Ejecuté: go test ./..."
```

### 3. Al Finalizar

Exporta la sesión completa:
```bash
plf-agent -mode export -session <tu_id>
```

## Estructura del Archivo PLF Generado

```plf
@meta                    # Metadata de la sesión
@thoughts                # Tu razonamiento interno
@actions                 # Comandos y archivos
@verification            # Cómo verificaste tu trabajo
@errors                  # Errores y recuperaciones
@progress                # Progreso de la tarea
@context_updates         # Nuevo conocimiento adquirido
@response                # Tu respuesta/solución final
@session_metadata        # Info técnica de la sesión
```

## Template Completo

El archivo `examples/agent_session_logger.plf` contiene el template completo.
Úsalo como referencia para entender qué documentar.

## Ejemplo de Sesión Completa

```bash
# 1. Iniciar sesión
$ plf-agent -mode generate -session fix-login-001 \
    -agent gemini \
    -task "Fix login bug where users can't authenticate after password reset"

# 2. Analizar código
$ plf-agent -mode log-thought -session fix-login-001 \
    "El problema parece estar en el token de reset. Voy a buscar dónde se genera."

# 3. Leer archivos relevantes
$ plf-agent -mode log-action -session fix-login-001 \
    "Leí: auth/handlers.go - found password_reset() function"

# 4. Encontrar el bug
$ plf-agent -mode log-thought -session fix-login-001 \
    "¡Encontré! El token expira en 0 segundos porque usa time.Duration(0)"

# 5. Ejecutar fix
$ plf-agent -mode log-action -session fix-login-001 \
    "Modifiqué: auth/handlers.go:45 - Changed token expiry to 24*time.Hour"

# 6. Verificar
$ plf-agent -mode log-action -session fix-login-001 \
    "Ejecuté: go test ./auth/... - PASS ✓"

# 7. Exportar
$ plf-agent -mode export -session fix-login-001
# Genera: session_fix-login-001_exported.plf
```

## Para Gemini CLI Específicamente

En Gemini CLI, puedes usar:

```bash
# Iniciar
/exec plf-agent -mode generate -session <id> -task "<task>"

# Cada pensamiento
/exec plf-agent -mode log-thought -session <id> "<tu pensamiento>"

# Cada acción
/exec plf-agent -mode log-action -session <id> "<comando o acción>"

# Ver progreso
/exec plf-agent -mode status -session <id>

# Exportar
/exec plf-agent -mode export -session <id>
```

## Beneficios de Usar PLF para Sesiones

1. **Auditoría Completa**: Registro de TODO el proceso
2. **Reproducibilidad**: Otros pueden entender cómo llegaste a la solución
3. **Learning**: Para mejorar prompts futuros
4. **Debugging**: Si algo falla, hay historial completo
5. **Compliance**: Para requisitos regulatorios o de calidad

## Formato PLF vs Sin Formato

### ❌ SIN PLF (Texto Libre)
```
Analicé el código, encontré el bug, lo arreglé, probé, funcionó.
```

### ✅ CON PLF
```plf
@thoughts
1. [10:00] - Initial analysis: login fails after password reset
2. [10:05] - Found token generation in auth/handlers.go
3. [10:10] - Identified bug: time.Duration(0) = instant expiry
4. [10:15] - Decision: Change to 24h expiry
   Reason: Industry standard for reset tokens
   Alternative: 1h expiry - too short for email delays

@actions
## ACTION #1
- command: read auth/handlers.go
- result: ✓ Found password_reset() at line 45

## ACTION #2  
- command: sed -i 's/time.Duration(0)/24*time.Hour/' auth/handlers.go
- result: ✓ Modified

## ACTION #3
- command: go test ./auth/... -v
- result: ✓ All tests pass

@verification
- Unit tests: PASS
- Integration test: PASS  
- Manual test: PASS
```

---

## Integración con tu Flujo de Trabajo

```
┌─────────────────────────────────────────────────────────────┐
│                    TU SESIÓN DE TRABAJO                     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Gemini CLI                                                 │
│      │                                                      │
│      ├──▶ /exec plf-agent generate ...                     │
│      │                                                      │
│      ├──▶ Analiza, piensas, ejecutas                       │
│      │      │                                               │
│      │      ├──▶ /exec plf-agent log-thought ...           │
│      │      │                                               │
│      │      └──▶ /exec plf-agent log-action ...            │
│      │                                                      │
│      └──▶ /exec plf-agent export ...                       │
│              │                                              │
│              ▼                                              │
│      session_<id>_exported.plf                             │
│              │                                              │
│              ├──▶ Git commit message                        │
│              ├──▶ Documentation                             │
│              └──▶ Code review                               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Comandos Rápidos de Referencia

```bash
# Iniciar nueva sesión
plf-agent -mode generate -session <id> -task "<descripción>"

# Log pensamiento
plf-agent -mode log-thought -session <id> "<texto>"

# Log acción
plf-agent -mode log-action -session <id> "<comando o descripción>"

# Ver estado
plf-agent -mode status -session <id>

# Exportar a PLF
plf-agent -mode export -session <id>

# Ver template
cat examples/agent_session_logger.plf
```