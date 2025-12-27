package xctrace

import (
	"runtime"
	"testing"
)

func TestNew(t *testing.T) {
	cli := New()
	if cli == nil {
		t.Error("New() returned nil")
	}
}

func TestRecordOptions(t *testing.T) {
	opts := RecordOptions{
		Template:  "SwiftUI",
		Device:    "iPhone 15",
		App:       "com.example.app",
		TimeLimit: "30s",
		OutTrace:  "output.trace",
	}

	if opts.Template != "SwiftUI" {
		t.Errorf("Template: expected 'SwiftUI', got %q", opts.Template)
	}
	if opts.Device != "iPhone 15" {
		t.Errorf("Device: expected 'iPhone 15', got %q", opts.Device)
	}
	if opts.App != "com.example.app" {
		t.Errorf("App: expected 'com.example.app', got %q", opts.App)
	}
	if opts.TimeLimit != "30s" {
		t.Errorf("TimeLimit: expected '30s', got %q", opts.TimeLimit)
	}
	if opts.OutTrace != "output.trace" {
		t.Errorf("OutTrace: expected 'output.trace', got %q", opts.OutTrace)
	}
}

func TestExportOptions(t *testing.T) {
	opts := ExportOptions{
		TracePath:      "/path/to/trace.trace",
		OutDir:         "/output/dir",
		Format:         "json",
		AdditionalArgs: []string{"--extra", "arg"},
	}

	if opts.TracePath != "/path/to/trace.trace" {
		t.Errorf("TracePath: expected '/path/to/trace.trace', got %q", opts.TracePath)
	}
	if opts.OutDir != "/output/dir" {
		t.Errorf("OutDir: expected '/output/dir', got %q", opts.OutDir)
	}
	if opts.Format != "json" {
		t.Errorf("Format: expected 'json', got %q", opts.Format)
	}
	if len(opts.AdditionalArgs) != 2 {
		t.Errorf("AdditionalArgs: expected 2 args, got %d", len(opts.AdditionalArgs))
	}
}

func TestEnsureDarwin_OnDarwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping test on non-darwin platform")
	}

	// Should not panic on darwin
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ensureDarwin panicked on darwin: %v", r)
		}
	}()

	ensureDarwin()
}

// TestListTemplates_Integration tests that xctrace is accessible
// This is an integration test that requires macOS with Xcode installed
func TestListTemplates_Integration(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping integration test on non-darwin platform")
	}
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cli := New()
	output, err := cli.ListTemplates()
	if err != nil {
		t.Logf("ListTemplates failed (may be expected if Xcode not installed): %v", err)
		// Don't fail - Xcode may not be installed in CI
		return
	}

	if output == "" {
		t.Log("ListTemplates returned empty output")
	} else {
		t.Logf("Found templates:\n%s", output)
	}
}

// TestExportHelp_Integration tests the export help command
func TestExportHelp_Integration(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping integration test on non-darwin platform")
	}
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cli := New()
	output, err := cli.ExportHelp()
	if err != nil {
		t.Logf("ExportHelp failed (may be expected if Xcode not installed): %v", err)
		return
	}

	if output == "" {
		t.Log("ExportHelp returned empty output")
	}
}
