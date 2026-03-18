# PLF-PY: Python Bindings for Prompt Language Format

Este paquete permite utilizar el motor de ingeniería de prompts **PLF** (escrito en Go) directamente desde Python. Ideal para integraciones con LangChain, LlamaIndex o cualquier script de IA.

## Instalación

1.  Descarga la biblioteca compartida para tu sistema operativo desde los [Releases](https://github.com/antnet1094/plf/releases) de GitHub:
    -   Windows: `libplf.dll`
    -   Linux: `libplf.so`
    -   macOS: `libplf.dylib`
2.  Coloca el archivo en la raíz de tu proyecto.
3.  Instala el paquete localmente:
    ```bash
    pip install .
    ```

## Uso Rápido

```python
from plf import render_plf

# Sustituye tus .md por .plf
variables = {"mensaje_usuario": "Hola mundo"}
prompt = render_plf("agente.plf", variables)

print(prompt)
```

## Beneficios
- **Velocidad nativa:** El motor core corre en Go.
- **Sin alucinaciones:** Usa todas las reglas estructurales de PLF.
- **Compatible con IA:** Genera prompts listos para enviar a cualquier LLM.
