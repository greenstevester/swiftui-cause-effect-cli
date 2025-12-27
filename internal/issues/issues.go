// Package issues detects SwiftUI performance anti-patterns from cause-effect graphs.
package issues

import (
	"fmt"
	"sort"
	"strings"

	"github.com/greenstevester/swiftui-cause-effect-cli/internal/graph"
)

// Severity represents how critical an issue is
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// IssueType categorizes the kind of performance problem
type IssueType string

const (
	IssueExcessiveRerender   IssueType = "excessive_rerender"
	IssueCascadingUpdate     IssueType = "cascading_update"
	IssueFrequentTrigger     IssueType = "frequent_trigger"
	IssueDeepDependencyChain IssueType = "deep_dependency_chain"
	IssueWholeObjectPassing  IssueType = "whole_object_passing"
	IssueTimerCascade        IssueType = "timer_cascade"
	IssueStateInBody         IssueType = "state_mutation_in_body"
	IssueUnnecessaryBinding  IssueType = "unnecessary_binding"
)

// Issue represents a detected performance problem
type Issue struct {
	ID          string    `json:"id"`
	Type        IssueType `json:"type"`
	Severity    Severity  `json:"severity"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"`

	// Affected components
	AffectedNodes []string `json:"affected_nodes"`
	CauseChain    []string `json:"cause_chain,omitempty"`

	// Metrics
	UpdateCount     int     `json:"update_count,omitempty"`
	CascadeDepth    int     `json:"cascade_depth,omitempty"`
	PerformanceHint string  `json:"performance_hint,omitempty"`
	Confidence      float64 `json:"confidence"` // 0.0 - 1.0

	// Source correlation (populated later)
	SourceFile string `json:"source_file,omitempty"`
	LineNumber int    `json:"line_number,omitempty"`
}

// Detector analyzes graphs for performance issues
type Detector struct {
	thresholds Thresholds
}

// Thresholds configures detection sensitivity
type Thresholds struct {
	ExcessiveRerenderCount int     // Views with more updates than this are flagged
	CascadeDepthLimit      int     // Dependency chains deeper than this are flagged
	FrequentTriggerCount   int     // Causes firing more than this are flagged
	HighConfidence         float64 // Confidence above this is "high"
}

// DefaultThresholds returns sensible defaults
func DefaultThresholds() Thresholds {
	return Thresholds{
		ExcessiveRerenderCount: 10,
		CascadeDepthLimit:      4,
		FrequentTriggerCount:   15,
		HighConfidence:         0.7,
	}
}

// NewDetector creates a detector with default thresholds
func NewDetector() *Detector {
	return &Detector{thresholds: DefaultThresholds()}
}

// NewDetectorWithThresholds creates a detector with custom thresholds
func NewDetectorWithThresholds(t Thresholds) *Detector {
	return &Detector{thresholds: t}
}

// Detect analyzes a graph and returns all detected issues
func (d *Detector) Detect(g *graph.Graph) []Issue {
	var issues []Issue
	issueID := 0

	nextID := func() string {
		issueID++
		return fmt.Sprintf("issue-%d", issueID)
	}

	// Detect excessive re-renders
	issues = append(issues, d.detectExcessiveRerenders(g, nextID)...)

	// Detect cascading updates
	issues = append(issues, d.detectCascadingUpdates(g, nextID)...)

	// Detect frequent triggers
	issues = append(issues, d.detectFrequentTriggers(g, nextID)...)

	// Detect deep dependency chains
	issues = append(issues, d.detectDeepChains(g, nextID)...)

	// Detect timer cascades
	issues = append(issues, d.detectTimerCascades(g, nextID)...)

	// Detect potential whole-object passing
	issues = append(issues, d.detectWholeObjectPassing(g, nextID)...)

	// Sort by severity
	sort.Slice(issues, func(i, j int) bool {
		return severityRank(issues[i].Severity) > severityRank(issues[j].Severity)
	})

	return issues
}

func severityRank(s Severity) int {
	switch s {
	case SeverityCritical:
		return 5
	case SeverityHigh:
		return 4
	case SeverityMedium:
		return 3
	case SeverityLow:
		return 2
	case SeverityInfo:
		return 1
	}
	return 0
}

func (d *Detector) detectExcessiveRerenders(g *graph.Graph, nextID func() string) []Issue {
	var issues []Issue

	for _, node := range g.Nodes {
		if node.Type != graph.NodeView {
			continue
		}
		if node.Count < d.thresholds.ExcessiveRerenderCount {
			continue
		}

		severity := SeverityMedium
		if node.Count > d.thresholds.ExcessiveRerenderCount*3 {
			severity = SeverityCritical
		} else if node.Count > d.thresholds.ExcessiveRerenderCount*2 {
			severity = SeverityHigh
		}

		issues = append(issues, Issue{
			ID:       nextID(),
			Type:     IssueExcessiveRerender,
			Severity: severity,
			Title:    fmt.Sprintf("Excessive re-renders in %s", node.Label),
			Description: fmt.Sprintf(
				"View '%s' updated %d times during the trace. This suggests the view's dependencies are changing more frequently than necessary.",
				node.Label, node.Count,
			),
			Impact:        "High CPU usage, potential frame drops, battery drain",
			AffectedNodes: []string{node.ID},
			UpdateCount:   node.Count,
			Confidence:    0.85,
			PerformanceHint: fmt.Sprintf(
				"Consider using EquatableView, extracting subviews, or checking if @ObservedObject can be replaced with more granular @State",
			),
		})
	}

	return issues
}

func (d *Detector) detectCascadingUpdates(g *graph.Graph, nextID func() string) []Issue {
	var issues []Issue

	// Find state nodes that trigger multiple views
	for _, node := range g.Nodes {
		if node.Type != graph.NodeState {
			continue
		}

		// Count outgoing edges to views
		viewsAffected := 0
		var affectedViews []string
		for _, edge := range g.Edges {
			if edge.From != node.ID {
				continue
			}
			if targetNode, ok := g.Nodes[edge.To]; ok && targetNode.Type == graph.NodeView {
				viewsAffected++
				affectedViews = append(affectedViews, targetNode.Label)
			}
		}

		if viewsAffected >= 3 {
			severity := SeverityMedium
			if viewsAffected >= 6 {
				severity = SeverityHigh
			}

			issues = append(issues, Issue{
				ID:       nextID(),
				Type:     IssueCascadingUpdate,
				Severity: severity,
				Title:    fmt.Sprintf("State change cascades to %d views", viewsAffected),
				Description: fmt.Sprintf(
					"State '%s' triggers updates in %d different views: %s. Consider whether all views need to observe this entire state.",
					node.Label, viewsAffected, strings.Join(affectedViews, ", "),
				),
				Impact:        "Multiple views re-rendering simultaneously causes frame drops",
				AffectedNodes: append([]string{node.ID}, affectedViews...),
				CascadeDepth:  viewsAffected,
				Confidence:    0.75,
				PerformanceHint: "Split state into smaller pieces, use derived state, or pass only required properties to child views",
			})
		}
	}

	return issues
}

func (d *Detector) detectFrequentTriggers(g *graph.Graph, nextID func() string) []Issue {
	var issues []Issue

	for _, node := range g.Nodes {
		if node.Type != graph.NodeCause {
			continue
		}
		if node.Count < d.thresholds.FrequentTriggerCount {
			continue
		}

		severity := SeverityMedium
		if node.Count > d.thresholds.FrequentTriggerCount*3 {
			severity = SeverityHigh
		}

		issues = append(issues, Issue{
			ID:       nextID(),
			Type:     IssueFrequentTrigger,
			Severity: severity,
			Title:    fmt.Sprintf("Frequent trigger: %s (%d times)", node.Label, node.Count),
			Description: fmt.Sprintf(
				"Cause '%s' fired %d times. If this triggers state updates, it may cause excessive view re-renders.",
				node.Label, node.Count,
			),
			Impact:        "Potential performance bottleneck if each trigger causes view updates",
			AffectedNodes: []string{node.ID},
			UpdateCount:   node.Count,
			Confidence:    0.7,
			PerformanceHint: "Consider debouncing, throttling, or batching updates from this trigger",
		})
	}

	return issues
}

func (d *Detector) detectDeepChains(g *graph.Graph, nextID func() string) []Issue {
	var issues []Issue

	// Find longest path from any cause to any view
	for _, startNode := range g.Nodes {
		if startNode.Type != graph.NodeCause {
			continue
		}

		chain := d.findLongestChain(g, startNode.ID, make(map[string]bool))
		if len(chain) > d.thresholds.CascadeDepthLimit {
			severity := SeverityMedium
			if len(chain) > d.thresholds.CascadeDepthLimit*2 {
				severity = SeverityHigh
			}

			chainLabels := make([]string, len(chain))
			for i, id := range chain {
				if n, ok := g.Nodes[id]; ok {
					chainLabels[i] = n.Label
				} else {
					chainLabels[i] = id
				}
			}

			issues = append(issues, Issue{
				ID:       nextID(),
				Type:     IssueDeepDependencyChain,
				Severity: severity,
				Title:    fmt.Sprintf("Deep dependency chain (%d levels)", len(chain)),
				Description: fmt.Sprintf(
					"Update chain has %d levels: %s. Deep chains increase latency and make debugging harder.",
					len(chain), strings.Join(chainLabels, " → "),
				),
				Impact:        "Increased latency, harder to trace bugs, potential for unnecessary updates",
				AffectedNodes: chain,
				CauseChain:    chainLabels,
				CascadeDepth:  len(chain),
				Confidence:    0.8,
				PerformanceHint: "Consider flattening the dependency tree or using derived state to reduce chain depth",
			})
		}
	}

	return issues
}

func (d *Detector) findLongestChain(g *graph.Graph, nodeID string, visited map[string]bool) []string {
	if visited[nodeID] {
		return nil
	}
	visited[nodeID] = true
	defer func() { visited[nodeID] = false }()

	longest := []string{nodeID}
	for _, edge := range g.Edges {
		if edge.From != nodeID {
			continue
		}
		subChain := d.findLongestChain(g, edge.To, visited)
		if len(subChain)+1 > len(longest) {
			longest = append([]string{nodeID}, subChain...)
		}
	}
	return longest
}

func (d *Detector) detectTimerCascades(g *graph.Graph, nextID func() string) []Issue {
	var issues []Issue

	for _, node := range g.Nodes {
		if node.Type != graph.NodeCause {
			continue
		}
		label := strings.ToLower(node.Label)
		if !strings.Contains(label, "timer") && !strings.Contains(label, "interval") {
			continue
		}

		// Find all views affected by this timer
		affected := d.findReachableViews(g, node.ID)
		if len(affected) >= 2 {
			issues = append(issues, Issue{
				ID:       nextID(),
				Type:     IssueTimerCascade,
				Severity: SeverityHigh,
				Title:    fmt.Sprintf("Timer triggers %d view updates", len(affected)),
				Description: fmt.Sprintf(
					"Timer '%s' causes updates to %d views. Timers that trigger broad UI updates can cause consistent frame drops.",
					node.Label, len(affected),
				),
				Impact:        "Consistent frame drops at timer interval, battery drain",
				AffectedNodes: append([]string{node.ID}, affected...),
				CauseChain:    []string{node.Label},
				Confidence:    0.9,
				PerformanceHint: "Use TimelineView for animations, limit timer scope, or update only changed data",
			})
		}
	}

	return issues
}

func (d *Detector) detectWholeObjectPassing(g *graph.Graph, nextID func() string) []Issue {
	var issues []Issue

	// Heuristic: if a state node has a generic name and affects many views
	for _, node := range g.Nodes {
		if node.Type != graph.NodeState {
			continue
		}

		label := strings.ToLower(node.Label)
		isGeneric := strings.Contains(label, "model") ||
			strings.Contains(label, "viewmodel") ||
			strings.Contains(label, "store") ||
			strings.Contains(label, "state") ||
			strings.Contains(label, "object")

		if !isGeneric {
			continue
		}

		affected := d.countAffectedViews(g, node.ID)
		if affected >= 3 {
			issues = append(issues, Issue{
				ID:       nextID(),
				Type:     IssueWholeObjectPassing,
				Severity: SeverityMedium,
				Title:    fmt.Sprintf("Possible whole-object observation: %s", node.Label),
				Description: fmt.Sprintf(
					"'%s' appears to be a model/state object affecting %d views. Views may be re-rendering when only part of the object changes.",
					node.Label, affected,
				),
				Impact:        "Unnecessary re-renders when unrelated properties change",
				AffectedNodes: []string{node.ID},
				Confidence:    0.6, // Lower confidence - this is heuristic
				PerformanceHint: "Use @Observable with fine-grained properties, or pass only required data to child views",
			})
		}
	}

	return issues
}

func (d *Detector) findReachableViews(g *graph.Graph, startID string) []string {
	visited := make(map[string]bool)
	var views []string

	var dfs func(id string)
	dfs = func(id string) {
		if visited[id] {
			return
		}
		visited[id] = true

		if node, ok := g.Nodes[id]; ok && node.Type == graph.NodeView {
			views = append(views, id)
		}

		for _, edge := range g.Edges {
			if edge.From == id {
				dfs(edge.To)
			}
		}
	}

	dfs(startID)
	return views
}

func (d *Detector) countAffectedViews(g *graph.Graph, nodeID string) int {
	count := 0
	for _, edge := range g.Edges {
		if edge.From != nodeID {
			continue
		}
		if targetNode, ok := g.Nodes[edge.To]; ok && targetNode.Type == graph.NodeView {
			count++
		}
	}
	return count
}
