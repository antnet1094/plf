# PLF-JS: Node.js Bindings for Prompt Language Format

Usa el motor de ingeniería de prompts **PLF** (escrito en Go) directamente en tus aplicaciones de **Node.js, Express o Next.js**.

## Instalación

1.  Descarga la biblioteca compartida para tu sistema operativo desde los [Releases](https://github.com/antnet1094/plf/releases) de GitHub.
2.  Coloca el archivo en la carpeta `lib/` del paquete o en la raíz de tu proyecto.
3.  Instala mediante npm:
    ```bash
    npm install .
    ```

## Uso Rápido

```javascript
const { renderPlf } = require('plf-js');

// Renderizar prompt desde archivo .plf
const variables = { mensaje_usuario: "Hola desde Node.js" };
const prompt = renderPlf("agente.plf", variables);

console.log(prompt);
```

## Beneficios
- **Velocidad Core en Go:** Lógica de parseo nativa.
- **Sin alucinaciones:** Usa el estándar PLF de fronteras de conocimiento.
- **Integración con Ecosistema JS:** Ideal para SaaS de IA.
