---
name: swiftuice-analyze
description: This skill should be used when the user asks to analyze SwiftUI performance, optimize SwiftUI views, fix SwiftUI re-render issues, or work with Instruments traces for iOS/macOS apps. It provides AI-actionable analysis with issue detection, fix suggestions, and source code correlation.
---

# SwiftUI Performance Analysis with swiftuice

This skill enables analysis of SwiftUI performance issues using the `swiftuice` CLI tool. It produces structured JSON reports designed for AI consumption, including issue detection, fix suggestions with code examples, and source file correlation.

## When to Use This Skill

Use this skill when:
- User asks to analyze SwiftUI performance
- User has an Instruments trace file (.trace)
- User wants to optimize SwiftUI view updates
- User reports excessive re-renders or slow UI
- User mentions "Cause & Effect" graph from Instruments
- User wants to find performance anti-patterns in SwiftUI code

## Prerequisites

Before using this skill, ensure:
1. **swiftuice is installed**: Check with `swiftuice version`
2. **macOS environment**: This tool only works on macOS
3. **Xcode installed**: Required for `xcrun xctrace`

If swiftuice is not installed, install it:
```bash
go install github.com/greenstevester/swiftui-cause-effect-cli/cmd/swiftuice@latest
```

Or build from source if the repo is available locally.

## Workflow

### Step 1: Identify the Input

Determine what input is available:

| Input Type | Command |
|------------|---------|
| `.trace` file | `swiftuice analyze -in path/to/trace.trace` |
| Exported directory | `swiftuice analyze -in path/to/exported/` |
| Need to record | `swiftuice record -app <bundle-id>` first |

### Step 2: Run Analysis

Execute the analysis with source correlation if Swift source is available:

```bash
# With source correlation (recommended)
swiftuice analyze -in <trace-or-export> -source <swift-project-root> -stdout

# Without source correlation
swiftuice analyze -in <trace-or-export> -stdout
```

The `-stdout` flag outputs JSON directly for parsing.

### Step 3: Parse the JSON Output

The output contains these key sections:

```json
{
  "summary": {
    "performance_score": 65,
    "health_status": "warning",
    "issues_found": 3,
    "critical_issues": 1
  },
  "issues": [...],
  "source_correlations": [...],
  "agent_instructions": {...}
}
```

### Step 4: Implement Fixes

For each issue in the `issues` array:

1. Check the `severity` (critical, high, medium, low)
2. Review `suggested_fixes` with `code_before` and `code_after` examples
3. Use `source_correlations` to navigate to the affected files
4. Apply the fix pattern that best matches the codebase

## Issue Types and Fix Patterns

### Excessive Re-render (`excessive_rerender`)

**Problem**: A view updates too frequently (>10 times per user action)

**Fix approaches**:
1. **Implement Equatable** on the View to control when it re-renders
2. **Extract subview** to isolate frequently-updating content
3. **Use @Observable** (iOS 17+) for fine-grained observation

Example fix:
```swift
// Before
struct ItemRow: View {
    let item: Item
    var body: some View { ... }
}

// After - Add Equatable
struct ItemRow: View, Equatable {
    let item: Item
    var body: some View { ... }

    static func == (lhs: ItemRow, rhs: ItemRow) -> Bool {
        lhs.item.id == rhs.item.id
    }
}
```

### Cascading Update (`cascading_update`)

**Problem**: Single state change triggers many view updates

**Fix approaches**:
1. **Use derived state** with computed properties
2. **Split state** into focused objects
3. **Scope ObservableObject** to only views that need it

### Timer Cascade (`timer_cascade`)

**Problem**: Timer/animation causes broad UI updates

**Fix approaches**:
1. **Use TimelineView** for time-based content
2. **Limit timer scope** to only the view that needs it
3. **Extract animated content** into isolated subview

### Whole-Object Passing (`whole_object_passing`)

**Problem**: Passing entire model objects causes unnecessary re-renders

**Fix approaches**:
1. **Pass primitives** instead of whole objects
2. **Define focused protocols** with minimal data requirements
3. **Use @Bindable** (iOS 17+) for specific properties

## Recording a New Trace

If user needs to create a trace:

```bash
# Record for 15 seconds
swiftuice record -app com.company.appname -time 15s -out trace.trace

# Then analyze
swiftuice analyze -in trace.trace -source ./MyApp -stdout
```

## Human-Readable Summary

For user-facing summary instead of JSON:

```bash
swiftuice summarize -in <trace-or-export> -out summary.md -dot graph.dot
```

This creates:
- `summary.md`: Markdown summary with top issues
- `graph.dot`: Graphviz diagram (render with `dot -Tpng graph.dot -o graph.png`)

## Verification After Fixes

After implementing fixes:

1. Record a new trace with the same user flow
2. Run analysis again
3. Compare `performance_score` (should increase)
4. Verify `issues_found` decreased
5. Check specific view update counts reduced

## Troubleshooting

**"no parseable Cause & Effect data found"**
- The SwiftUI instrument may not have captured data
- Try recording longer or with more UI interaction
- Open trace in Instruments to verify data exists

**swiftuice not found**
- Install with: `go install github.com/greenstevester/swiftui-cause-effect-cli/cmd/swiftuice@latest`
- Or add Go bin to PATH: `export PATH=$PATH:$(go env GOPATH)/bin`

**Source correlation shows 0 matches**
- Verify `-source` points to the Swift project root
- Check that .swift files exist in the directory
- Source correlation uses symbol matching, works best with clear View names

## References

For detailed fix patterns with more code examples, see:
- `references/fix-patterns.md` - Complete fix pattern catalog
- `references/issue-types.md` - Detailed issue type documentation
