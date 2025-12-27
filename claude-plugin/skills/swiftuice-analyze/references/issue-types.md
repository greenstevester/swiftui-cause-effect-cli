# SwiftUI Performance Issue Types

Detailed documentation of all issue types detected by swiftuice.

## Issue Severity Levels

| Severity | Score Impact | Description |
|----------|--------------|-------------|
| `critical` | -25 points | Causes significant UI jank, frame drops, or poor UX |
| `high` | -10 points | Noticeable performance degradation |
| `medium` | -3 points | Suboptimal but may not be user-visible |
| `low` | -1 point | Minor inefficiency |
| `info` | 0 points | Informational, potential improvement |

## Issue Type: excessive_rerender

**ID**: `excessive_rerender`

**Description**: A view is updating more frequently than expected, typically due to unnecessary state invalidation or overly broad observation.

### Detection Criteria

- View update count exceeds threshold (default: 10 updates)
- Severity scales with count:
  - 10-24 updates: `medium`
  - 25-49 updates: `high`
  - 50+ updates: `critical`

### Common Causes

1. **Observing entire object when only needing one property**
   - Using `@ObservedObject` on large view models
   - All views re-render when any property changes

2. **Parent view state causing child re-renders**
   - State in parent invalidates all children
   - Even unchanged children re-evaluate body

3. **Frequent timer/animation updates**
   - Timer at high level propagates to all subviews
   - Animation values causing full tree updates

### JSON Example

```json
{
  "id": "issue-1",
  "type": "excessive_rerender",
  "severity": "high",
  "title": "Excessive re-renders in ItemRow",
  "description": "View 'ItemRow' updated 47 times during the trace period",
  "affected_nodes": ["ItemRow"],
  "metrics": {
    "update_count": 47,
    "threshold": 10
  }
}
```

### Recommended Fixes

1. Implement `Equatable` on the View
2. Extract frequently-updating content to subview
3. Use `@Observable` (iOS 17+) for fine-grained observation
4. Pass primitives instead of whole objects

---

## Issue Type: cascading_update

**ID**: `cascading_update`

**Description**: A single state change triggers updates to many views, creating a "cascade" effect through the view hierarchy.

### Detection Criteria

- State node has edges to 4+ view nodes
- Or state change triggers chain of 3+ subsequent updates

### Common Causes

1. **Monolithic state object**
   - Single `@EnvironmentObject` with many properties
   - All views that access it re-render together

2. **Computed state triggering multiple publishes**
   - Setting multiple `@Published` properties in one action
   - Each publish triggers separate update cycle

3. **Deep view hierarchy with shared state**
   - State at root, many leaves depend on it
   - Change at root cascades through entire tree

### JSON Example

```json
{
  "id": "issue-2",
  "type": "cascading_update",
  "severity": "medium",
  "title": "Cascading update from AppState",
  "description": "State change in 'AppState' triggers updates to 6 views",
  "affected_nodes": ["AppState", "View1", "View2", "View3", "View4", "View5", "View6"],
  "metrics": {
    "cascade_width": 6,
    "source_node": "AppState"
  }
}
```

### Recommended Fixes

1. Split state into focused objects
2. Use computed properties instead of stored state
3. Scope `ObservableObject` access to minimal requirements
4. Consider `@Observable` for automatic dependency tracking

---

## Issue Type: timer_cascade

**ID**: `timer_cascade`

**Description**: A timer or periodic event is causing broad UI updates beyond what should be affected.

### Detection Criteria

- Cause node labeled with timer/animation keywords
- Connected to 2+ view nodes through state

### Common Causes

1. **Timer at high level in view hierarchy**
   - Timer in root view or view model
   - State update propagates to all children

2. **Animation using @State instead of withAnimation**
   - Manual timer-based animation
   - Causes full view tree evaluation

3. **Polling data refresh**
   - Periodic data fetch updates shared state
   - All views observing that state re-render

### JSON Example

```json
{
  "id": "issue-3",
  "type": "timer_cascade",
  "severity": "high",
  "title": "Timer causing cascade to ClockView and 3 others",
  "description": "Timer 'Timer fired' triggers updates through 'time state' to 4 views",
  "affected_nodes": ["Timer fired", "time state", "ClockView", "HeaderView", "StatusBar", "Dashboard"],
  "metrics": {
    "views_affected": 4,
    "timer_source": "Timer fired"
  }
}
```

### Recommended Fixes

1. Use `TimelineView` for time-based content
2. Isolate timer to only the view that needs it
3. Extract animated content into separate subview
4. Use `withAnimation` for value-based animations

---

## Issue Type: deep_dependency_chain

**ID**: `deep_dependency_chain`

**Description**: Updates propagate through a long chain of dependencies, increasing latency and complexity.

### Detection Criteria

- Path length from cause to view exceeds threshold (default: 5)
- Multiple intermediate state nodes in chain

### Common Causes

1. **Over-abstracted state management**
   - Many layers between user action and view
   - Each layer adds update overhead

2. **Derived state chains**
   - State A computes State B computes State C
   - Each step is a separate update cycle

3. **Coordinator/mediator patterns**
   - Events routed through multiple objects
   - Unnecessary indirection

### JSON Example

```json
{
  "id": "issue-4",
  "type": "deep_dependency_chain",
  "severity": "low",
  "title": "Deep dependency chain (depth: 6)",
  "description": "Update path: UserTap -> ActionHandler -> StateManager -> ViewModel -> Derived -> FinalView",
  "affected_nodes": ["UserTap", "ActionHandler", "StateManager", "ViewModel", "Derived", "FinalView"],
  "metrics": {
    "chain_depth": 6,
    "threshold": 5
  }
}
```

### Recommended Fixes

1. Flatten state management where possible
2. Remove unnecessary indirection layers
3. Consider direct state mutation for simple cases
4. Use SwiftUI's built-in state management

---

## Issue Type: whole_object_passing

**ID**: `whole_object_passing`

**Description**: Entire model objects are passed to views when only specific properties are needed, causing unnecessary re-renders.

### Detection Criteria

- View receives complex object type
- Only accesses subset of properties
- Object has many properties (heuristic)

### Common Causes

1. **Convenience over optimization**
   - Easier to pass whole object than extract properties
   - Works but creates hidden dependencies

2. **Model objects with many properties**
   - User, Product, Order with 10+ fields
   - View only needs 2-3 fields

3. **Reference semantics confusion**
   - Assuming reference type won't cause re-render
   - SwiftUI tracks identity, not just value

### JSON Example

```json
{
  "id": "issue-5",
  "type": "whole_object_passing",
  "severity": "medium",
  "title": "Whole object 'User' passed to ProfileBadge",
  "description": "View 'ProfileBadge' receives entire User object but only uses 'name' and 'avatarURL'",
  "affected_nodes": ["User", "ProfileBadge"],
  "metrics": {
    "object_type": "User",
    "properties_used": 2,
    "total_properties": 15
  }
}
```

### Recommended Fixes

1. Pass only needed primitive values
2. Create focused protocol for required properties
3. Use struct with only needed fields
4. Consider `@Bindable` for specific property binding (iOS 17+)

---

## Understanding the Cause-Effect Graph

The analysis builds a directed graph:

```
[Cause] → [State] → [View]
```

### Node Types

| Type | Description | Examples |
|------|-------------|----------|
| `cause` | Event that initiates change | Button tap, Timer, Notification |
| `state` | State that was modified | @State, @Published, @Observable |
| `view` | View that re-rendered | Any SwiftUI View body evaluation |

### Edge Semantics

- **Cause → State**: "triggers" - Event caused state change
- **State → View**: "updates" - State change caused view update
- **State → State**: "derives" - One state computed from another

### Analyzing the Graph

1. **Fan-out from state**: Many edges from one state = cascading update
2. **High count on view**: Number indicates re-render frequency
3. **Long paths**: Deep chains indicate complex dependencies
4. **Cycles**: Potential infinite update loops (rare but critical)

## Confidence Levels

Source correlations include confidence scores:

| Confidence | Meaning |
|------------|---------|
| 0.9 - 1.0 | Exact match (struct definition found) |
| 0.7 - 0.9 | High confidence (strong pattern match) |
| 0.5 - 0.7 | Medium confidence (partial match) |
| 0.3 - 0.5 | Low confidence (fuzzy match) |
| < 0.3 | Not reported (too uncertain) |

Use high-confidence correlations to navigate directly to source. For lower confidence, verify manually before applying fixes.
