<div align="center">

# `archscan`

**Architectural drift detection and AI guardrail generation for AI-assisted codebases.**

[![Release](https://img.shields.io/github/v/release/JuannIM/archscan?style=flat-square&color=00D1B2)](https://github.com/JuannIM/archscan/releases)
[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)

[Features](#features) • [Installation](#installation) • [Quickstart](#quickstart) • [CI/CD](#cicd-integration) • [Support the Project](#support-the-project)

</div>

---

## ⚡ The Problem: AI Coding Tools Cause Architectural Drift

Tools like **Cursor, GitHub Copilot, and Claude Code** let you ship features 10x faster. But LLMs have **context blindness** — they optimize for the active file, oblivious to your broader architecture.

Over weeks of AI-accelerated development, codebases suffer from silent decay:

- **Layer boundary leaks** — UI components query the database directly
- **Logic duplication** — AI re-implements helpers that already exist
- **Silent error swallowing** — empty `catch` / `except: pass` / `_ = err` everywhere
- **God functions** — 30-line handlers ballooning to 400+ lines
- **Hardcoded secrets** — mock tokens leaking into production paths
- **Naming drift** — mixed camelCase/snake_case in the same codebase

**`archscan` stops the rot.** Single Go binary, zero dependencies, scans 800+ file repos in under 6 seconds.

---

## 🚀 Features

- **Blazing fast** — pure Go, no runtime deps, `< 6s` on large repos
- **Multi-language** — Go, Python, Java, TypeScript/JavaScript
- **Layer boundary enforcement** — custom architectural rules
- **AI guardrail generation** — auto-generates `.cursorrules` and `CLAUDE.md`
- **Anti-pattern detection** — duplicates, silent errors, god functions, hardcoded secrets
- **CI/CD ready** — JSON & Markdown output, non-zero exit on violations
- **100% Free & Open Source** — no limits, no paywalls.

---

## 📦 Installation

```bash
# macOS / Linux (one-liner)
curl -fsSL https://raw.githubusercontent.com/JuannIM/archscan/main/install.sh | sh

# Go install
go install github.com/JuannIM/archscan@latest

# Or download binary from GitHub Releases
```

---

## 🛠 Quickstart

```bash
# Basic scan
archscan /path/to/repo

# Generate AI rule files (.cursorrules + CLAUDE.md)
archscan /path/to/repo --rules

# Markdown output for CI/PR comments
archscan /path/to/repo --format markdown

# JSON for automated pipelines
archscan /path/to/repo --format json
```

### Sample output

```
⚡ archscan — Architectural Drift Detector
   Scanning: /home/user/myrepo

┌─────────────────────────────────────────────────────┐
│  Scan Results — Go                                  │
│  Files scanned: 835   Violations: 12               │
└─────────────────────────────────────────────────────┘

🔴 CRITICAL (3)
──────────────────────────────────────────────────────

  1. Silent error suppression
     Category: AntiPattern
     Error is captured but silently ignored (_ = err).
     Files:
       → cmd/auth/helpers.go:450
     Fix: Handle the error explicitly: log it, wrap it, or propagate it.

  2. Potential hardcoded credential or secret
     Category: AntiPattern
     Files:
       → internal/payment/client.go:18
     Fix: Use environment variables or a secrets manager.

  3. Layer violation: presentation → infrastructure
     Category: BoundaryViolation
     Files:
       → pkg/api/handler.go:12
     Fix: Route through the application/service layer.

Architecture Health Score: 62/100 (Degraded)
```

---

## 🔄 CI/CD Integration

```yaml
# .github/workflows/archscan.yml
name: Architecture Check
on:
  pull_request:
    branches: [main]

jobs:
  archscan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install archscan
        run: curl -fsSL https://raw.githubusercontent.com/JuannIM/archscan/main/install.sh | sh

      - name: Run scan
        run: archscan . --format markdown --rules
```

---

## ☕ Support the Project

`archscan` is 100% free and open source. If it saved you or your team hours of untangling AI-generated spaghetti code, consider supporting the development!

- **GitHub Sponsors:** [github.com/sponsors/JuannIM](https://github.com/sponsors/JuannIM)
- **Buy me a coffee:** [buymeacoffee.com/JuannIM](https://buymeacoffee.com/JuannIM)

---

## 🤝 Community

If you find this tool helpful, **please consider giving it a ⭐️ on GitHub!** It helps the project reach more developers.

- **Bugs & Feature Requests:** Found an issue or have an idea? Please [open an issue on GitHub](https://github.com/JuannIM/archscan/issues).
- **Email:** shanys.mora@gmail.com

---

<div align="center">
Built with Go · Stop shipping architectural debt at AI speed.
</div>
