# Plan de Implementación: Fase 2 de Evolución PLF (Nivel Enterprise)

Tras agotar la **Fase 1** (Soporte Nativo de *Function Calling* y *JSON Schemas Rígidos*), la cual solucionó los problemas principales con modelos ML locales (0.8B), hemos analizado en profundidad la base de código actual para identificar las mayores debilidades arquitectónicas y de diseño de *prompting* restantes.

Aquí presentamos el plan arquitectónico para la **Fase 2 de Desarrollo**, atacando las 4 áreas críticas donde PLF aún depende de "texto estático" o carece de fiabilidad analítica.

## 1. Integración con MCP (Model Context Protocol) en `@context`
**El Problema:** La sección `@context` define la "Verdad Absoluta" del agente, pero actualmente es 100% estática. Escribes `PostgreSQL: puerto 5432` y eso se queda hardcodeado. Para agentes reales de soporte o ventas, el contexto cambia constantemente.
**El Objetivo:** Convertir PLF en un orquestador dinámico habilitando el acceso en tiempo de ejecución a servidores MCP y llamadas externas de red, inyectando la información fresca antes de que el LLM la vea.

### Cambios Arquitectónicos Módulo [pkg/renderer/renderer.go](file:///d:/ant/plf-v1.0/plf/pkg/renderer/renderer.go):
- Se debe extender la interfaz [RenderOptions](file:///d:/ant/plf-v1.0/plf/pkg/types/types.go#207-215) para aceptar un `ContextResolver` (*interfaz*).
- Durante [buildSystem()](file:///d:/ant/plf-v1.0/plf/pkg/renderer/renderer.go#41-55), al llegar a [writeContext()](file:///d:/ant/plf-v1.0/plf/pkg/renderer/renderer.go#80-98), si una entrada tiene el prefijo `MCP:` o `DYNAMIC:`, el `renderer` suspenderá la ejecución, hará la petición I/O al servidor MCP/Webhook, obtendrá la data real (ej: *estado actual de la BD o cotización en vivo*), y concatenará el resultado verificado en la cadena del prompt.

## 2. Validación Semántica en `@output` y `@tools`
**El Problema:** El framework [pkg/validator/validator.go](file:///d:/ant/plf-v1.0/plf/pkg/validator/validator.go) apenas revisa si hay nombres duplicados en `@tools` y si falta el `@role`, pero es ciego frente a congruencias semánticas. Si yo digo en mis *rules* `IF error -> ejecuta herramienta alert_ops`, el validador no chequea si `alert_ops` siquiera existe en `@tools`. 
**El Objetivo:** Construir una red de dependencias léxicas en la etapa de validación.

### Cambios Arquitectónicos Módulo [pkg/validator/validator.go](file:///d:/ant/plf-v1.0/plf/pkg/validator/validator.go):
- **[checkCrossSection](file:///d:/ant/plf-v1.0/plf/pkg/validator/validator.go#150-170)**: Validar que cualquier comando o *tool* mencionado en `@chain` o en las `@rules` coincida obligatoriamente con una entrada válida en `@tools`.
- Validar rigurosamente los tipos especificados en el `OutputConfig.Fields` (si dice `edad(integ)`, lanzar error de typo).

## 3. Minificación de Prompts (Ahorro de Tokens)
**El Problema:** Por temas de legibilidad humana, el compilador actual inyecta toneladas de ASCII Art (`╔════╗`), multiplicidad de espacios y retornos de carro. Modelos como Qwen cobran (o gastan RAM) por cada token. El ASCII Art gasta al menos 60 tokens innecesarios.
**El Objetivo:** Soportar un flag de "producción" que empaquete el prompt a su densidad entrópica máxima sin romper las directrices estructurales de los LLMs.

### Cambios Arquitectónicos Módulo [pkg/renderer/renderer.go](file:///d:/ant/plf-v1.0/plf/pkg/renderer/renderer.go):
- Implementar un flag `Minify bool` en [RenderOptions](file:///d:/ant/plf-v1.0/plf/pkg/types/types.go#207-215).
- Si `Minify=true`, deshabilitar la impresión gráfica [writeHeader](file:///d:/ant/plf-v1.0/plf/pkg/renderer/renderer.go#56-70), remover el decorador en [divider()](file:///d:/ant/plf-v1.0/plf/pkg/renderer/renderer.go#304-311), y filtrar dobles saltos de línea `\n\n`. La meta es bajar el peso del System Prompt al menos un 15-20%.

## 4. Framework Cíclico de Evaluación (`plf eval`)
**El Problema:** El desarrollo de agentes de IA es probabilístico. No existe forma actual de probar que cambiar una palabra en la regla `@rules: ALWAYS do good` rompió la capacidad del modelo para usar una herramienta.
**El Objetivo:** Inyectar una herramienta de CLI nativa que consuma casos de prueba estandarizados frente a un LLM y mida latencia / tasa de acierto.

### Cambios Arquitectónicos:
- **NUEVO Módulo `pkg/evaluator/evaluator.go`**: Procesador de YAMLs de testing.
- **[cmd/plf/main.go](file:///d:/ant/plf-v1.0/plf/cmd/plf/main.go)**: Añadir subcomando `eval <agent.plf> <testsuite.yaml>`. Enrutará peticiones simuladas al motor Nexus local o provider genérico y validará automáticamente si la respuesta (ej. formato JSON) es perfecta o no.

--- 

## Conclusión Ejecutiva
Procederemos implementando estas cuatro mejoras sistemáticamente bajo la modalidad de Extracción/Inyección. Comenzaremos con MCP (cambio de mayor impacto funcional inmediato) y terminaremos con CLI/Evaluaciones.
