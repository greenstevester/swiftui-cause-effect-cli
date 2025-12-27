---
description: Analyze SwiftUI performance from an Instruments trace or exported data
allowed-tools: Bash(*), Read, Write, Glob
---

# SwiftUI Performance Analysis

Analyzing SwiftUI performance using swiftuice...

## Step 1: Check Prerequisites

First, verify swiftuice is installed:

```bash
swiftuice version 2>/dev/null || echo "NOT_INSTALLED"
```

If not installed, install it:

```bash
go install github.com/greenstevester/swiftui-cause-effect-cli/cmd/swiftuice@latest
```

## Step 2: Identify Input

Looking for trace files or exported data...

Check for .trace files:
```bash
find . -name "*.trace" -type d 2>/dev/null | head -5
```

Check for exported directories:
```bash
ls -la exported/ 2>/dev/null || ls -la *exported*/ 2>/dev/null || echo "No exported directory found"
```

## Step 3: Identify Source Root

Find Swift project root:
```bash
find . -name "*.xcodeproj" -o -name "*.xcworkspace" -o -name "Package.swift" 2>/dev/null | head -3
```

## Step 4: Run Analysis

Based on what was found, run the appropriate analysis command:

If trace file found:
```bash
swiftuice analyze -in <TRACE_PATH> -source <SOURCE_ROOT> -stdout
```

If exported directory found:
```bash
swiftuice analyze -in <EXPORTED_DIR> -source <SOURCE_ROOT> -stdout
```

## Step 5: Parse Results

The JSON output contains:

1. **summary**: Overall health score and issue counts
2. **issues**: Detected problems with severity and suggested fixes
3. **source_correlations**: Links to Swift source files
4. **agent_instructions**: Priorities and constraints for fixing

## Step 6: Implement Fixes

For each issue in order of severity (critical first):

1. Navigate to the affected source file using source_correlations
2. Apply the suggested fix pattern from suggested_fixes
3. Use the code_before/code_after examples as templates

## Step 7: Verify

After implementing fixes:
- Record a new trace
- Run analysis again
- Compare performance_score (should increase)

---

**Usage**: `/analyze-swiftui`

**Arguments**:
- Optionally specify trace path: `/analyze-swiftui path/to/trace.trace`
- Optionally specify source root: `/analyze-swiftui -source ./MyApp`
