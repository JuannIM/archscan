# archscan Pro — Stop AI Coding Drift Before It Becomes Legacy Debt

AI coding assistants like Cursor, Copilot, and Claude Code write code at 100x speed.
Unfortunately, they also introduce architectural rot at 100x speed.

Because LLMs operate within localized file windows, they suffer from context blindness:
they don't know your domain boundaries, they duplicate existing utility functions, they
swallow errors in empty catch blocks to keep tests green, and they leak infrastructure
dependencies straight into your presentation layer.

**archscan** is a high-performance static analysis CLI written in Go that scans 800+
file repositories in under 6 seconds to catch and prevent AI-induced architectural drift.

---

## Why Upgrade to Pro?

The Free tier covers small projects (up to 200 files) with standard anti-pattern checks.
**archscan Pro** is for engineering teams and senior developers who need strict
architectural hygiene across production codebases.

### What's included in Pro:

- **Unlimited scanning** — No file caps. Scan large monorepos in single-digit seconds.
- **Layer Boundary Enforcement** — Define dependency graphs (e.g., Presentation → Domain → Infrastructure) and fail CI when AI code crosses the line.
- **AI Guardrail Generation** — Auto-generates `.cursorrules` and `CLAUDE.md` files from your real codebase structure. Feed architectural constraints directly into Cursor and Claude Code to prevent bad patterns before they're generated.
- **JSON & Markdown output** — Structured reports for GitHub Actions PR comments, dashboards, and automated pipelines.
- **Priority email support** — archscan@proton.me

---

## Transparent Pricing

- **Monthly:** $9/month — Cancel anytime.
- **Annual:** $79/year — Save ~27%.

After purchase, you'll receive a license key. Activate with:

```bash
archscan activate --email you@example.com --key ARCHSCAN-XXXX-XXXX-XXXX-XXXX
```

Keep AI development velocity without inheriting an unmaintainable codebase.
