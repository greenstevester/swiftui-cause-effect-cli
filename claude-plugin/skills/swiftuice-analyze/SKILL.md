---
name: swiftuice-analyze
description: This skill should be used when the user asks to review SwiftUI code for performance issues, optimize SwiftUI views, fix re-render problems, or analyze Instruments traces. It can review code directly for anti-patterns OR analyze trace data for deeper insights.
---

# SwiftUI Performance Analysis

This skill helps identify and fix SwiftUI performance anti-patterns. It works in two modes:

1. **Code Review Mode** (no trace needed) - Review Swift files for common anti-patterns
2. **Trace Analysis Mode** - Deep analysis with actual performance data from Instruments

## When to Use This Skill

Use this skill when:
- User asks to review SwiftUI code for performance issues
- User wants to optimize SwiftUI views
- User reports slow UI or excessive re-renders
- User has an Instruments trace file to analyze
- User mentions @State, @ObservedObject, view updates, or re-renders

## Mode 1: Code Review (No Trace Required)

Review SwiftUI source files directly for performance anti-patterns. This is the fastest way to find issues.

### What to Look For

Scan Swift files for these anti-patterns:

| Pattern | Problem | Fix |
|---------|---------|-----|
| `@ObservedObject` on large objects | All views re-render on any property change | Use `@Observable` (iOS 17+) or split state |
| Passing whole model to child views | Unnecessary re-renders when unused properties change | Pass only needed primitives |
| Timer/animation at parent level | Cascades updates to all children | Isolate to affected view only |
| Missing `Equatable` on data-driven views | Can't skip unnecessary re-renders | Implement `Equatable` |
| `@State` for derived values | Redundant state, sync issues | Use computed properties |
| `@EnvironmentObject` for local state | All consumers re-render together | Scope to minimal views |

### Code Review Process

1. Find SwiftUI view files:
   ```bash
   find . -name "*.swift" -exec grep -l "struct.*View" {} \;
   ```

2. For each view, check:
   - What state properties does it observe? (`@State`, `@ObservedObject`, `@EnvironmentObject`)
   - Does it pass whole objects to child views?
   - Does it have timers or frequent updates?
   - Could it implement `Equatable`?

3. Apply fixes from `references/fix-patterns.md`

### Example Review

```swift
// ANTI-PATTERN: Whole object passing
struct UserList: View {
    @ObservedObject var viewModel: AppViewModel  // Large object

    var body: some View {
        ForEach(viewModel.users) { user in
            UserRow(user: user)  // Passes entire User object
        }
    }
}

// FIXED: Pass only needed data
struct UserList: View {
    @ObservedObject var viewModel: AppViewModel

    var body: some View {
        ForEach(viewModel.users) { user in
            UserRow(name: user.name, avatarURL: user.avatarURL)
        }
    }
}
```

## Mode 2: Trace Analysis (Deeper Insights)

For quantitative data on actual re-render counts and update chains, use trace analysis.

### Prerequisites for Trace Analysis

1. **swiftuice installed**: `go install github.com/greenstevester/swiftui-cause-effect-cli/cmd/swiftuice@latest`
2. **macOS with Xcode**
3. **A trace file** (see Recording below)

### Recording a Trace

```bash
swiftuice record -app com.yourcompany.yourapp -time 15s -out trace.trace
```

Or: Xcode → Open Developer Tool → Instruments → SwiftUI template → Record → Save.

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
