# swiftuice - Claude Code Plugin

A Claude Code plugin for SwiftUI performance analysis. Enables AI agents to analyze Instruments traces, detect performance anti-patterns, and suggest fixes with code examples.

## Features

- **Skill: swiftuice-analyze** - Comprehensive workflow for analyzing SwiftUI performance
- **Command: /analyze-swiftui** - Quick analysis of traces in current project
- **Reference docs** - Detailed fix patterns and issue type documentation

## Installation

### Option 1: Install from GitHub (Recommended)

```bash
# In Claude Code
/plugin install github:greenstevester/swiftui-cause-effect-cli/claude-plugin
```

### Option 2: Local Installation

1. Clone the repository:
```bash
git clone https://github.com/greenstevester/swiftui-cause-effect-cli.git
```

2. Install the plugin locally:
```bash
# In Claude Code, from your project directory
claude --plugin-dir /path/to/swiftui-cause-effect-cli/claude-plugin
```

### Option 3: Copy to User Skills

Copy the skill to your user skills directory:
```bash
cp -r claude-plugin/skills/swiftuice-analyze ~/.claude/skills/
```

## Prerequisites

The plugin requires the `swiftuice` CLI tool:

```bash
# Install with Go
go install github.com/greenstevester/swiftui-cause-effect-cli/cmd/swiftuice@latest

# Verify installation
swiftuice version
```

Additional requirements:
- **macOS** (uses `xcrun xctrace`)
- **Xcode** with command-line tools installed

## Usage

### Using the Skill (Automatic)

The skill triggers automatically when you:
- Ask about SwiftUI performance
- Mention Instruments traces
- Want to optimize view updates
- Report excessive re-renders

Example prompts:
- "Analyze the SwiftUI performance in this trace"
- "Why is my ItemRow view updating so much?"
- "Help me optimize the cause-effect graph"
- "Fix the performance issues in my iOS app"

### Using the Command

```bash
/analyze-swiftui
```

Or with arguments:
```bash
/analyze-swiftui path/to/trace.trace -source ./MyApp
```

### Direct CLI Usage

```bash
# Analyze a trace with source correlation
swiftuice analyze -in trace.trace -source ./MyApp -stdout

# Record a new trace
swiftuice record -app com.company.myapp -time 15s -out trace.trace

# Generate human-readable summary
swiftuice summarize -in trace.trace -out summary.md -dot graph.dot
```

## What Gets Detected

| Issue Type | Description | Severity Range |
|------------|-------------|----------------|
| `excessive_rerender` | Views updating too frequently | Medium - Critical |
| `cascading_update` | Single state change triggers many views | Medium - High |
| `timer_cascade` | Timer causing broad UI updates | Medium - High |
| `deep_dependency_chain` | Long update propagation paths | Low - Medium |
| `whole_object_passing` | Passing entire objects when primitives suffice | Low - Medium |

## Fix Suggestions

Each detected issue includes fix suggestions with:
- **Approach**: Description of the fix strategy
- **Code Before**: Example of problematic code
- **Code After**: Example of fixed code
- **Effort**: Low, Medium, or High
- **Impact**: Expected improvement level

## Plugin Structure

```
claude-plugin/
├── .claude-plugin/
│   └── plugin.json          # Plugin manifest
├── commands/
│   └── analyze-swiftui.md   # /analyze-swiftui command
├── skills/
│   └── swiftuice-analyze/
│       ├── SKILL.md         # Main skill definition
│       └── references/
│           ├── fix-patterns.md    # Detailed fix patterns
│           └── issue-types.md     # Issue type documentation
└── README.md                # This file
```

## Example Output

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
  ],
  "agent_instructions": {
    "task_description": "SwiftUI performance issues detected...",
    "priority": ["[high] Excessive re-renders in ItemRow"],
    "constraints": ["Maintain existing functionality..."]
  }
}
```

## Contributing

Contributions welcome! Please see the main repository for guidelines:
https://github.com/greenstevester/swiftui-cause-effect-cli

## License

MIT
