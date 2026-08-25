<div align="center">

# `archscan`

**Architectural drift detection and AI guardrail generation for AI-assisted codebases.**

[![Release](https://img.shields.io/github/v/release/archscan/archscan?style=flat-square&color=00D1B2)](https://github.com/archscan/archscan/releases)
[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT%20%2F%20Commercial-blue?style=flat-square)](LICENSE)
[![Polar](https://img.shields.io/badge/Polar.sh-Get%20Pro-FF4500?style=flat-square)](https://polar.sh/archscan)

[Features](#features) • [Installation](#installation) • [Quickstart](#quickstart) • [Free vs Pro](#free-vs-pro) • [CI/CD](#cicd-integration) • [Activate Pro](#activating-pro-license)

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
- **Layer boundary enforcement** *(Pro)* — custom architectural rules
- **AI guardrail generation** *(Pro)* — auto-generates `.cursorrules` and `CLAUDE.md`
- **Anti-pattern detection** — duplicates, silent errors, god functions, hardcoded secrets
- **CI/CD ready** — JSON & Markdown output, non-zero exit on violations

---

## 📦 Installation

```bash
# macOS / Linux (one-liner)
curl -fsSL https://raw.githubusercontent.com/archscan/archscan/main/install.sh | sh

# Go install
go install github.com/archscan/archscan@latest

# Or download binary from GitHub Releases
```

---

## 🛠 Quickstart

```bash
# Basic scan
archscan /path/to/repo

# Generate AI rule files (.cursorrules + CLAUDE.md)  — Pro
archscan /path/to/repo --rules

# Markdown output for CI/PR comments  — Pro
archscan /path/to/repo --format markdown

# JSON for automated pipelines  — Pro
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
     Category: BoundaryViolation  [PRO]
     Files:
       → pkg/api/handler.go:12
     Fix: Route through the application/service layer.

Architecture Health Score: 62/100 (Degraded)
```

---

## 📊 Free vs Pro

| Feature | Free | Pro ($9/mo · $79/yr) |
| :--- | :---: | :---: |
| File limit per scan | 200 files | **Unlimited** |
| Silent error / secret detection | ✅ | ✅ |
| God function detection | ✅ | ✅ |
| Naming inconsistency checks | ✅ | ✅ |
| **Layer boundary violation detection** | ❌ | ✅ |
| **Generate `.cursorrules`** | ❌ | ✅ |
| **Generate `CLAUDE.md`** | ❌ | ✅ |
| **JSON & Markdown output** | ❌ | ✅ |
| **CI/CD integration** | exit code only | ✅ full output |
| Support | GitHub Issues | Priority email |

---

## 🔑 Activating Pro License

After purchasing at [polar.sh/archscan](https://polar.sh/archscan), activate with:

```bash
archscan activate --email you@example.com --key ARCHSCAN-XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX
```

Check status anytime:

```bash
archscan license
```

For CI/CD, set the `ARCHSCAN_LICENSE_KEY` environment variable:

```yaml
env:
  ARCHSCAN_LICENSE_KEY: ${{ secrets.ARCHSCAN_LICENSE_KEY }}
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
        run: curl -fsSL https://raw.githubusercontent.com/archscan/archscan/main/install.sh | sh

      - name: Run scan
        env:
          ARCHSCAN_LICENSE_KEY: ${{ secrets.ARCHSCAN_LICENSE_KEY }}
        run: archscan . --format markdown --rules
```

---

## 🤝 Community & Support

- **Get Pro:** [polar.sh/archscan](https://polar.sh/archscan)
- **Issues:** [GitHub Issues](https://github.com/archscan/archscan/issues)
- **Email:** archscan@proton.me

---

<div align="center">
Built with Go · Stop shipping architectural debt at AI speed.
</div>
