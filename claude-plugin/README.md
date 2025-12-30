# swiftuice - Claude Code Plugin

**Find and fix SwiftUI performance issues** - just ask.

## What It Does

Say "find and fix SwiftUI performance issues" and the skill will:

1. **Review your code** for anti-patterns (works immediately)
2. **Help record a trace** if you want deeper insights
3. **Analyze and suggest fixes** with before/after code examples

No setup required for code review. Trace recording guided step-by-step.

## Installation

### Option 1: Plugin Marketplace (Recommended)

```bash
/plugin marketplace add greenstevester/swiftui-cause-effect-cli
/plugin install swiftuice-analyze@greenstevester-swiftui-cause-effect-cli
```

Or use the interactive installer:
```bash
/plugin
```
Then navigate to **Discover** tab and search for "swiftuice".

### Option 2: Project-Level (for teams)

```bash
/plugin install swiftuice-analyze@greenstevester-swiftui-cause-effect-cli --scope project
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

### Just Ask

Simply say:

> **"Find and fix SwiftUI performance issues"**

The skill handles the full workflow:

1. **Reviews your code** for anti-patterns (immediate results)
2. **Offers to record a trace** if you want deeper insights
3. **Analyzes the data** and suggests fixes with code examples

### Other Ways to Trigger

- "My app UI is slow"
- "Optimize my SwiftUI views"
- "Why is my view re-rendering so much?"
- "Help with SwiftUI performance"

### If You Already Have a Trace

- "Analyze trace.trace for performance issues"

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
