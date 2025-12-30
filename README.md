# swiftuice - SwiftUI Performance for Claude Code

Find and fix SwiftUI performance issues — just ask Claude.

## Quick Start

**1. Add the marketplace and install:**

```bash
/plugin marketplace add greenstevester/swiftui-cause-effect-cli
/plugin install swiftuice-analyze@greenstevester-swiftui-cause-effect-cli
```

Or use the interactive installer:
```bash
/plugin
```
Then navigate to **Discover** tab and search for "swiftuice".

**2. Just say:**

> "Find and fix SwiftUI performance issues"

That's it. Claude reviews your code for anti-patterns and suggests fixes with before/after examples.

## What It Does

| Mode | What Happens | Setup Required |
|------|--------------|----------------|
| **Code Review** | Scans SwiftUI views for anti-patterns | None — works immediately |
| **Trace Analysis** | Records runtime data for quantitative insights | CLI + Xcode |

**Start with code review** — no trace file needed. Add trace analysis later for deeper metrics.

## Detected Issues

| Issue Type | Description |
|------------|-------------|
| `excessive_rerender` | Views updating too frequently |
| `cascading_update` | Single state change triggers many views |
| `timer_cascade` | Timer causing broad UI updates |
| `deep_dependency_chain` | Long update propagation paths |
| `whole_object_passing` | Model objects causing unnecessary re-renders |

Each issue includes **suggested fixes** with code examples, effort level, and expected impact.

## Installation Options

### Option 1: Plugin Marketplace (Recommended)

```bash
/plugin marketplace add greenstevester/swiftui-cause-effect-cli
/plugin install swiftuice-analyze@greenstevester-swiftui-cause-effect-cli
```

Or interactively:
```bash
/plugin
```

### Option 2: Project-Level (for teams)

Install to current project only:

```bash
/plugin install swiftuice-analyze@greenstevester-swiftui-cause-effect-cli --scope project
```

## Usage

### Just Ask Claude

Say any of these:

- "Find and fix SwiftUI performance issues"
- "My app UI is slow"
- "Optimize my SwiftUI views"
- "Why is my view re-rendering so much?"

Claude will:
1. Review your SwiftUI code for anti-patterns
2. Offer to record a trace for deeper insights
3. Suggest fixes with before/after code examples

### If You Have a Trace File

> "Analyze trace.trace for performance issues"

### Using the Command

```bash
/swiftuice-analyze
```

Or with arguments:
```bash
/swiftuice-analyze path/to/trace.trace
```

## Trace Analysis (Optional)

For quantitative data on actual re-render counts and update chains, install the CLI:

### Install CLI

```bash
go install github.com/greenstevester/swiftui-cause-effect-cli/cmd/swiftuice@latest
```

**Requirements:** macOS, Xcode with command-line tools, Go 1.22+

### Record a Trace

```bash
# Record for 15 seconds while interacting with your app
swiftuice record -app com.yourcompany.yourapp -time 15s -out trace.trace
```

Then tell Claude: "Analyze trace.trace for performance issues"

### How It Works

```
┌─────────────────────────────────────────────────────────────────┐
│                        swiftuice                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   record ──► export ──► analyze ──► Claude ──► Code Fixes      │
│      │          │           │                                   │
│      ▼          ▼           ▼                                   │
│   .trace    XML/JSON    analysis.json                           │
│                         (issues, fixes,                         │
│                          source correlation)                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

The tool builds a **Cause → State → View** relationship graph from Instruments trace data.

## Fix Suggestions

The skill provides specific fix patterns for each issue type:

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

## Example Output

When analyzing a trace, you get structured JSON like this:

```json
{
  "summary": {
    "performance_score": 65,
    "health_status": "warning",
    "issues_found": 3,
    "critical_issues": 1
  },
  "issues": [
    {
      "type": "excessive_rerender",
      "severity": "high",
      "title": "Excessive re-renders in ItemRow",
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
      "line_number": 12,
      "confidence": 0.95
    }
  ]
}
```

---

## CLI Reference

For power users who want direct CLI access.

### Commands

#### `swiftuice analyze` (AI-friendly)

```bash
swiftuice analyze -in <path> [options]

Options:
  -in       Input directory (from export) or .trace path (required)
  -source   Swift source root for code correlation (optional)
  -out      Output JSON file (default: analysis.json)
  -stdout   Output to stdout instead of file
  -compact  Output compact JSON (for piping)
```

#### `swiftuice record`

```bash
swiftuice record -app <bundle-id> [options]

Options:
  -app       App bundle ID or path to .app (required)
  -template  Instruments template (default: SwiftUI)
  -device    Device name or UDID (optional)
  -time      Recording duration (default: 10s)
  -out       Output .trace path (default: swiftui.trace)
```

#### `swiftuice export`

```bash
swiftuice export -trace <path> [options]

Options:
  -trace   Input .trace path (required)
  -out     Output directory (default: exported)
  -format  Export format: auto|xml|json|csv (default: auto)
```

#### `swiftuice summarize` (human-readable)

```bash
swiftuice summarize -in <path> [options]

Options:
  -in    Input directory or .trace path (required)
  -out   Summary markdown output (default: summary.md)
  -dot   Graphviz .dot output (default: graph.dot)
```

### Direct CLI Workflow

```bash
# 1. Record a trace
swiftuice record -app com.yourcompany.yourapp -time 15s -out trace.trace

# 2. Export to parseable format
swiftuice export -trace trace.trace -out exported/

# 3. Generate AI report
swiftuice analyze -in exported/ -source ./YourApp -out analysis.json

# Or pipe directly
swiftuice analyze -in exported/ -stdout | your-tool
```

---

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

- **AI-First**: Output designed for Claude and other AI tools
- **Actionable**: Every issue includes concrete fix suggestions with code
- **Correlatable**: Trace data linked to source files where possible
- **Standalone**: Works in CI, locally, or as a Claude Code skill

## License

MIT
