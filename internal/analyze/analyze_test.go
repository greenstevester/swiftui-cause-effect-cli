package analyze

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/greenstevester/swiftui-cause-effect-cli/internal/graph"
)

func TestAsString(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		def      string
		expected string
	}{
		{"nil returns default", nil, "default", "default"},
		{"string returns value", "hello", "default", "hello"},
		{"empty string returns empty", "", "default", ""},
		{"int returns default", 42, "default", "default"},
		{"bool returns default", true, "default", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := asString(tt.input, tt.def)
			if got != tt.expected {
				t.Errorf("asString(%v, %q) = %q, want %q", tt.input, tt.def, got, tt.expected)
			}
		})
	}
}

func TestAsInt(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		def      int
		expected int
	}{
		{"nil returns default", nil, 99, 99},
		{"float64 returns int", float64(42.7), 0, 42},
		{"int returns value", 123, 0, 123},
		{"int64 returns value", int64(456), 0, 456},
		{"string returns default", "not a number", 99, 99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := asInt(tt.input, tt.def)
			if got != tt.expected {
				t.Errorf("asInt(%v, %d) = %d, want %d", tt.input, tt.def, got, tt.expected)
			}
		})
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		kind     string
		label    string
		expected graph.NodeType
	}{
		{"state", "", graph.NodeState},
		{"State", "", graph.NodeState},
		{"", "@State var count", graph.NodeState},
		{"", "@ObservedObject", graph.NodeState},
		{"view", "", graph.NodeView},
		{"View", "", graph.NodeView},
		{"", "body() called", graph.NodeView},
		{"", "View update triggered", graph.NodeView}, // matches "view update" pattern
		{"cause", "", graph.NodeCause},
		{"", "button tap", graph.NodeCause},
		{"", "gesture recognized", graph.NodeCause},
		{"", "timer fired", graph.NodeCause},
		{"unknown", "random text", graph.NodeOther},
		{"", "", graph.NodeOther},
	}

	for _, tt := range tests {
		t.Run(tt.kind+"/"+tt.label, func(t *testing.T) {
			got := classify(tt.kind, tt.label)
			if got != tt.expected {
				t.Errorf("classify(%q, %q) = %s, want %s", tt.kind, tt.label, got, tt.expected)
			}
		})
	}
}

func TestTrim(t *testing.T) {
	tests := []struct {
		input    string
		max      int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is too long", 10, "this is t…"},
		{"", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := trim(tt.input, tt.max)
			if got != tt.expected {
				t.Errorf("trim(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.expected)
			}
		})
	}
}

func TestEscapeDOT(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{`has "quotes"`, `has \"quotes\"`},
		{"has\\backslash", "has\\\\backslash"},
		{"has\nnewline", "has newline"},
		{"has\rcarriage", "has carriage"},
		{"mixed \"quote\" and\nstuff", `mixed \"quote\" and stuff`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := escapeDOT(tt.input)
			if got != tt.expected {
				t.Errorf("escapeDOT(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIdOrHash(t *testing.T) {
	// When ID is provided, use it
	got := idOrHash("myid", "some label")
	if got != "myid" {
		t.Errorf("idOrHash with ID: expected 'myid', got %q", got)
	}

	// When ID is empty, generate from label
	got1 := idOrHash("", "test label")
	got2 := idOrHash("", "test label")
	if got1 != got2 {
		t.Errorf("idOrHash should be deterministic: %q != %q", got1, got2)
	}
	if !strings.HasPrefix(got1, "n") {
		t.Errorf("generated ID should start with 'n': %q", got1)
	}

	// Different labels should produce different IDs
	got3 := idOrHash("", "different label")
	if got1 == got3 {
		t.Errorf("different labels should produce different IDs: %q == %q", got1, got3)
	}
}

func TestParseTextReader_StateDetection(t *testing.T) {
	input := `
some random line
@State var counter = 0
another line
@ObservedObject var model
`
	g := graph.New()
	stats := &summaryStats{}

	err := parseTextReader(strings.NewReader(input), g, stats)
	if err != nil {
		t.Fatalf("parseTextReader failed: %v", err)
	}

	// Should have detected 2 state nodes
	stateCount := 0
	for _, n := range g.Nodes {
		if n.Type == graph.NodeState {
			stateCount++
		}
	}
	if stateCount != 2 {
		t.Errorf("expected 2 state nodes, got %d", stateCount)
	}
}

func TestParseTextReader_ViewDetection(t *testing.T) {
	input := `
ContentView body() called
View update triggered
`
	g := graph.New()
	stats := &summaryStats{}

	err := parseTextReader(strings.NewReader(input), g, stats)
	if err != nil {
		t.Fatalf("parseTextReader failed: %v", err)
	}

	viewCount := 0
	for _, n := range g.Nodes {
		if n.Type == graph.NodeView {
			viewCount++
		}
	}
	if viewCount != 2 {
		t.Errorf("expected 2 view nodes, got %d", viewCount)
	}
}

func TestParseTextReader_CauseDetection(t *testing.T) {
	input := `
button tapped
gesture recognized
timer fired
`
	g := graph.New()
	stats := &summaryStats{}

	err := parseTextReader(strings.NewReader(input), g, stats)
	if err != nil {
		t.Fatalf("parseTextReader failed: %v", err)
	}

	causeCount := 0
	for _, n := range g.Nodes {
		if n.Type == graph.NodeCause {
			causeCount++
		}
	}
	if causeCount != 3 {
		t.Errorf("expected 3 cause nodes, got %d", causeCount)
	}
}

func TestParseTextReader_EdgeCreation(t *testing.T) {
	// Simulate: cause -> state -> view sequence
	input := `
button tap event
@State change detected
View body updated
`
	g := graph.New()
	stats := &summaryStats{}

	err := parseTextReader(strings.NewReader(input), g, stats)
	if err != nil {
		t.Fatalf("parseTextReader failed: %v", err)
	}

	// Should have edges connecting them
	if len(g.Edges) < 2 {
		t.Errorf("expected at least 2 edges, got %d", len(g.Edges))
	}
}

func TestRenderDOT(t *testing.T) {
	g := graph.New()
	g.UpsertNode(&graph.Node{ID: "c1", Label: "Button Tap", Type: graph.NodeCause})
	g.UpsertNode(&graph.Node{ID: "s1", Label: "@State count", Type: graph.NodeState})
	g.UpsertNode(&graph.Node{ID: "v1", Label: "CounterView", Type: graph.NodeView, Count: 5})
	g.AddEdge(graph.Edge{From: "c1", To: "s1", Label: "causes"})
	g.AddEdge(graph.Edge{From: "s1", To: "v1", Label: "updates"})

	dot := renderDOT(g)

	// Check structure
	if !strings.Contains(dot, "digraph CauseEffect") {
		t.Error("missing digraph declaration")
	}
	if !strings.Contains(dot, "rankdir=LR") {
		t.Error("missing rankdir")
	}
	// Check nodes with correct shapes
	if !strings.Contains(dot, `"c1"`) {
		t.Error("missing cause node c1")
	}
	if !strings.Contains(dot, `"s1"`) {
		t.Error("missing state node s1")
	}
	if !strings.Contains(dot, `"v1"`) {
		t.Error("missing view node v1")
	}
	// Check edges
	if !strings.Contains(dot, `"c1" -> "s1"`) {
		t.Error("missing edge c1 -> s1")
	}
	if !strings.Contains(dot, `"s1" -> "v1"`) {
		t.Error("missing edge s1 -> v1")
	}
	// Check count annotation
	if !strings.Contains(dot, "count=5") {
		t.Error("missing count annotation")
	}
}

func TestRenderMarkdown(t *testing.T) {
	g := graph.New()
	g.UpsertNode(&graph.Node{ID: "c1", Label: "Tap", Type: graph.NodeCause})
	g.UpsertNode(&graph.Node{ID: "s1", Label: "State", Type: graph.NodeState})
	g.UpsertNode(&graph.Node{ID: "v1", Label: "View", Type: graph.NodeView, Count: 10})
	g.AddEdge(graph.Edge{From: "c1", To: "s1"})

	stats := &summaryStats{FilesParsed: 3, Hints: []string{"test hint"}}

	md := renderMarkdown(g, stats)

	if !strings.Contains(md, "# SwiftUI Cause & Effect Summary") {
		t.Error("missing title")
	}
	if !strings.Contains(md, "Parsed 3 files") {
		t.Error("missing file count")
	}
	if !strings.Contains(md, "Causes: 1") {
		t.Error("missing cause count")
	}
	if !strings.Contains(md, "State changes: 1") {
		t.Error("missing state count")
	}
	if !strings.Contains(md, "View updates: 1") {
		t.Error("missing view count")
	}
	if !strings.Contains(md, "test hint") {
		t.Error("missing hint")
	}
}

func TestParseJSON_WithNodesAndEdges(t *testing.T) {
	jsonContent := `{
		"nodes": [
			{"id": "n1", "label": "Button", "type": "cause"},
			{"id": "n2", "label": "@State", "type": "state"},
			{"id": "n3", "label": "ContentView", "type": "view", "count": 5}
		],
		"edges": [
			{"from": "n1", "to": "n2", "label": "triggers"},
			{"from": "n2", "to": "n3", "label": "updates"}
		]
	}`

	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "test.json")
	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	g := graph.New()
	stats := &summaryStats{}

	err := parseJSON(jsonPath, g, stats)
	if err != nil {
		t.Fatalf("parseJSON failed: %v", err)
	}

	if len(g.Nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 2 {
		t.Errorf("expected 2 edges, got %d", len(g.Edges))
	}

	// Verify node n3 has count
	if n3, ok := g.Nodes["n3"]; ok {
		if n3.Count != 5 {
			t.Errorf("expected n3 count=5, got %d", n3.Count)
		}
	} else {
		t.Error("node n3 not found")
	}
}

func TestSummarize_NoData(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an empty file that won't produce any nodes
	emptyFile := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(emptyFile, []byte("nothing useful here\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Summarize(Options{
		Input:      tmpDir,
		OutSummary: filepath.Join(tmpDir, "summary.md"),
		OutDOT:     filepath.Join(tmpDir, "graph.dot"),
	})

	if err != ErrNoData {
		t.Errorf("expected ErrNoData, got %v", err)
	}
}

func TestSummarize_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with parseable content
	content := `button tap happened
@State var counter changed
View body() called
`
	if err := os.WriteFile(filepath.Join(tmpDir, "trace.txt"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	summaryPath := filepath.Join(tmpDir, "summary.md")
	dotPath := filepath.Join(tmpDir, "graph.dot")

	result, err := Summarize(Options{
		Input:      tmpDir,
		OutSummary: summaryPath,
		OutDOT:     dotPath,
	})

	if err != nil {
		t.Fatalf("Summarize failed: %v", err)
	}

	if result.SummaryPath != summaryPath {
		t.Errorf("expected SummaryPath %q, got %q", summaryPath, result.SummaryPath)
	}
	if result.DotPath != dotPath {
		t.Errorf("expected DotPath %q, got %q", dotPath, result.DotPath)
	}

	// Verify files were created
	if _, err := os.Stat(summaryPath); os.IsNotExist(err) {
		t.Error("summary.md was not created")
	}
	if _, err := os.Stat(dotPath); os.IsNotExist(err) {
		t.Error("graph.dot was not created")
	}
}
