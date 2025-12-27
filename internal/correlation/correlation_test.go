package correlation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/greenstevester/swiftui-cause-effect-cli/internal/graph"
)

func TestExtractSymbols(t *testing.T) {
	tests := []struct {
		label    string
		expected []string
	}{
		{
			label:    "ItemRow",
			expected: []string{"ItemRow"},
		},
		{
			label:    "ContentView updated",
			expected: []string{"ContentView"},
		},
		{
			label:    "@State var counter",
			expected: []string{"State", "counter"},
		},
	}

	for _, tt := range tests {
		symbols := extractSymbols(tt.label)
		for _, exp := range tt.expected {
			found := false
			for _, sym := range symbols {
				if sym == exp {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("extractSymbols(%q): expected to contain %q, got %v", tt.label, exp, symbols)
			}
		}
	}
}

func TestIsCommonWord(t *testing.T) {
	tests := []struct {
		word     string
		expected bool
	}{
		{"the", true},
		{"var", true},
		{"view", true},
		{"ItemRow", false},
		{"counter", false},
		{"MyView", false},
	}

	for _, tt := range tests {
		result := isCommonWord(tt.word)
		if result != tt.expected {
			t.Errorf("isCommonWord(%q): got %v, expected %v", tt.word, result, tt.expected)
		}
	}
}

func TestDedupe(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b", "d"}
	result := dedupe(input)
	expected := []string{"a", "b", "c", "d"}

	if len(result) != len(expected) {
		t.Errorf("dedupe: got %d items, expected %d", len(result), len(expected))
	}

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("dedupe[%d]: got %q, expected %q", i, result[i], exp)
		}
	}
}

func TestMatchViewDeclaration(t *testing.T) {
	tests := []struct {
		line       string
		symbol     string
		wantMatch  bool
		minConf    float64
	}{
		{"struct ContentView: View {", "ContentView", true, 0.9},
		{"struct ItemRow: View, Equatable {", "ItemRow", true, 0.9},
		{"struct MyView: SomeProtocol, View {", "MyView", true, 0.9},
		{"var contentView = ContentView()", "ContentView", false, 0.0},
		{"struct HelperView {", "HelperView", true, 0.8}, // Has "View" in name
	}

	for _, tt := range tests {
		matched, conf := matchViewDeclaration(tt.line, tt.symbol)
		if matched != tt.wantMatch {
			t.Errorf("matchViewDeclaration(%q, %q): got match=%v, expected %v", tt.line, tt.symbol, matched, tt.wantMatch)
		}
		if matched && conf < tt.minConf {
			t.Errorf("matchViewDeclaration(%q, %q): got conf=%v, expected >= %v", tt.line, tt.symbol, conf, tt.minConf)
		}
	}
}

func TestMatchStateDeclaration(t *testing.T) {
	tests := []struct {
		line      string
		symbol    string
		wantMatch bool
		minConf   float64
	}{
		{"@State var counter: Int = 0", "counter", true, 0.9},
		{"@State private var isEnabled = false", "isEnabled", true, 0.9},
		{"@StateObject var viewModel = ViewModel()", "viewModel", true, 0.9},
		{"@ObservedObject var model: Model", "model", true, 0.9},
		{"@Binding var value: String", "value", true, 0.8},
		{"var counter = 0", "counter", false, 0.0},
	}

	for _, tt := range tests {
		matched, conf := matchStateDeclaration(tt.line, tt.symbol)
		if matched != tt.wantMatch {
			t.Errorf("matchStateDeclaration(%q, %q): got match=%v, expected %v", tt.line, tt.symbol, matched, tt.wantMatch)
		}
		if matched && conf < tt.minConf {
			t.Errorf("matchStateDeclaration(%q, %q): got conf=%v, expected >= %v", tt.line, tt.symbol, conf, tt.minConf)
		}
	}
}

func TestMatchCausePattern(t *testing.T) {
	tests := []struct {
		line      string
		symbol    string
		wantMatch bool
	}{
		{"Button(\"Tap me\") { doSomething() }", "Tap", true},
		{".onTapGesture { handleTap() }", "tap", true},
		{"Timer.scheduledTimer(withTimeInterval: 1.0)", "timer", true},
		{"NotificationCenter.default.post", "notification", true},
		{"let x = 5", "x", false},
	}

	for _, tt := range tests {
		matched, _ := matchCausePattern(tt.line, tt.symbol)
		if matched != tt.wantMatch {
			t.Errorf("matchCausePattern(%q, %q): got match=%v, expected %v", tt.line, tt.symbol, matched, tt.wantMatch)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		max      int
		expected string
	}{
		{"short", 10, "short"},
		{"a very long string that exceeds the limit", 20, "a very long strin..."},
		{"exact", 5, "exact"},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.max)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d): got %q, expected %q", tt.input, tt.max, result, tt.expected)
		}
	}
}

func TestNewCorrelator(t *testing.T) {
	// Create a temp directory with a Swift file
	tmpDir := t.TempDir()
	swiftFile := filepath.Join(tmpDir, "ContentView.swift")
	err := os.WriteFile(swiftFile, []byte("struct ContentView: View {\n    var body: some View {\n        Text(\"Hello\")\n    }\n}"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewCorrelator(tmpDir)
	if err != nil {
		t.Fatalf("NewCorrelator failed: %v", err)
	}

	if c.SwiftFileCount() != 1 {
		t.Errorf("Expected 1 Swift file, got %d", c.SwiftFileCount())
	}

	if c.GetSourceRoot() != tmpDir {
		t.Errorf("GetSourceRoot: got %q, expected %q", c.GetSourceRoot(), tmpDir)
	}
}

func TestCorrelate(t *testing.T) {
	// Create a temp directory with a Swift file
	tmpDir := t.TempDir()
	swiftFile := filepath.Join(tmpDir, "ContentView.swift")
	content := `struct ContentView: View {
    @State var counter: Int = 0

    var body: some View {
        Button("Tap") { counter += 1 }
    }
}`
	err := os.WriteFile(swiftFile, []byte(content), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewCorrelator(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	g := graph.New()
	g.UpsertNode(&graph.Node{ID: "v1", Label: "ContentView", Type: graph.NodeView})
	g.UpsertNode(&graph.Node{ID: "s1", Label: "counter", Type: graph.NodeState})
	g.UpsertNode(&graph.Node{ID: "c1", Label: "Button tap", Type: graph.NodeCause})

	matches := c.Correlate(g)

	// Should find at least some matches
	if len(matches) == 0 {
		t.Error("Expected at least one source match")
	}

	// Check that ContentView was matched
	hasViewMatch := false
	for _, m := range matches {
		if m.TraceNodeID == "v1" && m.Confidence >= 0.9 {
			hasViewMatch = true
		}
	}
	if !hasViewMatch {
		t.Error("Expected high-confidence match for ContentView")
	}
}

func TestBestMatch(t *testing.T) {
	tmpDir := t.TempDir()
	swiftFile := filepath.Join(tmpDir, "Test.swift")
	err := os.WriteFile(swiftFile, []byte("struct TestView: View { }"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewCorrelator(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	g := graph.New()
	g.UpsertNode(&graph.Node{ID: "v1", Label: "TestView", Type: graph.NodeView})

	// Correlate first to populate cache
	c.Correlate(g)

	// Now check BestMatch
	best := c.BestMatch("v1")
	if best == nil {
		t.Error("Expected BestMatch to return a match")
	}

	// Check for non-existent node
	noMatch := c.BestMatch("nonexistent")
	if noMatch != nil {
		t.Error("Expected nil for non-existent node")
	}
}

func TestCorrelatorSkipsCommonDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	mainFile := filepath.Join(tmpDir, "Main.swift")
	os.WriteFile(mainFile, []byte("struct Main {}"), 0o644)

	// Create files in directories that should be skipped
	gitDir := filepath.Join(tmpDir, ".git")
	os.MkdirAll(gitDir, 0o755)
	os.WriteFile(filepath.Join(gitDir, "config.swift"), []byte("// git"), 0o644)

	podsDir := filepath.Join(tmpDir, "Pods")
	os.MkdirAll(podsDir, 0o755)
	os.WriteFile(filepath.Join(podsDir, "Pod.swift"), []byte("// pod"), 0o644)

	c, err := NewCorrelator(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Should only find the Main.swift file, not the ones in .git or Pods
	if c.SwiftFileCount() != 1 {
		t.Errorf("Expected 1 Swift file (skipping .git and Pods), got %d", c.SwiftFileCount())
	}
}
