package analyze

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/greenstevester/swiftui-cause-effect-cli/internal/export"
	"github.com/greenstevester/swiftui-cause-effect-cli/internal/graph"
	"github.com/greenstevester/swiftui-cause-effect-cli/internal/xctrace"
)

var ErrNoData = errors.New("no parseable cause-and-effect data")

type Options struct {
	Input      string // dir from export OR .trace path
	OutSummary string
	OutDOT     string
	XcTrace    *xctrace.CLI
}

type Result struct {
	SummaryPath string
	DotPath     string
}

// AnalysisResult contains the parsed graph and metadata for further processing
type AnalysisResult struct {
	Graph       *graph.Graph
	InputDir    string
	FilesParsed int
	Hints       []string
}

// ParseTrace parses a trace or export directory and returns the graph for further analysis
func ParseTrace(opts Options) (*AnalysisResult, error) {
	if opts.XcTrace == nil {
		opts.XcTrace = xctrace.New()
	}
	inputInfo, err := os.Stat(opts.Input)
	if err != nil {
		return nil, err
	}

	inputDir := opts.Input
	if !inputInfo.IsDir() && strings.HasSuffix(strings.ToLower(opts.Input), ".trace") {
		tmpDir := filepath.Join(filepath.Dir(opts.Input), "exported")
		if err := export.ExportTrace(opts.XcTrace, export.Options{TracePath: opts.Input, OutDir: tmpDir, Format: "auto"}); err != nil {
			return nil, err
		}
		inputDir = tmpDir
	}

	g := graph.New()
	stats := &summaryStats{}
	if err := parseDirectory(inputDir, g, stats); err != nil {
		return nil, err
	}
	if len(g.Nodes) == 0 || len(g.Edges) == 0 {
		return nil, ErrNoData
	}

	return &AnalysisResult{
		Graph:       g,
		InputDir:    inputDir,
		FilesParsed: stats.FilesParsed,
		Hints:       stats.Hints,
	}, nil
}

func Summarize(opts Options) (Result, error) {
	if opts.XcTrace == nil {
		opts.XcTrace = xctrace.New()
	}
	inputInfo, err := os.Stat(opts.Input)
	if err != nil {
		return Result{}, err
	}

	inputDir := opts.Input
	if !inputInfo.IsDir() && strings.HasSuffix(strings.ToLower(opts.Input), ".trace") {
		// Convenience: if user passed a .trace, export it first.
		tmpDir := filepath.Join(filepath.Dir(opts.Input), "exported")
		if err := export.ExportTrace(opts.XcTrace, export.Options{TracePath: opts.Input, OutDir: tmpDir, Format: "auto"}); err != nil {
			return Result{}, err
		}
		inputDir = tmpDir
	}

	g := graph.New()
	stats := &summaryStats{}
	if err := parseDirectory(inputDir, g, stats); err != nil {
		return Result{}, err
	}
	if len(g.Nodes) == 0 || len(g.Edges) == 0 {
		return Result{}, ErrNoData
	}

	summary := renderMarkdown(g, stats)
	if err := os.WriteFile(opts.OutSummary, []byte(summary), 0o644); err != nil {
		return Result{}, err
	}
	dot := renderDOT(g)
	if err := os.WriteFile(opts.OutDOT, []byte(dot), 0o644); err != nil {
		return Result{}, err
	}
	return Result{SummaryPath: opts.OutSummary, DotPath: opts.OutDOT}, nil
}

type summaryStats struct {
	FilesParsed int
	Hints       []string
}

func parseDirectory(dir string, g *graph.Graph, stats *summaryStats) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".json":
			if err := parseJSON(path, g, stats); err != nil {
				// best-effort: keep going
				stats.Hints = append(stats.Hints, fmt.Sprintf("JSON parse skipped %s: %v", filepath.Base(path), err))
			}
			stats.FilesParsed++
		case ".xml", ".csv", ".txt":
			if err := parseTextLike(path, g, stats); err != nil {
				stats.Hints = append(stats.Hints, fmt.Sprintf("text parse skipped %s: %v", filepath.Base(path), err))
			}
			stats.FilesParsed++
		}
		return nil
	})
}

// parseJSON tries to interpret a few likely export shapes.
func parseJSON(path string, g *graph.Graph, stats *summaryStats) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Strategy 1: generic "nodes" + "edges" arrays (common graph export shape)
	var obj map[string]any
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}

	nodesRaw, hasNodes := obj["nodes"].([]any)
	edgesRaw, hasEdges := obj["edges"].([]any)
	if hasNodes && hasEdges {
		for _, n := range nodesRaw {
			m, ok := n.(map[string]any)
			if !ok {
				continue
			}
			id := asString(m["id"], asString(m["uuid"], ""))
			label := asString(m["label"], asString(m["title"], ""))
			kind := strings.ToLower(asString(m["type"], asString(m["kind"], "")))
			count := asInt(m["count"], asInt(m["updates"], 0))
			g.UpsertNode(&graph.Node{ID: idOrHash(id, label), Label: label, Type: classify(kind, label), Count: count})
		}
		for _, e := range edgesRaw {
			m, ok := e.(map[string]any)
			if !ok {
				continue
			}
			from := asString(m["from"], asString(m["source"], ""))
			to := asString(m["to"], asString(m["target"], ""))
			label := asString(m["label"], asString(m["reason"], ""))
			if from == "" || to == "" {
				continue
			}
			g.AddEdge(graph.Edge{From: from, To: to, Label: label})
		}
		return nil
	}

	// Strategy 2: scan JSON strings for cause/state/view triplets.
	return parseTextReader(strings.NewReader(string(b)), g, stats)
}

func parseTextLike(path string, g *graph.Graph, stats *summaryStats) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return parseTextReader(f, g, stats)
}

var (
	reState = regexp.MustCompile(`(?i)(state\s+change|\bstate\b|@state|@observedobject|@stateobject|\benvironment\b)`) // heuristic
	reView  = regexp.MustCompile(`(?i)(view\s+body\s+update|view\s+update|\bbody\(\)|\bView\b)`)                     // heuristic
	reCause = regexp.MustCompile(`(?i)(gesture|tap|button|timer|notification|publisher|async|network|animation|scene)`)      // heuristic
)

// parseTextReader is a fallback that builds a graph from recognizable tokens.
func parseTextReader(r io.Reader, g *graph.Graph, stats *summaryStats) error {
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)

	var lastCause string
	var lastState string
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		switch {
		case reState.MatchString(line):
			id := idOrHash("", line)
			g.UpsertNode(&graph.Node{ID: id, Label: trim(line, 120), Type: graph.NodeState})
			lastState = id
			if lastCause != "" {
				g.AddEdge(graph.Edge{From: lastCause, To: lastState, Label: "causes"})
			}
		case reView.MatchString(line):
			vid := idOrHash("", line)
			g.UpsertNode(&graph.Node{ID: vid, Label: trim(line, 120), Type: graph.NodeView})
			if lastState != "" {
				g.AddEdge(graph.Edge{From: lastState, To: vid, Label: "updates"})
			}
		case reCause.MatchString(line):
			cid := idOrHash("", line)
			g.UpsertNode(&graph.Node{ID: cid, Label: trim(line, 120), Type: graph.NodeCause})
			lastCause = cid
		}
	}
	return s.Err()
}

func renderDOT(g *graph.Graph) string {
	var b strings.Builder
	b.WriteString("digraph CauseEffect {\n")
	b.WriteString("  rankdir=LR;\n")
	for _, n := range g.Nodes {
		shape := "box"
		switch n.Type {
		case graph.NodeCause:
			shape = "ellipse"
		case graph.NodeState:
			shape = "diamond"
		case graph.NodeView:
			shape = "box"
		}
		label := escapeDOT(n.Label)
		if n.Count > 0 {
			label = fmt.Sprintf("%s\\ncount=%d", label, n.Count)
		}
		b.WriteString(fmt.Sprintf("  \"%s\" [shape=%s,label=\"%s\"];\n", n.ID, shape, label))
	}
	for _, e := range g.Edges {
		lbl := escapeDOT(e.Label)
		if lbl != "" {
			b.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\"];\n", e.From, e.To, lbl))
		} else {
			b.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", e.From, e.To))
		}
	}
	b.WriteString("}\n")
	return b.String()
}

func renderMarkdown(g *graph.Graph, stats *summaryStats) string {
	var causes, states, views []*graph.Node
	for _, n := range g.Nodes {
		switch n.Type {
		case graph.NodeCause:
			causes = append(causes, n)
		case graph.NodeState:
			states = append(states, n)
		case graph.NodeView:
			views = append(views, n)
		}
	}
	sort.Slice(views, func(i, j int) bool { return views[i].Count > views[j].Count })
	if len(views) > 10 {
		views = views[:10]
	}

	var b strings.Builder
	b.WriteString("# SwiftUI Cause & Effect Summary\n\n")
	b.WriteString(fmt.Sprintf("Parsed %d files. Nodes: %d, Edges: %d.\n\n", stats.FilesParsed, len(g.Nodes), len(g.Edges)))
	b.WriteString("## What this tool could extract\n")
	b.WriteString(fmt.Sprintf("- Causes: %d\n- State changes: %d\n- View updates: %d\n\n", len(causes), len(states), len(views)))
	b.WriteString("## Top view-update nodes (best effort)\n")
	if len(views) == 0 {
		b.WriteString("No explicit counts found in exported data.\n\n")
	} else {
		for _, v := range views {
			b.WriteString(fmt.Sprintf("- %s (count=%d)\n", v.Label, v.Count))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Notes\n")
	b.WriteString("- The SwiftUI Cause & Effect Graph is collected by the SwiftUI instrument (Xcode 26) and is primarily designed for interactive use in Instruments.\n")
	b.WriteString("- Export schemas can change; this CLI uses heuristic parsing and may miss relationships.\n")
	b.WriteString("- If export produced no parseable artifacts, open the .trace in Instruments and use the Cause & Effect Graph UI.\n\n")

	if len(stats.Hints) > 0 {
		b.WriteString("## Parse hints\n")
		for _, h := range stats.Hints {
			b.WriteString("- " + h + "\n")
		}
		b.WriteString("\n")
	}
	b.WriteString("## Outputs\n")
	b.WriteString("- Graphviz: see the generated `.dot` file (render with `dot -Tpng graph.dot -o graph.png`).\n")
	return b.String()
}

func asString(v any, def string) string {
	if v == nil {
		return def
	}
	s, ok := v.(string)
	if ok {
		return s
	}
	return def
}

func asInt(v any, def int) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	case json.Number:
		i, err := t.Int64()
		if err == nil {
			return int(i)
		}
	}
	return def
}

func classify(kind string, label string) graph.NodeType {
	k := strings.ToLower(kind)
	l := strings.ToLower(label)
	if strings.Contains(k, "state") || reState.MatchString(l) {
		return graph.NodeState
	}
	if strings.Contains(k, "view") || reView.MatchString(l) {
		return graph.NodeView
	}
	if strings.Contains(k, "cause") || reCause.MatchString(l) {
		return graph.NodeCause
	}
	return graph.NodeOther
}

func trim(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func escapeDOT(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}

func idOrHash(id string, label string) string {
	if id != "" {
		return id
	}
	// stable-ish ID derived from label
	h := 0
	for _, r := range label {
		h = (h*31 + int(r)) & 0x7fffffff
	}
	return fmt.Sprintf("n%d", h)
}
