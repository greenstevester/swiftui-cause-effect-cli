package graph

import (
	"testing"
)

func TestNewGraph(t *testing.T) {
	g := New()
	if g == nil {
		t.Fatal("New() returned nil")
	}
	if g.Nodes == nil {
		t.Error("Nodes map is nil")
	}
	if g.Edges == nil {
		t.Error("Edges slice is nil")
	}
	if len(g.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 0 {
		t.Errorf("expected 0 edges, got %d", len(g.Edges))
	}
}

func TestUpsertNode_New(t *testing.T) {
	g := New()
	node := &Node{
		ID:    "n1",
		Label: "Test Node",
		Type:  NodeCause,
		Count: 5,
	}
	g.UpsertNode(node)

	if len(g.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(g.Nodes))
	}
	got, ok := g.Nodes["n1"]
	if !ok {
		t.Fatal("node n1 not found")
	}
	if got.Label != "Test Node" {
		t.Errorf("expected label 'Test Node', got '%s'", got.Label)
	}
	if got.Type != NodeCause {
		t.Errorf("expected type NodeCause, got %s", got.Type)
	}
	if got.Count != 5 {
		t.Errorf("expected count 5, got %d", got.Count)
	}
}

func TestUpsertNode_UpdateExisting(t *testing.T) {
	g := New()

	// Insert initial node with empty label and NodeOther type
	g.UpsertNode(&Node{ID: "n1", Label: "", Type: NodeOther, Count: 3})

	// Upsert with more info - should update label and type
	g.UpsertNode(&Node{ID: "n1", Label: "Updated Label", Type: NodeState, Count: 10})

	if len(g.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(g.Nodes))
	}
	got := g.Nodes["n1"]
	if got.Label != "Updated Label" {
		t.Errorf("expected label 'Updated Label', got '%s'", got.Label)
	}
	if got.Type != NodeState {
		t.Errorf("expected type NodeState, got %s", got.Type)
	}
	if got.Count != 10 {
		t.Errorf("expected count 10, got %d", got.Count)
	}
}

func TestUpsertNode_DoesNotDowngrade(t *testing.T) {
	g := New()

	// Insert node with good data
	g.UpsertNode(&Node{ID: "n1", Label: "Good Label", Type: NodeView, Count: 5})

	// Upsert with empty/lower values - should NOT downgrade
	g.UpsertNode(&Node{ID: "n1", Label: "", Type: NodeOther, Count: 2})

	got := g.Nodes["n1"]
	if got.Label != "Good Label" {
		t.Errorf("label should not be overwritten with empty, got '%s'", got.Label)
	}
	if got.Type != NodeView {
		t.Errorf("type should not be downgraded to NodeOther, got %s", got.Type)
	}
	// Count should NOT be downgraded (existing 5 > new 2)
	if got.Count != 5 {
		t.Errorf("count should not be downgraded, expected 5, got %d", got.Count)
	}
}

func TestAddEdge(t *testing.T) {
	g := New()

	g.AddEdge(Edge{From: "a", To: "b", Label: "causes"})
	g.AddEdge(Edge{From: "b", To: "c", Label: "updates"})

	if len(g.Edges) != 2 {
		t.Errorf("expected 2 edges, got %d", len(g.Edges))
	}

	e1 := g.Edges[0]
	if e1.From != "a" || e1.To != "b" || e1.Label != "causes" {
		t.Errorf("edge 0 mismatch: %+v", e1)
	}

	e2 := g.Edges[1]
	if e2.From != "b" || e2.To != "c" || e2.Label != "updates" {
		t.Errorf("edge 1 mismatch: %+v", e2)
	}
}

func TestNodeTypes(t *testing.T) {
	tests := []struct {
		nodeType NodeType
		expected string
	}{
		{NodeCause, "cause"},
		{NodeState, "state"},
		{NodeView, "view"},
		{NodeOther, "other"},
	}

	for _, tt := range tests {
		if string(tt.nodeType) != tt.expected {
			t.Errorf("NodeType %v: expected %q, got %q", tt.nodeType, tt.expected, string(tt.nodeType))
		}
	}
}
