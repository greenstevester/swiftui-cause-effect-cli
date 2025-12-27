package export

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/greenstevester/swiftui-cause-effect-cli/internal/xctrace"
)

type Options struct {
	TracePath string
	OutDir    string
	Format    string // auto|xml|json|csv
}

func ExportTrace(cli *xctrace.CLI, opts Options) error {
	if opts.TracePath == "" {
		return fmt.Errorf("-trace is required")
	}
	if opts.OutDir == "" {
		return fmt.Errorf("-out is required")
	}
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return fmt.Errorf("create out dir: %w", err)
	}

	format := strings.ToLower(strings.TrimSpace(opts.Format))
	if format == "" || format == "auto" {
		format = pickFormat(cli)
	}

	_, err := cli.Export(xctrace.ExportOptions{
		TracePath: opts.TracePath,
		OutDir:    opts.OutDir,
		Format:    format,
	})
	if err != nil {
		return err
	}

	meta := filepath.Join(opts.OutDir, "EXPORT_FORMAT.txt")
	_ = os.WriteFile(meta, []byte(format+"\n"), 0o644)
	return nil
}

func pickFormat(cli *xctrace.CLI) string {
	help, err := cli.ExportHelp()
	if err != nil {
		return "xml"
	}
	l := strings.ToLower(help)
	// Prefer JSON if supported; it's easiest to parse.
	if strings.Contains(l, "json") {
		return "json"
	}
	if strings.Contains(l, "xml") {
		return "xml"
	}
	if strings.Contains(l, "csv") {
		return "csv"
	}
	return "xml"
}
