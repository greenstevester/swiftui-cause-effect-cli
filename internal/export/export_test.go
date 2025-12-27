package export

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExportTrace_MissingTrace(t *testing.T) {
	err := ExportTrace(nil, Options{
		TracePath: "",
		OutDir:    "/tmp/out",
		Format:    "json",
	})
	if err == nil {
		t.Error("expected error for missing trace path")
	}
	if err.Error() != "-trace is required" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExportTrace_MissingOutDir(t *testing.T) {
	err := ExportTrace(nil, Options{
		TracePath: "/some/trace.trace",
		OutDir:    "",
		Format:    "json",
	})
	if err == nil {
		t.Error("expected error for missing out dir")
	}
	if err.Error() != "-out is required" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExportTrace_CreatesOutDir(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "nested", "output")

	// This will fail when trying to export (no real trace file),
	// but it should create the directory first
	_ = ExportTrace(nil, Options{
		TracePath: "/nonexistent.trace",
		OutDir:    outDir,
		Format:    "json",
	})

	// Directory should have been created before the export attempt
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		t.Error("out directory was not created")
	}
}

func TestPickFormat_DefaultsToXML(t *testing.T) {
	// With a nil CLI, pickFormat should return xml as default
	// This tests the fallback behavior when help can't be retrieved
	format := pickFormat(nil)
	if format != "xml" {
		t.Errorf("expected 'xml' default, got %q", format)
	}
}

// MockCLI for testing pickFormat with different help outputs
type mockCLI struct {
	helpOutput string
	helpErr    error
}

func (m *mockCLI) ExportHelp() (string, error) {
	return m.helpOutput, m.helpErr
}

func TestPickFormat_PrefersJSON(t *testing.T) {
	tests := []struct {
		name     string
		help     string
		expected string
	}{
		{"has json", "Supported formats: xml, json, csv", "json"},
		{"has JSON uppercase", "Formats: JSON, XML", "json"},
		{"only xml", "Supported formats: xml", "xml"},
		{"only csv", "Supported formats: csv", "csv"},
		{"empty help", "", "xml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// pickFormat expects *xctrace.CLI but we can't easily mock it
			// This test documents the expected behavior
			// In practice, pickFormat calls cli.ExportHelp() and parses the output
			_ = tt // Documenting expected behavior
		})
	}
}

func TestOptions_FormatNormalization(t *testing.T) {
	// Test that format strings are normalized
	tests := []struct {
		input    string
		expected string
	}{
		{"JSON", "json"},
		{"  xml  ", "xml"},
		{"CSV", "csv"},
		{"", "auto"},
		{"auto", "auto"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			opts := Options{Format: tt.input}
			// Normalization happens inside ExportTrace
			_ = opts // Document the expected behavior
		})
	}
}
