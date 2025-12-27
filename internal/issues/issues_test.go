package issues

import (
	"testing"

	"github.com/greenstevester/swiftui-cause-effect-cli/internal/graph"
)

func TestNewDetector(t *testing.T) {
	d := NewDetector()
	if d == nil {
		t.Error("NewDetector returned nil")
	}
}

func TestDefaultThresholds(t *testing.T) {
	th := DefaultThresholds()
	if th.ExcessiveRerenderCount <= 0 {
		t.Error("ExcessiveRerenderCount should be positive")
	}
	if th.CascadeDepthLimit <= 0 {
		t.Error("CascadeDepthLimit should be positive")
	}
	if th.FrequentTriggerCount <= 0 {
		t.Error("FrequentTriggerCount should be positive")
	}
}

func TestDetect_ExcessiveRerender(t *testing.T) {
	g := graph.New()
	g.UpsertNode(&graph.Node{ID: "v1", Label: "ItemRow", Type: graph.NodeView, Count: 50})
	g.UpsertNode(&graph.Node{ID: "s1", Label: "@State", Type: graph.NodeState})
	g.AddEdge(graph.Edge{From: "s1", To: "v1"})

	d := NewDetector()
	issues := d.Detect(g)

	found := false
	for _, issue := range issues {
		if issue.Type == IssueExcessiveRerender {
			found = true
			if issue.Severity != SeverityCritical && issue.Severity != SeverityHigh {
				t.Errorf("Expected high/critical severity for 50 updates, got %s", issue.Severity)
			}
		}
	}
	if !found {
		t.Error("Expected to detect excessive rerender issue")
	}
}

func TestDetect_CascadingUpdate(t *testing.T) {
	g := graph.New()
	g.UpsertNode(&graph.Node{ID: "s1", Label: "AppState", Type: graph.NodeState})
	g.UpsertNode(&graph.Node{ID: "v1", Label: "View1", Type: graph.NodeView})
	g.UpsertNode(&graph.Node{ID: "v2", Label: "View2", Type: graph.NodeView})
	g.UpsertNode(&graph.Node{ID: "v3", Label: "View3", Type: graph.NodeView})
	g.UpsertNode(&graph.Node{ID: "v4", Label: "View4", Type: graph.NodeView})
	g.AddEdge(graph.Edge{From: "s1", To: "v1"})
	g.AddEdge(graph.Edge{From: "s1", To: "v2"})
	g.AddEdge(graph.Edge{From: "s1", To: "v3"})
	g.AddEdge(graph.Edge{From: "s1", To: "v4"})

	d := NewDetector()
	issues := d.Detect(g)

	found := false
	for _, issue := range issues {
		if issue.Type == IssueCascadingUpdate {
			found = true
		}
	}
	if !found {
		t.Error("Expected to detect cascading update issue")
	}
}

func TestDetect_TimerCascade(t *testing.T) {
	g := graph.New()
	g.UpsertNode(&graph.Node{ID: "c1", Label: "Timer fired", Type: graph.NodeCause})
	g.UpsertNode(&graph.Node{ID: "s1", Label: "time state", Type: graph.NodeState})
	g.UpsertNode(&graph.Node{ID: "v1", Label: "ClockView", Type: graph.NodeView})
	g.UpsertNode(&graph.Node{ID: "v2", Label: "HeaderView", Type: graph.NodeView})
	g.AddEdge(graph.Edge{From: "c1", To: "s1"})
	g.AddEdge(graph.Edge{From: "s1", To: "v1"})
	g.AddEdge(graph.Edge{From: "s1", To: "v2"})

	d := NewDetector()
	issues := d.Detect(g)

	found := false
	for _, issue := range issues {
		if issue.Type == IssueTimerCascade {
			found = true
		}
	}
	if !found {
		t.Error("Expected to detect timer cascade issue")
	}
}

func TestDetect_NoIssues(t *testing.T) {
	g := graph.New()
	g.UpsertNode(&graph.Node{ID: "c1", Label: "Button tap", Type: graph.NodeCause})
	g.UpsertNode(&graph.Node{ID: "s1", Label: "@State", Type: graph.NodeState})
	g.UpsertNode(&graph.Node{ID: "v1", Label: "ContentView", Type: graph.NodeView, Count: 2})
	g.AddEdge(graph.Edge{From: "c1", To: "s1"})
	g.AddEdge(graph.Edge{From: "s1", To: "v1"})

	d := NewDetector()
	issues := d.Detect(g)

	// Should have no high-severity issues for this simple graph
	for _, issue := range issues {
		if issue.Severity == SeverityCritical || issue.Severity == SeverityHigh {
			t.Errorf("Unexpected high-severity issue: %s", issue.Title)
		}
	}
}

func TestSeverityRank(t *testing.T) {
	if severityRank(SeverityCritical) <= severityRank(SeverityHigh) {
		t.Error("Critical should rank higher than High")
	}
	if severityRank(SeverityHigh) <= severityRank(SeverityMedium) {
		t.Error("High should rank higher than Medium")
	}
	if severityRank(SeverityMedium) <= severityRank(SeverityLow) {
		t.Error("Medium should rank higher than Low")
	}
	if severityRank(SeverityLow) <= severityRank(SeverityInfo) {
		t.Error("Low should rank higher than Info")
	}
}
