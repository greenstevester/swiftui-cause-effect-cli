package xctrace

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type CLI struct{}

func New() *CLI { return &CLI{} }

type RecordOptions struct {
	Template  string // Instruments template name
	Device    string // name or UDID (optional)
	App       string // bundle id or path to .app
	TimeLimit string // e.g. 10s, 1m
	OutTrace  string // output .trace path
}

type ExportOptions struct {
	TracePath       string
	OutDir          string
	Format          string // auto|xml|json|csv
	AdditionalArgs  []string
}

func (c *CLI) Record(opts RecordOptions) error {
	ensureDarwin()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	args := []string{"xctrace", "record"}
	if opts.Template != "" {
		args = append(args, "--template", opts.Template)
	}
	if opts.Device != "" {
		args = append(args, "--device", opts.Device)
	}
	if opts.TimeLimit != "" {
		args = append(args, "--time-limit", opts.TimeLimit)
	}
	if opts.OutTrace != "" {
		args = append(args, "--output", opts.OutTrace)
	}
	// Launch handling varies by Xcode version. We try a conservative form:
	//   --launch -- <bundle-id-or-app-path>
	// If it fails, users can record in Instruments and pass the .trace to `export`.
	args = append(args, "--launch", "--", opts.App)
	_, _, err := run(ctx, args)
	return err
}

func (c *CLI) Export(opts ExportOptions) (string, error) {
	ensureDarwin()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	args := []string{"xctrace", "export", "--input", opts.TracePath, "--output", opts.OutDir}
	if opts.Format != "" && opts.Format != "auto" {
		args = append(args, "--output-format", opts.Format)
	}
	args = append(args, opts.AdditionalArgs...)
	stdout, _, err := run(ctx, args)
	return stdout, err
}

func (c *CLI) ListTemplates() (string, error) {
	ensureDarwin()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	stdout, _, err := run(ctx, []string{"xctrace", "list", "templates"})
	return stdout, err
}

func (c *CLI) ExportHelp() (string, error) {
	ensureDarwin()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	stdout, _, err := run(ctx, []string{"xctrace", "export", "--help"})
	return stdout, err
}

func ensureDarwin() {
	if runtime.GOOS != "darwin" {
		panic("swiftuice requires macOS (darwin)")
	}
}

func run(ctx context.Context, args []string) (string, string, error) {
	cmd := exec.CommandContext(ctx, "xcrun", args...)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	stdout := outb.String()
	stderr := errb.String()
	if err != nil {
		msg := strings.TrimSpace(stderr)
		if msg == "" {
			msg = strings.TrimSpace(stdout)
		}
		if msg == "" {
			msg = err.Error()
		}
		return stdout, stderr, fmt.Errorf("xcrun %s failed: %s", strings.Join(args, " "), msg)
	}
	return stdout, stderr, nil
}
