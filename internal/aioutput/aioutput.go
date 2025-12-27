// Package aioutput provides structured output for AI agents and automation tools.
package aioutput

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/greenstevester/swiftui-cause-effect-cli/internal/correlation"
	"github.com/greenstevester/swiftui-cause-effect-cli/internal/graph"
	"github.com/greenstevester/swiftui-cause-effect-cli/internal/issues"
	"github.com/greenstevester/swiftui-cause-effect-cli/internal/suggestions"
)

// Report is the complete AI-friendly output structure
type Report struct {
	// Metadata
	Version   string    `json:"version"`
	Generated time.Time `json:"generated"`
	Tool      string    `json:"tool"`

	// Input information
	Input InputInfo `json:"input"`

	// Summary statistics
	Summary Summary `json:"summary"`

	// Detected issues with fixes
	Issues []IssueWithFixes `json:"issues"`

	// The cause-effect graph
	Graph GraphData `json:"graph"`

	// Source code correlations
	SourceCorrelations []correlation.SourceMatch `json:"source_correlations,omitempty"`

	// High-level recommendations
	Recommendations []suggestions.Recommendation `json:"recommendations"`

	// AI agent instructions
	AgentInstructions AgentInstructions `json:"agent_instructions"`
}

// InputInfo describes what was analyzed
type InputInfo struct {
	TracePath    string `json:"trace_path,omitempty"`
	ExportDir    string `json:"export_dir,omitempty"`
	SourceRoot   string `json:"source_root,omitempty"`
	FilesParsed  int    `json:"files_parsed"`
	SwiftFiles   int    `json:"swift_files,omitempty"`
}

// Summary provides high-level metrics
type Summary struct {
	TotalCauses       int     `json:"total_causes"`
	TotalStateChanges int     `json:"total_state_changes"`
	TotalViewUpdates  int     `json:"total_view_updates"`
	TotalEdges        int     `json:"total_edges"`
	IssuesFound       int     `json:"issues_found"`
	CriticalIssues    int     `json:"critical_issues"`
	HighIssues        int     `json:"high_issues"`
	PerformanceScore  int     `json:"performance_score"` // 0-100
	HealthStatus      string  `json:"health_status"`     // good, warning, critical
}

// IssueWithFixes combines an issue with its applicable fixes
type IssueWithFixes struct {
	issues.Issue
	SuggestedFixes []suggestions.Fix `json:"suggested_fixes"`
}

// GraphData is a simplified graph representation for AI consumption
type GraphData struct {
	Nodes []NodeData `json:"nodes"`
	Edges []EdgeData `json:"edges"`
}

// NodeData is a node in AI-friendly format
type NodeData struct {
	ID          string  `json:"id"`
	Label       string  `json:"label"`
	Type        string  `json:"type"` // cause, state, view, other
	UpdateCount int     `json:"update_count,omitempty"`
	SourceFile  string  `json:"source_file,omitempty"`
	LineNumber  int     `json:"line_number,omitempty"`
	Confidence  float64 `json:"source_confidence,omitempty"`
}

// EdgeData is an edge in AI-friendly format
type EdgeData struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label,omitempty"`
}

// AgentInstructions provides guidance for AI agents
type AgentInstructions struct {
	TaskDescription string   `json:"task_description"`
	Priority        []string `json:"priority"`
	Constraints     []string `json:"constraints"`
	SuccessCriteria []string `json:"success_criteria"`
	Context         string   `json:"context"`
}

// Generator creates AI reports
type Generator struct {
	detector   *issues.Detector
	correlator *correlation.Correlator
}

// NewGenerator creates a report generator
func NewGenerator(sourceRoot string) (*Generator, error) {
	var correlator *correlation.Correlator
	var err error

	if sourceRoot != "" {
		correlator, err = correlation.NewCorrelator(sourceRoot)
		if err != nil {
			// Non-fatal - we can still generate report without source correlation
			correlator = nil
		}
	}

	return &Generator{
		detector:   issues.NewDetector(),
		correlator: correlator,
	}, nil
}

// GenerateOptions configures report generation
type GenerateOptions struct {
	TracePath   string
	ExportDir   string
	SourceRoot  string
	FilesParsed int
}

// Generate creates a complete AI report from a graph
func (g *Generator) Generate(gr *graph.Graph, opts GenerateOptions) *Report {
	// Detect issues
	detectedIssues := g.detector.Detect(gr)

	// Generate fixes for each issue
	issuesWithFixes := make([]IssueWithFixes, len(detectedIssues))
	for i, issue := range detectedIssues {
		issuesWithFixes[i] = IssueWithFixes{
			Issue:          issue,
			SuggestedFixes: suggestions.GenerateFixes(issue),
		}
	}

	// Correlate with source if available
	var sourceMatches []correlation.SourceMatch
	if g.correlator != nil {
		sourceMatches = g.correlator.Correlate(gr)
	}

	// Build graph data with source info
	graphData := g.buildGraphData(gr, sourceMatches)

	// Calculate summary
	summary := g.calculateSummary(gr, detectedIssues)

	// Generate recommendations
	recs := suggestions.GenerateRecommendations(detectedIssues)

	// Build agent instructions
	agentInstructions := g.buildAgentInstructions(summary, detectedIssues)

	swiftFiles := 0
	if g.correlator != nil {
		swiftFiles = g.correlator.SwiftFileCount()
	}

	return &Report{
		Version:   "1.0",
		Generated: time.Now().UTC(),
		Tool:      "swiftuice",
		Input: InputInfo{
			TracePath:   opts.TracePath,
			ExportDir:   opts.ExportDir,
			SourceRoot:  opts.SourceRoot,
			FilesParsed: opts.FilesParsed,
			SwiftFiles:  swiftFiles,
		},
		Summary:            summary,
		Issues:             issuesWithFixes,
		Graph:              graphData,
		SourceCorrelations: sourceMatches,
		Recommendations:    recs,
		AgentInstructions:  agentInstructions,
	}
}

func (g *Generator) buildGraphData(gr *graph.Graph, matches []correlation.SourceMatch) GraphData {
	// Build lookup for source matches
	matchLookup := make(map[string]*correlation.SourceMatch)
	for i := range matches {
		if existing, ok := matchLookup[matches[i].TraceNodeID]; !ok || matches[i].Confidence > existing.Confidence {
			matchLookup[matches[i].TraceNodeID] = &matches[i]
		}
	}

	nodes := make([]NodeData, 0, len(gr.Nodes))
	for _, node := range gr.Nodes {
		nd := NodeData{
			ID:          node.ID,
			Label:       node.Label,
			Type:        string(node.Type),
			UpdateCount: node.Count,
		}
		if match, ok := matchLookup[node.ID]; ok {
			nd.SourceFile = match.RelativePath
			nd.LineNumber = match.LineNumber
			nd.Confidence = match.Confidence
		}
		nodes = append(nodes, nd)
	}

	edges := make([]EdgeData, 0, len(gr.Edges))
	for _, edge := range gr.Edges {
		edges = append(edges, EdgeData{
			From:  edge.From,
			To:    edge.To,
			Label: edge.Label,
		})
	}

	return GraphData{Nodes: nodes, Edges: edges}
}

func (g *Generator) calculateSummary(gr *graph.Graph, detected []issues.Issue) Summary {
	var causes, states, views int
	for _, node := range gr.Nodes {
		switch node.Type {
		case graph.NodeCause:
			causes++
		case graph.NodeState:
			states++
		case graph.NodeView:
			views++
		}
	}

	var critical, high int
	for _, issue := range detected {
		switch issue.Severity {
		case issues.SeverityCritical:
			critical++
		case issues.SeverityHigh:
			high++
		}
	}

	// Calculate performance score (100 = no issues, 0 = critical problems)
	score := 100
	score -= critical * 25
	score -= high * 10
	score -= (len(detected) - critical - high) * 3
	if score < 0 {
		score = 0
	}

	status := "good"
	if score < 50 {
		status = "critical"
	} else if score < 75 {
		status = "warning"
	}

	return Summary{
		TotalCauses:       causes,
		TotalStateChanges: states,
		TotalViewUpdates:  views,
		TotalEdges:        len(gr.Edges),
		IssuesFound:       len(detected),
		CriticalIssues:    critical,
		HighIssues:        high,
		PerformanceScore:  score,
		HealthStatus:      status,
	}
}

func (g *Generator) buildAgentInstructions(summary Summary, detected []issues.Issue) AgentInstructions {
	var priority []string

	// Prioritize by issue severity
	for _, issue := range detected {
		if issue.Severity == issues.SeverityCritical || issue.Severity == issues.SeverityHigh {
			priority = append(priority, fmt.Sprintf("[%s] %s", issue.Severity, issue.Title))
		}
	}
	if len(priority) == 0 {
		priority = append(priority, "Review medium-priority issues if any")
	}

	constraints := []string{
		"Maintain existing functionality - do not break features",
		"Prefer minimal changes over large refactors",
		"Test changes thoroughly before committing",
		"Consider iOS version compatibility of suggested fixes",
		"Preserve existing code style and patterns",
	}

	successCriteria := []string{
		"Reduce view update counts for flagged views",
		"Eliminate or mitigate critical and high severity issues",
		"Improve performance score (current: " + fmt.Sprintf("%d", summary.PerformanceScore) + ")",
		"Verify fixes with Instruments after changes",
	}

	context := fmt.Sprintf(
		"SwiftUI performance analysis found %d issues (%d critical, %d high). "+
			"The cause-effect graph shows %d causes triggering %d state changes affecting %d views. "+
			"Focus on reducing unnecessary view updates and breaking cascade chains.",
		summary.IssuesFound, summary.CriticalIssues, summary.HighIssues,
		summary.TotalCauses, summary.TotalStateChanges, summary.TotalViewUpdates,
	)

	taskDesc := "Analyze the SwiftUI performance issues and implement fixes to improve rendering efficiency."
	if summary.HealthStatus == "critical" {
		taskDesc = "URGENT: Critical SwiftUI performance issues detected. Prioritize fixes to prevent frame drops and poor user experience."
	} else if summary.HealthStatus == "warning" {
		taskDesc = "SwiftUI performance issues detected that may impact user experience. Review and implement suggested fixes."
	}

	return AgentInstructions{
		TaskDescription: taskDesc,
		Priority:        priority,
		Constraints:     constraints,
		SuccessCriteria: successCriteria,
		Context:         context,
	}
}

// WriteJSON writes the report as formatted JSON to a file
func (r *Report) WriteJSON(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// ToJSON returns the report as a JSON string
func (r *Report) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToCompactJSON returns the report as compact JSON (for piping)
func (r *Report) ToCompactJSON() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
