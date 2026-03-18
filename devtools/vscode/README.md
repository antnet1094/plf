# 🧠 Prompt Language Format (PLF) - Syntax Highlighting

**The official structured language for LLM prompt engineering.** 
Stop messy Markdown prompts. Start building deterministic, safe, and robust AI agents.

---

## 🚀 Key Features

- **🎯 Knowledge Boundaries:** Visual highlighting for `@context` ensuring the model stays within verified data.
- **⚡ Behavioral Directives:** Instant recognition of `NEVER`, `ALWAYS`, `IF`, and `MAX/MIN` rules.
- **🛡️ Uncertainty Fallbacks:** Explicit syntax for `@fallback` protocols to prevent hallucinations.
- **🔗 Reasoning Chains:** Clearly defined `@chain` steps for step-by-step logic execution.
- **📦 Variable Support:** Full highlighting for template variables `{{ variables }}`.

---

## 🛠️ Installation

1. Open **VS Code**.
2. Go to **Extensions** (`Ctrl+Shift+X`).
3. Search for **"PLF"** or **"Prompt Language Format"**.
4. Click **Install**.

---

## 🎯 Why PLF over Markdown?

Markdown is for humans; **PLF is for LLMs.** This extension makes writing PLF as easy as writing code, providing the visual feedback needed to build production-grade prompts.

| Feature | Prompt in .md | Prompt in PLF |
|---|---|---|
| Structure | Free-text, inconsistent | Rigid `@` sections |
| Validation | None (runtime error) | Pre-submission validation |
| Context | Mixed with instructions | Isolated `@context` boundary |
| Uncertainty | Model guesses | Explicit `@fallback` protocol |

---

## 🌍 Community & Support

- **GitHub:** [github.com/antnet1094/plf](https://github.com/antnet1094/plf)
- **License:** MIT

**Empower your agents with PLF. Build the future of AI with structure.**

---

*Developed by AntNetworks*
