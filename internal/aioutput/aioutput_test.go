package aioutput

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/greenstevester/swiftui-cause-effect-cli/internal/graph"
	"github.com/greenstevester/swiftui-cause-effect-cli/internal/issues"
)

func TestNewGenerator(t *testing.T) {
	// Test with empty source root
	g, err := NewGenerator("")
	if err != nil {
		t.Fatalf("NewGenerator with empty source root failed: %v", err)
	}
	if g == nil {
		t.Error("Expected non-nil generator")
	}
}

func TestNewGeneratorWithSourceRoot(t *testing.T) {
	tmpDir := t.TempDir()
	swiftFile := filepath.Join(tmpDir, "Test.swift")
	os.WriteFile(swiftFile, []byte("struct Test {}"), 0o644)

	g, err := NewGenerator(tmpDir)
	if err != nil {
		t.Fatalf("NewGenerator with source root failed: %v", err)
	}
	if g == nil {
		t.Error("Expected non-nil generator")
	}
	if g.correlator == nil {
		t.Error("Expected correlator to be initialized")
	}
}

func TestGenerate(t *testing.T) {
	gen, _ := NewGenerator("")

	gr := graph.New()
	gr.UpsertNode(&graph.Node{ID: "c1", Label: "Button tap", Type: graph.NodeCause})
	gr.UpsertNode(&graph.Node{ID: "s1", Label: "@State counter", Type: graph.NodeState})
	gr.UpsertNode(&graph.Node{ID: "v1", Label: "ContentView", Type: graph.NodeView, Count: 50})
	gr.AddEdge(graph.Edge{From: "c1", To: "s1"})
	gr.AddEdge(graph.Edge{From: "s1", To: "v1"})

	report := gen.Generate(gr, GenerateOptions{
		TracePath:   "test.trace",
		ExportDir:   "exported",
		FilesParsed: 1,
	})

	if report.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", report.Version)
	}

	if report.Tool != "swiftuice" {
		t.Errorf("Expected tool swiftuice, got %s", report.Tool)
	}

	if report.Summary.TotalCauses != 1 {
		t.Errorf("Expected 1 cause, got %d", report.Summary.TotalCauses)
	}

	if report.Summary.TotalStateChanges != 1 {
		t.Errorf("Expected 1 state change, got %d", report.Summary.TotalStateChanges)
	}

	if report.Summary.TotalViewUpdates != 1 {
		t.Errorf("Expected 1 view update, got %d", report.Summary.TotalViewUpdates)
	}

	if len(report.Graph.Nodes) != 3 {
		t.Errorf("Expected 3 graph nodes, got %d", len(report.Graph.Nodes))
	}

	if len(report.Graph.Edges) != 2 {
		t.Errorf("Expected 2 graph edges, got %d", len(report.Graph.Edges))
	}
}

func TestGenerateWithIssues(t *testing.T) {
	gen, _ := NewGenerator("")

	// Create a graph that triggers issues
	gr := graph.New()
	gr.UpsertNode(&graph.Node{ID: "v1", Label: "ItemRow", Type: graph.NodeView, Count: 50})
	gr.UpsertNode(&graph.Node{ID: "s1", Label: "@State", Type: graph.NodeState})
	gr.AddEdge(graph.Edge{From: "s1", To: "v1"})

	report := gen.Generate(gr, GenerateOptions{})

	// Should detect excessive rerender issue
	hasIssue := false
	for _, issue := range report.Issues {
		if issue.Type == issues.IssueExcessiveRerender {
			hasIssue = true
			if len(issue.SuggestedFixes) == 0 {
				t.Error("Expected suggested fixes for excessive rerender issue")
			}
		}
	}

	if !hasIssue {
		t.Error("Expected to detect excessive rerender issue")
	}

	// Performance score should reflect issues
	if report.Summary.PerformanceScore >= 100 {
		t.Error("Expected performance score to be reduced due to issues")
	}
}

func TestCalculateSummary(t *testing.T) {
	gen, _ := NewGenerator("")

	gr := graph.New()
	gr.UpsertNode(&graph.Node{ID: "c1", Type: graph.NodeCause})
	gr.UpsertNode(&graph.Node{ID: "c2", Type: graph.NodeCause})
	gr.UpsertNode(&graph.Node{ID: "s1", Type: graph.NodeState})
	gr.UpsertNode(&graph.Node{ID: "v1", Type: graph.NodeView})
	gr.UpsertNode(&graph.Node{ID: "v2", Type: graph.NodeView})
	gr.UpsertNode(&graph.Node{ID: "v3", Type: graph.NodeView})
	gr.AddEdge(graph.Edge{From: "c1", To: "s1"})
	gr.AddEdge(graph.Edge{From: "s1", To: "v1"})

	detected := []issues.Issue{}
	summary := gen.calculateSummary(gr, detected)

	if summary.TotalCauses != 2 {
		t.Errorf("Expected 2 causes, got %d", summary.TotalCauses)
	}

	if summary.TotalStateChanges != 1 {
		t.Errorf("Expected 1 state change, got %d", summary.TotalStateChanges)
	}

	if summary.TotalViewUpdates != 3 {
		t.Errorf("Expected 3 view updates, got %d", summary.TotalViewUpdates)
	}

	if summary.TotalEdges != 2 {
		t.Errorf("Expected 2 edges, got %d", summary.TotalEdges)
	}

	// No issues = perfect score
	if summary.PerformanceScore != 100 {
		t.Errorf("Expected score 100 with no issues, got %d", summary.PerformanceScore)
	}

	if summary.HealthStatus != "good" {
		t.Errorf("Expected health status 'good', got %s", summary.HealthStatus)
	}
}

func TestPerformanceScoreWithIssues(t *testing.T) {
	gen, _ := NewGenerator("")
	gr := graph.New()

	// Score thresholds: >= 75 = good, >= 50 = warning, < 50 = critical
	// Critical issue = -25 points, High issue = -10 points, Other = -3 points
	tests := []struct {
		issues         []issues.Issue
		expectedScore  int
		expectedStatus string
	}{
		{
			issues:         []issues.Issue{},
			expectedScore:  100,
			expectedStatus: "good",
		},
		{
			issues: []issues.Issue{
				{Severity: issues.SeverityCritical},
			},
			expectedScore:  75, // 100 - 25 = 75 >= 75 → good
			expectedStatus: "good",
		},
		{
			issues: []issues.Issue{
				{Severity: issues.SeverityCritical},
				{Severity: issues.SeverityCritical},
			},
			expectedScore:  50, // 100 - 50 = 50 < 75 → warning
			expectedStatus: "warning",
		},
		{
			issues: []issues.Issue{
				{Severity: issues.SeverityCritical},
				{Severity: issues.SeverityCritical},
				{Severity: issues.SeverityCritical},
			},
			expectedScore:  25, // 100 - 75 = 25 < 50 → critical
			expectedStatus: "critical",
		},
		{
			issues: []issues.Issue{
				{Severity: issues.SeverityHigh},
				{Severity: issues.SeverityHigh},
			},
			expectedScore:  80, // 100 - 20 = 80 >= 75 → good
			expectedStatus: "good",
		},
	}

	for _, tt := range tests {
		summary := gen.calculateSummary(gr, tt.issues)
		if summary.PerformanceScore != tt.expectedScore {
			t.Errorf("With %d issues: expected score %d, got %d", len(tt.issues), tt.expectedScore, summary.PerformanceScore)
		}
		if summary.HealthStatus != tt.expectedStatus {
			t.Errorf("With %d issues: expected status %s, got %s", len(tt.issues), tt.expectedStatus, summary.HealthStatus)
		}
	}
}

func TestAgentInstructions(t *testing.T) {
	gen, _ := NewGenerator("")

	gr := graph.New()
	gr.UpsertNode(&graph.Node{ID: "v1", Label: "ItemRow", Type: graph.NodeView, Count: 50})
	gr.UpsertNode(&graph.Node{ID: "s1", Label: "@State", Type: graph.NodeState})
	gr.AddEdge(graph.Edge{From: "s1", To: "v1"})

	report := gen.Generate(gr, GenerateOptions{})

	if report.AgentInstructions.TaskDescription == "" {
		t.Error("Expected non-empty task description")
	}

	if len(report.AgentInstructions.Constraints) == 0 {
		t.Error("Expected constraints")
	}

	if len(report.AgentInstructions.SuccessCriteria) == 0 {
		t.Error("Expected success criteria")
	}

	if report.AgentInstructions.Context == "" {
		t.Error("Expected non-empty context")
	}
}

func TestToJSON(t *testing.T) {
	gen, _ := NewGenerator("")
	gr := graph.New()
	gr.UpsertNode(&graph.Node{ID: "v1", Label: "Test", Type: graph.NodeView})

	report := gen.Generate(gr, GenerateOptions{})

	jsonStr, err := report.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed Report
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("ToJSON produced invalid JSON: %v", err)
	}

	if parsed.Version != "1.0" {
		t.Errorf("Parsed version mismatch: got %s, expected 1.0", parsed.Version)
	}
}

func TestToCompactJSON(t *testing.T) {
	gen, _ := NewGenerator("")
	gr := graph.New()
	gr.UpsertNode(&graph.Node{ID: "v1", Label: "Test", Type: graph.NodeView})

	report := gen.Generate(gr, GenerateOptions{})

	compactJSON, err := report.ToCompactJSON()
	if err != nil {
		t.Fatalf("ToCompactJSON failed: %v", err)
	}

	prettyJSON, _ := report.ToJSON()

	// Compact should be shorter (no indentation)
	if len(compactJSON) >= len(prettyJSON) {
		t.Error("Compact JSON should be shorter than pretty JSON")
	}

	// Verify it's valid JSON
	var parsed Report
	if err := json.Unmarshal([]byte(compactJSON), &parsed); err != nil {
		t.Fatalf("ToCompactJSON produced invalid JSON: %v", err)
	}
}

func TestWriteJSON(t *testing.T) {
	gen, _ := NewGenerator("")
	gr := graph.New()
	gr.UpsertNode(&graph.Node{ID: "v1", Label: "Test", Type: graph.NodeView})

	report := gen.Generate(gr, GenerateOptions{})

	tmpFile := filepath.Join(t.TempDir(), "report.json")
	if err := report.WriteJSON(tmpFile); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	var parsed Report
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Written file is not valid JSON: %v", err)
	}
}

func TestGraphDataStructure(t *testing.T) {
	gen, _ := NewGenerator("")

	gr := graph.New()
	gr.UpsertNode(&graph.Node{ID: "c1", Label: "Cause", Type: graph.NodeCause})
	gr.UpsertNode(&graph.Node{ID: "s1", Label: "State", Type: graph.NodeState})
	gr.UpsertNode(&graph.Node{ID: "v1", Label: "View", Type: graph.NodeView, Count: 10})
	gr.AddEdge(graph.Edge{From: "c1", To: "s1", Label: "triggers"})
	gr.AddEdge(graph.Edge{From: "s1", To: "v1", Label: "updates"})

	report := gen.Generate(gr, GenerateOptions{})

	// Verify node data
	nodeMap := make(map[string]NodeData)
	for _, n := range report.Graph.Nodes {
		nodeMap[n.ID] = n
	}

	if nodeMap["c1"].Type != "cause" {
		t.Errorf("Expected cause type, got %s", nodeMap["c1"].Type)
	}

	if nodeMap["v1"].UpdateCount != 10 {
		t.Errorf("Expected update count 10, got %d", nodeMap["v1"].UpdateCount)
	}

	// Verify edge data
	if len(report.Graph.Edges) != 2 {
		t.Errorf("Expected 2 edges, got %d", len(report.Graph.Edges))
	}

	hasTriggersEdge := false
	for _, e := range report.Graph.Edges {
		if e.Label == "triggers" && e.From == "c1" && e.To == "s1" {
			hasTriggersEdge = true
		}
	}
	if !hasTriggersEdge {
		t.Error("Expected triggers edge from c1 to s1")
	}
}

func TestRecommendationsGenerated(t *testing.T) {
	gen, _ := NewGenerator("")

	// Create graph that triggers issues
	gr := graph.New()
	gr.UpsertNode(&graph.Node{ID: "v1", Label: "ItemRow", Type: graph.NodeView, Count: 50})
	gr.UpsertNode(&graph.Node{ID: "s1", Label: "@State", Type: graph.NodeState})
	gr.AddEdge(graph.Edge{From: "s1", To: "v1"})

	report := gen.Generate(gr, GenerateOptions{})

	// Should have recommendations based on detected issues
	if len(report.Recommendations) == 0 {
		t.Error("Expected recommendations for issues")
	}
}

func TestInputInfoCaptured(t *testing.T) {
	gen, _ := NewGenerator("")
	gr := graph.New()
	gr.UpsertNode(&graph.Node{ID: "v1", Label: "Test", Type: graph.NodeView})

	report := gen.Generate(gr, GenerateOptions{
		TracePath:   "/path/to/trace.trace",
		ExportDir:   "/path/to/exported",
		SourceRoot:  "/path/to/source",
		FilesParsed: 5,
	})

	if report.Input.TracePath != "/path/to/trace.trace" {
		t.Errorf("Expected trace path to be captured")
	}

	if report.Input.ExportDir != "/path/to/exported" {
		t.Errorf("Expected export dir to be captured")
	}

	if report.Input.SourceRoot != "/path/to/source" {
		t.Errorf("Expected source root to be captured")
	}

	if report.Input.FilesParsed != 5 {
		t.Errorf("Expected files parsed to be 5, got %d", report.Input.FilesParsed)
	}
}
