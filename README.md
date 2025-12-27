# swiftuice (SwiftUI Cause & Effect CLI)

A standalone CLI to **record**, **export**, and **summarize** data from the **SwiftUI instrument** in Instruments — including best-effort extraction of the **Cause & Effect Graph** introduced with **Xcode 26**.

> Reality check: Apple's Cause & Effect Graph is primarily an *interactive Instruments feature*.
> This CLI focuses on automation around `xcrun xctrace` and produces a best-effort graph from exported artifacts.

## How It Works

```
┌─────────────────────────────────────────────────────────────────┐
│                        swiftuice                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   record ──────► export ──────► summarize                       │
│      │              │               │                           │
│      ▼              ▼               ▼                           │
│  .trace file    XML/JSON/CSV    summary.md + graph.dot          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

The tool builds a **Cause → State → View** relationship graph:

| Node Type | Examples |
|-----------|----------|
| **Cause** | Tap, gesture, timer, notification, network call |
| **State** | @State, @ObservedObject, @StateObject, @Environment |
| **View**  | View body updates, SwiftUI re-renders |

## Requirements

- **macOS** (required - uses `xcrun`)
- **Xcode** installed with command line tools
- **Go 1.22+** (for building from source)
- **Graphviz** (optional, for rendering `.dot` files to images)

```bash
# Install Graphviz (optional, for graph rendering)
brew install graphviz
```

## Install

```bash
go install github.com/greenstevester/swiftui-cause-effect-cli/cmd/swiftuice@latest
```

Or build locally:

```bash
go build -o swiftuice ./cmd/swiftuice
```

Or using Task:

```bash
brew install go-task
task build
```

## Usage

### 1) Record a trace

```bash
swiftuice record \
  -template "SwiftUI" \
  -app com.yourcompany.yourapp \
  -time 15s \
  -out swiftui.trace
```

If launch-by-bundle-id doesn't work with your setup, record via Instruments.app and save the `.trace`, then start from `export`.

### 2) Export the trace

```bash
swiftuice export -trace swiftui.trace -out exported -format auto
```

### 3) Summarize + graph

```bash
swiftuice summarize -in exported -out summary.md -dot graph.dot
```

Render the graph:

```bash
dot -Tpng graph.dot -o graph.png
```

## Architecture

| Package | Purpose |
|---------|---------|
| `cmd/swiftuice` | CLI entry point, flag parsing, subcommand routing |
| `internal/xctrace` | Wrapper around `xcrun xctrace` commands |
| `internal/export` | Handles trace → file export |
| `internal/graph` | Node/Edge data structures for the cause-effect graph |
| `internal/analyze` | Parses exports, builds graph, renders DOT/Markdown |

### Graph Parsing Strategy

The analyzer uses a multi-strategy approach:

1. **Structured JSON** - If exports contain `nodes`/`edges` arrays, parse directly
2. **Heuristic text scanning** - Falls back to regex-based classification:
   - Matches `@State`, `@ObservedObject` → State nodes
   - Matches `body()`, `View` → View nodes
   - Matches `tap`, `gesture`, `timer` → Cause nodes
3. **Edge inference** - Links `Cause → State → View` based on parse order

## Design notes

- `swiftuice` is intentionally **standalone** (not an MCP server) so it can be used in CI, locally, or called from any agent/MCP stack.
- Exports are **best-effort**: if Xcode changes the export format, the parser falls back to heuristic extraction from text.
- No external Go dependencies - uses only the standard library.

## Development

```bash
# Install Task runner
brew install go-task

# Available commands
task --list

# Common workflows
task build          # Build binary
task test           # Run tests
task lint           # Run linters
task dev            # fmt + vet + test + build
```

## Roadmap ideas

- Add a `--bundle-id` launch mode if your `xctrace` supports it explicitly.
- Add a recorder helper that lists available templates/devices and validates the SwiftUI template exists.
- Add schema adapters once we observe real exports from Xcode 26 SwiftUI instrument.

## License

MIT
