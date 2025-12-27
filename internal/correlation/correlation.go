// Package correlation matches trace data to Swift source files.
package correlation

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/greenstevester/swiftui-cause-effect-cli/internal/graph"
)

// SourceMatch represents a correlation between trace data and source code
type SourceMatch struct {
	TraceNodeID   string  `json:"trace_node_id"`
	TraceLabel    string  `json:"trace_label"`
	NodeType      string  `json:"node_type"`
	FilePath      string  `json:"file_path"`
	RelativePath  string  `json:"relative_path"`
	LineNumber    int     `json:"line_number"`
	CodeSnippet   string  `json:"code_snippet,omitempty"`
	MatchType     string  `json:"match_type"` // exact, fuzzy, inferred
	Confidence    float64 `json:"confidence"` // 0.0 - 1.0
	MatchedSymbol string  `json:"matched_symbol,omitempty"`
}

// Correlator finds source file locations for graph nodes
type Correlator struct {
	sourceRoot string
	swiftFiles []string
	cache      map[string][]SourceMatch
}

// NewCorrelator creates a correlator for a Swift project
func NewCorrelator(sourceRoot string) (*Correlator, error) {
	c := &Correlator{
		sourceRoot: sourceRoot,
		cache:      make(map[string][]SourceMatch),
	}

	if err := c.indexSwiftFiles(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Correlator) indexSwiftFiles() error {
	return filepath.WalkDir(c.sourceRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if d.IsDir() {
			// Skip common non-source directories
			name := d.Name()
			if name == ".git" || name == "build" || name == "DerivedData" ||
				name == "Pods" || name == ".build" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(strings.ToLower(path), ".swift") {
			c.swiftFiles = append(c.swiftFiles, path)
		}
		return nil
	})
}

// Correlate finds source matches for all nodes in a graph
func (c *Correlator) Correlate(g *graph.Graph) []SourceMatch {
	var matches []SourceMatch

	for _, node := range g.Nodes {
		nodeMatches := c.findMatchesForNode(node)
		matches = append(matches, nodeMatches...)
	}

	return matches
}

// CorrelateNode finds source matches for a single node
func (c *Correlator) CorrelateNode(node *graph.Node) []SourceMatch {
	return c.findMatchesForNode(node)
}

func (c *Correlator) findMatchesForNode(node *graph.Node) []SourceMatch {
	if cached, ok := c.cache[node.ID]; ok {
		return cached
	}

	var matches []SourceMatch

	// Extract potential symbol names from the node label
	symbols := extractSymbols(node.Label)

	for _, filePath := range c.swiftFiles {
		fileMatches := c.searchFileForSymbols(filePath, node, symbols)
		matches = append(matches, fileMatches...)
	}

	// Sort by confidence
	sortByConfidence(matches)

	// Cache results
	c.cache[node.ID] = matches

	return matches
}

func extractSymbols(label string) []string {
	var symbols []string

	// Extract View names (CamelCase identifiers)
	viewPattern := regexp.MustCompile(`\b([A-Z][a-zA-Z0-9]*(?:View|Screen|Page|Cell|Row|Item)?)\b`)
	for _, match := range viewPattern.FindAllString(label, -1) {
		symbols = append(symbols, match)
	}

	// Extract property names (@State, @ObservedObject, etc.)
	propPattern := regexp.MustCompile(`@(?:State|ObservedObject|StateObject|EnvironmentObject|Binding|Environment)\s+(?:var\s+)?(\w+)`)
	for _, match := range propPattern.FindAllStringSubmatch(label, -1) {
		if len(match) > 1 {
			symbols = append(symbols, match[1])
		}
	}

	// Extract any identifier-like strings
	identPattern := regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\b`)
	for _, match := range identPattern.FindAllString(label, -1) {
		// Filter out common words
		if !isCommonWord(match) {
			symbols = append(symbols, match)
		}
	}

	return dedupe(symbols)
}

func isCommonWord(s string) bool {
	common := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"for": true, "in": true, "to": true, "of": true, "with": true,
		"var": true, "let": true, "func": true, "struct": true, "class": true,
		"view": true, "body": true, "some": true, "any": true,
		"true": true, "false": true, "nil": true,
		"update": true, "change": true, "trigger": true, "cause": true,
	}
	return common[strings.ToLower(s)]
}

func dedupe(strs []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range strs {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

func (c *Correlator) searchFileForSymbols(filePath string, node *graph.Node, symbols []string) []SourceMatch {
	var matches []SourceMatch

	file, err := os.Open(filePath)
	if err != nil {
		return matches
	}
	defer file.Close()

	relPath, _ := filepath.Rel(c.sourceRoot, filePath)
	if relPath == "" {
		relPath = filePath
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		for _, symbol := range symbols {
			match := c.matchLineForSymbol(line, symbol, node, filePath, relPath, lineNum)
			if match != nil {
				matches = append(matches, *match)
			}
		}
	}

	return matches
}

func (c *Correlator) matchLineForSymbol(line, symbol string, node *graph.Node, filePath, relPath string, lineNum int) *SourceMatch {
	// Skip if symbol not in line
	if !strings.Contains(line, symbol) {
		return nil
	}

	confidence := 0.0
	matchType := "fuzzy"
	matchedSymbol := symbol

	trimmedLine := strings.TrimSpace(line)

	switch node.Type {
	case graph.NodeView:
		// Look for struct declarations of Views
		if matched, conf := matchViewDeclaration(line, symbol); matched {
			confidence = conf
			matchType = "exact"
		} else if matched, conf := matchViewBody(line, symbol); matched {
			confidence = conf
			matchType = "inferred"
		}

	case graph.NodeState:
		// Look for @State, @ObservedObject, etc.
		if matched, conf := matchStateDeclaration(line, symbol); matched {
			confidence = conf
			matchType = "exact"
		}

	case graph.NodeCause:
		// Look for Button, .onTapGesture, Timer, etc.
		if matched, conf := matchCausePattern(line, symbol); matched {
			confidence = conf
			matchType = "exact"
		}

	default:
		// Generic symbol match
		if strings.Contains(line, symbol) {
			confidence = 0.3
		}
	}

	if confidence < 0.3 {
		return nil
	}

	return &SourceMatch{
		TraceNodeID:   node.ID,
		TraceLabel:    node.Label,
		NodeType:      string(node.Type),
		FilePath:      filePath,
		RelativePath:  relPath,
		LineNumber:    lineNum,
		CodeSnippet:   truncate(trimmedLine, 120),
		MatchType:     matchType,
		Confidence:    confidence,
		MatchedSymbol: matchedSymbol,
	}
}

func matchViewDeclaration(line, symbol string) (bool, float64) {
	// struct MyView: View
	pattern := regexp.MustCompile(`struct\s+` + regexp.QuoteMeta(symbol) + `\s*:\s*(?:\w+,\s*)*View`)
	if pattern.MatchString(line) {
		return true, 0.95
	}

	// Just struct declaration with View-like name
	pattern2 := regexp.MustCompile(`struct\s+` + regexp.QuoteMeta(symbol) + `\b`)
	if pattern2.MatchString(line) && strings.Contains(strings.ToLower(symbol), "view") {
		return true, 0.85
	}

	return false, 0.0
}

func matchViewBody(line, symbol string) (bool, float64) {
	// var body: some View
	if strings.Contains(line, "var body") && strings.Contains(line, "View") {
		return true, 0.5
	}
	return false, 0.0
}

func matchStateDeclaration(line, symbol string) (bool, float64) {
	// @State var symbol
	statePattern := regexp.MustCompile(`@(?:State|StateObject)\s+(?:private\s+)?var\s+` + regexp.QuoteMeta(symbol) + `\b`)
	if statePattern.MatchString(line) {
		return true, 0.95
	}

	// @ObservedObject var symbol
	observedPattern := regexp.MustCompile(`@(?:ObservedObject|EnvironmentObject)\s+(?:private\s+)?var\s+` + regexp.QuoteMeta(symbol) + `\b`)
	if observedPattern.MatchString(line) {
		return true, 0.9
	}

	// @Binding var symbol
	bindingPattern := regexp.MustCompile(`@Binding\s+(?:private\s+)?var\s+` + regexp.QuoteMeta(symbol) + `\b`)
	if bindingPattern.MatchString(line) {
		return true, 0.85
	}

	return false, 0.0
}

func matchCausePattern(line, symbol string) (bool, float64) {
	lowerLine := strings.ToLower(line)
	lowerSymbol := strings.ToLower(symbol)

	// Button action
	if strings.Contains(lowerLine, "button") && strings.Contains(lowerLine, lowerSymbol) {
		return true, 0.85
	}

	// Gesture handlers
	gestures := []string{"ontapgesture", "ondraggesture", "onlongpressgesture", "gesture"}
	for _, g := range gestures {
		if strings.Contains(lowerLine, g) {
			return true, 0.8
		}
	}

	// Timer
	if strings.Contains(lowerLine, "timer") {
		return true, 0.75
	}

	// Notification
	if strings.Contains(lowerLine, "notificationcenter") || strings.Contains(lowerLine, "onreceive") {
		return true, 0.7
	}

	return false, 0.0
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func sortByConfidence(matches []SourceMatch) {
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].Confidence > matches[i].Confidence {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}
}

// BestMatch returns the highest confidence match for a node ID
func (c *Correlator) BestMatch(nodeID string) *SourceMatch {
	if matches, ok := c.cache[nodeID]; ok && len(matches) > 0 {
		return &matches[0]
	}
	return nil
}

// GetSourceRoot returns the configured source root
func (c *Correlator) GetSourceRoot() string {
	return c.sourceRoot
}

// SwiftFileCount returns the number of indexed Swift files
func (c *Correlator) SwiftFileCount() int {
	return len(c.swiftFiles)
}
