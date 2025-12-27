package graph

type NodeType string

const (
	NodeCause NodeType = "cause"
	NodeState NodeType = "state"
	NodeView  NodeType = "view"
	NodeOther NodeType = "other"
)

type Node struct {
	ID    string
	Label string
	Type  NodeType
	Count int // optional metric (e.g. view updates)
}

type Edge struct {
	From  string
	To    string
	Label string
}

type Graph struct {
	Nodes map[string]*Node
	Edges []Edge
}

func New() *Graph {
	return &Graph{Nodes: map[string]*Node{}, Edges: []Edge{}}
}

func (g *Graph) UpsertNode(n *Node) {
	if existing, ok := g.Nodes[n.ID]; ok {
		if existing.Label == "" {
			existing.Label = n.Label
		}
		if existing.Type == NodeOther {
			existing.Type = n.Type
		}
		if n.Count > existing.Count {
			existing.Count = n.Count
		}
		return
	}
	g.Nodes[n.ID] = n
}

func (g *Graph) AddEdge(e Edge) {
	g.Edges = append(g.Edges, e)
}
