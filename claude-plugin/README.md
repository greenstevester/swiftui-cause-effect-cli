# swiftuice - Claude Code Plugin

A Claude Code plugin for SwiftUI performance analysis. Review code for anti-patterns **without needing a trace file**, or get deeper insights with Instruments trace analysis.

## Features

- **Code Review** - Review SwiftUI views for anti-patterns (no trace required)
- **Record Trace** - Guided workflow to record an Instruments trace
- **Analyze Trace** - Deep analysis with actual performance data
- **Skill: swiftuice-analyze** - Auto-triggers on SwiftUI performance questions
- **Command: /analyze-swiftui** - Quick analysis of traces in current project
- **Reference docs** - Detailed fix patterns with before/after code examples

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

### For Code Review Mode (No Trace Required)

Just have SwiftUI source files in your project. The skill will review them for anti-patterns.

### For Trace Analysis Mode (Optional, Deeper Insights)

1. **Install swiftuice CLI**:
   ```bash
   go install github.com/greenstevester/swiftui-cause-effect-cli/cmd/swiftuice@latest
   ```

2. **Record a trace**:
   ```bash
   swiftuice record -app com.yourcompany.yourapp -time 15s -out trace.trace
   ```

3. **System requirements**: macOS with Xcode

## Usage

### Code Review (Recommended Starting Point)

Ask the skill to review your SwiftUI code:

- "Review my SwiftUI views for performance issues"
- "Check this view for anti-patterns"
- "Why might my ItemRow be re-rendering too much?"
- "Optimize the state management in my app"

The skill will scan your Swift files and identify issues like:
- Whole object passing
- Missing Equatable
- Overly broad @ObservedObject usage
- Timer/animation cascades

### Record a Trace (For Deeper Analysis)

If you want quantitative data on actual re-render counts:

- "Help me record a SwiftUI performance trace"
- "Profile my app for performance issues"
- "I want to capture a trace of my app"

The skill will guide you through finding your bundle ID and recording.

### Analyze a Trace

If you already have a trace file:

- "Analyze the SwiftUI performance in trace.trace"
- "What's causing the re-renders in this trace?"

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
