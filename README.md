# swiftuice (SwiftUI Cause & Effect CLI)

A standalone CLI to **record**, **export**, and **analyze** data from the **SwiftUI instrument** in Instruments — producing **AI-actionable performance reports** with issue detection, fix suggestions, and source code correlation.

> Designed for AI agents: The `analyze` command produces structured JSON that enables Claude, GPT, or other AI tools to automatically suggest and implement SwiftUI performance fixes.

## How It Works

```
┌─────────────────────────────────────────────────────────────────┐
│                        swiftuice                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   record ──► export ──► analyze ──► AI Agent ──► Code Fixes    │
│      │          │           │                                   │
│      ▼          ▼           ▼                                   │
│   .trace    XML/JSON    analysis.json                           │
│                         (issues, fixes,                         │
│                          source correlation)                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

The tool builds a **Cause → State → View** relationship graph and detects performance anti-patterns:

| Issue Type | Description |
|------------|-------------|
| `excessive_rerender` | Views updating too frequently |
| `cascading_update` | Single state change triggers many views |
| `timer_cascade` | Timer causing broad UI updates |
| `deep_dependency_chain` | Long update propagation paths |
| `whole_object_passing` | Model objects causing unnecessary re-renders |

## Requirements

- **macOS** (required - uses `xcrun`)
- **Xcode** installed with command line tools
- **Go 1.22+** (for building from source)
- **Graphviz** (optional, for rendering `.dot` files)

## Install

```bash
go install github.com/greenstevester/swiftui-cause-effect-cli/cmd/swiftuice@latest
```

Or build locally:

```bash
go build -o swiftuice ./cmd/swiftuice
```

## Quick Start

### For AI Agents (Recommended)

```bash
# Analyze a trace and get AI-actionable JSON
swiftuice analyze -in exported/ -source ./MyApp -out analysis.json

# Or pipe directly to an AI tool
swiftuice analyze -in exported/ -stdout | your-ai-tool
```

### Manual Workflow

```bash
# 1. Record a trace
swiftuice record -app com.yourcompany.yourapp -time 15s -out trace.trace

# 2. Export to parseable format
swiftuice export -trace trace.trace -out exported/

# 3. Generate AI report
swiftuice analyze -in exported/ -source ./YourApp -out analysis.json
```

## AI Integration

The `analyze` command produces a structured JSON report designed for AI consumption:

```json
{
  "version": "1.0",
  "summary": {
    "performance_score": 65,
    "health_status": "warning",
    "issues_found": 3,
    "critical_issues": 1
  },
  "issues": [
    {
      "id": "issue-1",
      "type": "excessive_rerender",
      "severity": "high",
      "title": "Excessive re-renders in ItemRow",
      "description": "View 'ItemRow' updated 47 times...",
      "affected_nodes": ["ItemRow"],
      "suggested_fixes": [
        {
          "approach": "Implement Equatable on View",
          "code_before": "struct ItemRow: View { ... }",
          "code_after": "struct ItemRow: View, Equatable { ... }",
          "effort": "low",
          "impact": "high"
        }
      ]
    }
  ],
  "source_correlations": [
    {
      "trace_label": "ItemRow",
      "file_path": "Views/ItemRow.swift",
      "line_number": 45,
      "confidence": 0.95
    }
  ],
  "agent_instructions": {
    "task_description": "SwiftUI performance issues detected...",
    "priority": ["[high] Excessive re-renders in ItemRow"],
    "constraints": ["Maintain existing functionality..."],
    "success_criteria": ["Reduce view update counts..."]
  }
}
```

### Using as a Claude Code Skill

Create a skill that uses swiftuice:

```bash
# In your .claude/skills/swiftui-perf.md
swiftuice analyze -in $TRACE_DIR -source $PROJECT_ROOT -stdout
```

The AI agent can then:
1. Parse the JSON output
2. Navigate to source files using `source_correlations`
3. Apply fixes from `suggested_fixes` with code examples
4. Verify improvements by re-running analysis

### Output Fields

| Field | Description |
|-------|-------------|
| `summary` | High-level metrics: score, issue counts |
| `issues` | Detected problems with severity and affected nodes |
| `issues[].suggested_fixes` | Concrete code changes with before/after examples |
| `graph` | The cause-effect graph with source file mappings |
| `source_correlations` | Links between trace data and Swift source files |
| `recommendations` | General best-practice suggestions |
| `agent_instructions` | Task description, priorities, constraints for AI |

## Architecture

| Package | Purpose |
|---------|---------|
| `cmd/swiftuice` | CLI entry point, subcommand routing |
| `internal/xctrace` | Wrapper around `xcrun xctrace` |
| `internal/export` | Trace → file export |
| `internal/graph` | Node/Edge data structures |
| `internal/analyze` | Parses exports, builds cause-effect graph |
| `internal/issues` | Detects performance anti-patterns |
| `internal/correlation` | Matches trace data to Swift source files |
| `internal/suggestions` | Fix templates with code examples |
| `internal/aioutput` | Generates structured JSON for AI agents |

## Fix Suggestions

The tool provides specific fix patterns for each issue type:

### Excessive Re-renders
- **Equatable View**: Implement `Equatable` to control re-renders
- **Extract Subview**: Isolate frequently-updating content
- **@Observable**: Migrate to fine-grained observation (iOS 17+)

### Cascading Updates
- **Derived State**: Use computed properties instead of stored state
- **Split State**: Break monolithic state into focused objects

### Timer Cascades
- **TimelineView**: Use SwiftUI's optimized time-based view
- **Limit Scope**: Move timer to only the view that needs it

### Whole-Object Passing
- **Pass Primitives**: Extract only needed properties
- **Focused Protocol**: Define minimal data requirements

## Commands

### `swiftuice analyze` (AI-friendly)

```bash
swiftuice analyze -in <path> [options]

Options:
  -in       Input directory (from export) or .trace path (required)
  -source   Swift source root for code correlation (optional)
  -out      Output JSON file (default: analysis.json)
  -stdout   Output to stdout instead of file
  -compact  Output compact JSON (for piping)
```

### `swiftuice record`

```bash
swiftuice record -app <bundle-id> [options]

Options:
  -app       App bundle ID or path to .app (required)
  -template  Instruments template (default: SwiftUI)
  -device    Device name or UDID (optional)
  -time      Recording duration (default: 10s)
  -out       Output .trace path (default: swiftui.trace)
```

### `swiftuice export`

```bash
swiftuice export -trace <path> [options]

Options:
  -trace   Input .trace path (required)
  -out     Output directory (default: exported)
  -format  Export format: auto|xml|json|csv (default: auto)
```

### `swiftuice summarize` (human-readable)

```bash
swiftuice summarize -in <path> [options]

Options:
  -in    Input directory or .trace path (required)
  -out   Summary markdown output (default: summary.md)
  -dot   Graphviz .dot output (default: graph.dot)
```

## Development

```bash
# Install dependencies
brew install go-task graphviz

# Build
task build

# Run tests
task test

# Run linters
task lint
```

## Design Principles

- **AI-First**: Output is designed for machine consumption, not just humans
- **Actionable**: Every issue includes concrete fix suggestions with code
- **Correlatable**: Trace data is linked to source files where possible
- **Standalone**: No MCP server required - works in CI, locally, or as a skill

## License

MIT
