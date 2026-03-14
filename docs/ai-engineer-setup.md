# AI Engineer Setup

Global AI tooling configuration for OpenCode and GitHub Copilot.

## Overview

This setup provides a consistent AI-assisted engineering environment across IDEs,
sharing a single memory store and enforcing DDD + Clean Architecture standards for Go and React projects.

## File Structure

```
~/.config/opencode/
├── opencode.json          # Global OpenCode config (MCPs, permissions, compaction, agents)
├── AGENTS.md              # Global rules loaded into every session
├── memory.jsonl           # Shared knowledge graph (auto-created on first use)
├── agents/
│   ├── reviewer.md        # @reviewer — read-only code review subagent
│   └── architect.md       # @architect — architecture analysis subagent
└── commands/
    ├── test.md            # /test — run tests and fix failures
    ├── review.md          # /review — invoke reviewer subagent on changes
    ├── plan-feature.md    # /plan-feature <name> — plan a feature across all layers
    └── check-arch.md      # /check-arch — full architecture audit

~/.config/ai-rules/
└── instructions.md        # Single source of truth for shared coding rules (all platforms)

~/.vscode/
└── mcp.json               # Shared MCP config for GitHub Copilot

.github/
├── copilot-instructions.md          # Symlink → ~/.config/ai-rules/instructions.md
└── instructions/
    └── go.instructions.md           # Go-specific rules (applyTo: **/*.go)
```

## MCPs

| Server | Enabled globally | Purpose |
|--------|-----------------|---------|
| `memory` | No (per-agent) | Persistent knowledge graph across sessions |
| `sequential-thinking` | No (per-agent) | Multi-step reasoning for complex tasks |
| `context7` | No (manual) | Documentation lookup |

MCPs are registered globally but disabled as tools by default (`tools: { "memory_*": false, ... }`).
They are selectively enabled per-agent to avoid context bloat:
- `build` agent: memory tools enabled
- `plan` agent: memory + sequential-thinking enabled
- `reviewer` subagent: memory tools enabled
- `architect` subagent: memory + sequential-thinking enabled

### Shared Memory

All platforms (OpenCode, Copilot) use the same file:
```
/home/ale/.config/opencode/memory.jsonl
```

This is configured via:
- **OpenCode**: `environment.MEMORY_FILE_PATH` in `opencode.json`
- **Copilot**: `env.MEMORY_FILE_PATH` in `~/.vscode/mcp.json`

## GitHub Copilot

### Instruction files

Copilot reads two layers of instructions from `.github/`:

| File | Scope | Mechanism |
|------|-------|-----------|
| `.github/copilot-instructions.md` | All files | Symlink to `~/.config/ai-rules/instructions.md` |
| `.github/instructions/go.instructions.md` | `**/*.go` only | `applyTo` frontmatter |

The symlink approach keeps a single source of truth. To add a new project:
```bash
ln -sf ~/.config/ai-rules/instructions.md /path/to/project/.github/copilot-instructions.md
```

### Updating shared rules

Edit `~/.config/ai-rules/instructions.md` — all symlinked projects pick up the change immediately.
Project-specific overrides go in `.github/instructions/<topic>.instructions.md` with `applyTo` frontmatter.

## Custom Agents

### `@reviewer` (subagent)

Read-only. Reviews code against the checklist:
- Architecture layer violations
- Go/React coding standards
- Security issues (SQL injection, secret leaks)
- Test coverage gaps

Output grouped by: **Critical / Warning / Suggestion**

### `@architect` (subagent)

Read-only. Analyzes architecture using sequential thinking:
- Maps components to their correct layers
- Identifies dependency direction violations
- Checks domain purity and port completeness

## Custom Commands

| Command | Description |
|---------|-------------|
| `/test` | Run full test suite with race detector; fix any failures |
| `/review [scope]` | Invoke `@reviewer` on changed files or specified scope |
| `/plan-feature <name>` | Plan a feature across all 5 architecture layers |
| `/check-arch` | Full architecture audit via `@architect` |

## Permissions

Solo developer setup — permissive by default:
- All bash commands: `allow`
- Exceptions requiring confirmation: `rm -rf *`, `git push --force*`
- `.env` files: never read (enforced via AGENTS.md rules)

## Compaction

```json
{ "auto": true, "prune": true, "reserved": 10000 }
```

Aggressive compaction to keep context windows efficient. Old tool outputs are pruned automatically.
