# swiftuice (SwiftUI Cause & Effect CLI)

A standalone CLI to **record**, **export**, and **summarize** data from the **SwiftUI instrument** in Instruments — including best-effort extraction of the **Cause & Effect Graph** introduced with **Xcode 26**.

> Reality check: Apple’s Cause & Effect Graph is primarily an *interactive Instruments feature*.
> This CLI focuses on automation around `xcrun xctrace` and produces a best-effort graph from exported artifacts.

## Requirements

- macOS
- Xcode installed
- Command line tools available (`xcrun`, `xctrace`)

## Install

```bash
go install github.com/greenstevester/swiftui-cause-effect-cli/cmd/swiftuice@latest
```

Or build locally:

```bash
go build -o swiftuice ./cmd/swiftuice
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

If launch-by-bundle-id doesn’t work with your setup, record via Instruments.app and save the `.trace`, then start from `export`.

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

## Design notes

- `swiftuice` is intentionally **standalone** (not an MCP server) so it can be used in CI, locally, or called from any agent/MCP stack.
- Exports are **best-effort**: if Xcode changes the export format, the parser falls back to heuristic extraction from text.

## Roadmap ideas

- Add a `--bundle-id` launch mode if your `xctrace` supports it explicitly.
- Add a recorder helper that lists available templates/devices and validates the SwiftUI template exists.
- Add schema adapters once we observe real exports from Xcode 26 SwiftUI instrument.

