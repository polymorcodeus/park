<p align="center">
  <source media="(prefers-color-scheme: dark)" srcset="images/park-dark.png">
  <source media="(prefers-color-scheme: light)" srcset="images/park-light.png">
  <img alt="Project Logo" src="images/park-dark.png" width="128">
</p>

# park

[![Go Version](https://img.shields.io/github/go-mod/go-version/polymorcodeus/park)](https://go.dev/) [![Build Status](https://img.shields.io/github/actions/workflow/status/polymorcodeus/park/ci.yml?branch=main)](https://github.com/polymorcodeus/park/actions) [![Go Report Card](https://goreportcard.com/badge/github.com/polymorcodeus/park)](https://goreportcard.com/report/github.com/polymorcodeus/park)

**A parking lot for markdown notes, organized as IPAA (Inbox / Projects / Areas / Archive).**

Part of the [polymorcodeus](https://github.com/polymorcodeus) suite of CLI tooling for dotfile and knowledge management.

Park surfaces notes mid-coding-session and keeps the Inbox skim-able through frontmatter synopses. Movement between the four categories is a bidirectional *category decision*, not a linear workflow: a Project that turns out to be open-ended becomes an Area, an Area that gets scoped down becomes a Project, and either can go stale into Archive. The repo itself is structured as a standalone Go module with importable `internal/*` packages, designed to be wired into a larger personal CLI.

## Quick Demo

```bash
park init
park new "revisit dashboard caching approach" \
  -s "current TTL feels wrong, worth a spike" -src my-cli-tool
park new "keep an eye on API rate limits" -c area
park assist
park reclassify 1767786622-idea.md -c projects
```

## Install

```bash
go install github.com/polymorcodeus/park@latest
```

Or build from source:

```bash
git clone https://github.com/polymorcodeus/park.git
cd park
make build
```

Requires Go 1.26.4+.

## Setup

```bash
park init     # creates _inbox/, _projects/, _areas/, _archive/
```

`park init` is idempotent: running it again reports which folders already exist and only creates missing ones. To verify initialization without changing anything -- useful in setup scripts and CI -- use `park check`:

```bash
park check    # exits 0 when all category folders exist, 1 otherwise
```

The default root is the OS user config directory (`~/.config/park` on Linux, `~/Library/Application Support/park` on macOS). Scope notes to a specific location by setting `PARK_ROOT`:

```bash
export PARK_ROOT="$(pwd)/.park"
```

## Storage

Notes are plain markdown files with frontmatter, one folder per category:

```bash
~/.config/park/
├── config              # TOML config
├── _inbox/             # default landing zone
├── _projects/          # bounded work with a defined end state
├── _areas/             # ongoing responsibilities
└── _archive/           # inactive, retained for reference
```

### Frontmatter

```yaml
---
category: inbox
created: 2026-07-16
source: terminal
synopsis: current TTL feels wrong, worth a spike
---
```

Frontmatter is parsed line-by-line — no YAML dependency. Four fields:

| field | purpose |
|-------|---------|
| `category` | redundant with folder, kept so ad-hoc files still self-describe |
| `created` | ISO 8601 date, set at park time |
| `source` | where the note came from (repo, chat, stray idea) |
| `synopsis` | one-line summary — the entire mechanism that makes triage fast |

`synopsis` is the key design decision: read it, decide whether to open the file, move on. `source` lets future-you reconstruct *why* a note exists without re-reading it.

### Reclassification

`park reclassify <file> -c <category>` rewrites frontmatter *before* moving the file. A failed move never leaves a note in a half-updated state. Same-category moves are rejected.

## Commands

| command | purpose |
|---------|---------|
| `init` | create the category folders (idempotent) |
| `check` | verify category folders exist (useful for automation) |
| `new [title]` | park a new note (alias `add`) |
| `assist` | open the tabbed TUI browser |
| `show <file>` | glamour-render a note, no TUI |
| `reclassify <file> -c <cat>` | reclassify a note (alias `recat`) |
| `config` | print the default TOML config |

### `park new` options

| option | purpose |
|--------|---------|
| `-s, --synopsis` | one-line summary (what it is / why it matters) |
| `-src, --source` | origin of the note (repo name, chat, etc.) |
| `-c, --category` | park directly into a specific category |
| `-f, --from-file` | ingest an existing markdown file |
| `--filename` | explicit file slug |

### Ingestion

`park new` accepts content through two non-interactive paths:

**From a file** (`-f, --from-file`):

```bash
park new -f ~/Downloads/meeting-notes.md \
  -s "Q3 planning recap" -src "Slack export"
```

The file is read, wrapped with frontmatter, and moved into the park. The original is left untouched.

**From stdin** (pipe):

```bash
cat research-summary.md | park new "llm context window research" \
  -s "survey of recent papers" -src "Claude"
```

When stdin is not a terminal, `park new` reads the entire input as the note body. If a filename or synopsis is also provided on the command line, the note is parked immediately -- no TUI form opens. When content is piped in without enough metadata to build a complete note, the TUI form opens pre-filled with the piped body for interactive completion.

### Global options

| option | env | default | purpose |
|--------|-----|---------|---------|
| `--park-root <path>` | `PARK_ROOT` | OS user config dir/park | root directory for parked notes |
| `--park-config <path>` | `PARK_CONFIG` | `<root>/config` | path to the config file |

## TUI

`park assist` opens a CLI tabbed browser across all configured categories and allows for interactively moving documents
between categories.

| key | action |
|-----|--------|
| `↑`/`↓` | move selection |
| `enter`/`v` | view full note (glamour render) |
| `→`/`l` / `ctrl+n` | next category |
| `←`/`h` / `ctrl+p` | previous category |
| `?` | toggle full help |
| `i`/`p`/`a`/`x` | reclassify selected note to inbox/projects/areas/archive |
| `/` | filter (category shortcuts disabled while typing) |
| `esc` | cancel filter when filtering |
| `q` / `ctrl+c` | quit |

## Agentic workflows

`park new` supports two ingestion paths that fit naturally into agentic pipelines -- AI-generated summaries, chat logs, coding session captures, or any markdown emitted by another tool.

### Pipe from an agent or script

Any process that writes markdown to stdout can park it in one command:

```bash
# Park a coding session summary generated by an LLM
llm "summarize what we just did" | park new "session recap $(date +%F)" \
  -s "$(llm 'one-line summary of this session')" -src "Claude"

# Park a markdown report from a script
./scripts/generate-daily-report.sh | park new "daily report $(date +%F)" \
  -s "automated daily metrics" -src "cron"
```

When piped content arrives with a filename and synopsis, the note is parked immediately with no TUI -- ideal for unattended scripts, cron jobs, or agent tool calls.

### Ingest an existing file

Markdown files produced by other tools land in the park without manual copying:

```bash
park new -f /tmp/claude-output.md \
  -s "database schema redesign proposal" -src "Claude"
```

The file is read, frontmatter is injected, and it is moved into the category folder. The original file stays in place.

### Verify setup from an agent

Before running other commands, an agent can ensure the park storage is initialized without side effects:

```bash
park check || park init
```

`park check` exits 0 when all configured category folders exist and exits 1 (printing the missing paths) when they do not. This makes it safe to call idempotently in setup scripts and agent tool loops.

### Reclassify from an agent

Automation does not have to stop at ingestion. A scheduled job or agent can reclassify stale inbox items:

```bash
# Move everything older than 30 days from inbox to archive
find "$PARK_ROOT/_inbox" -name "*.md" -mtime +30 -exec park reclassify {} -c archive \;
```

## Configuration

Config lives at `<park-root>/config` as TOML. Run `park config` to print the default:

```toml
default_category = "inbox"

[[categories]]
  name = "inbox"
  path = "~/.config/park/_inbox"
  key = "i"

[[categories]]
  name = "projects"
  path = "~/.config/park/_projects"
  key = "p"

[[categories]]
  name = "areas"
  path = "~/.config/park/_areas"
  key = "a"

[[categories]]
  name = "archive"
  path = "~/.config/park/_archive"
  key = "x"
```

Categories are fully configurable: name, storage path, and TUI hotkey. Add or remove categories to match your workflow. `default_category` is where `park new` lands notes when `-c` is omitted.

## Design

Core decisions that shape the project:

- **Composable architecture.** The six `internal/*` packages (`config`, `note`, `store`, `render`, `model`, `theme`, `fs`) are self-contained with minimal coupling. Importing `internal/model` into another CLI is just wiring commands to exported functions — `cmd/park` exists solely as a reference consumer.
- **No YAML dependency for frontmatter.** Parsed line-by-line as flat `key: value` pairs. One less dependency, zero ambiguity about which YAML dialect.
- **Atomic reclassification.** Frontmatter is rewritten before the file is moved. A failed rename doesn't corrupt state.
- **Synopsis-first triage.** The frontmatter `synopsis` field is the entire inbox UX. No metadata-surfing required.

## Tech stack

| component | library |
|-----------|---------|
| TUI framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| TUI components | [Bubbles](https://github.com/charmbracelet/bubbles) |
| styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| markdown rendering | [Glamour](https://github.com/charmbracelet/glamour) |
| CLI framework | [urfave/cli/v3](https://github.com/urfave/cli) |
| config format | [BurntSushi/toml](https://github.com/BurntSushi/toml) |

## Contributing

```bash
git clone https://github.com/polymorcodeus/park.git
cd park
make check    # fmt, vet, lint, test
```
