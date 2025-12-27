package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/greenstevester/swiftui-cause-effect-cli/internal/analyze"
	"github.com/greenstevester/swiftui-cause-effect-cli/internal/export"
	"github.com/greenstevester/swiftui-cause-effect-cli/internal/xctrace"
)

func main() {
	os.Exit(run())
}

func run() int {
	if len(os.Args) < 2 {
		usage()
		return 2
	}

	sub := os.Args[1]
	switch sub {
	case "record":
		return cmdRecord(os.Args[2:])
	case "export":
		return cmdExport(os.Args[2:])
	case "summarize":
		return cmdSummarize(os.Args[2:])
	case "version":
		fmt.Println("swiftuice v0.1.0")
		return 0
	case "help", "-h", "--help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", sub)
		usage()
		return 2
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `swiftuice — extract & summarize SwiftUI Cause & Effect Graph traces

Usage:
  swiftuice record    [flags]
  swiftuice export    [flags]
  swiftuice summarize [flags]

Commands:
  record     Record an Instruments trace using xctrace (expects the SwiftUI template to exist)
  export     Export a .trace file into parseable artifacts (XML/JSON/CSV) using xctrace
  summarize  Build a best-effort Cause→State→View graph + a human summary + Graphviz .dot

Run 'swiftuice <command> -h' for command flags.`)
}

func cmdRecord(args []string) int {
	fs := flag.NewFlagSet("record", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var template string
	var device string
	var app string
	var timeLimit string
	var out string
	fs.StringVar(&template, "template", "SwiftUI", "Instruments template name (e.g. SwiftUI)")
	fs.StringVar(&device, "device", "", "Device name or UDID (optional; defaults to whatever xctrace picks)")
	fs.StringVar(&app, "app", "", "App bundle id (preferred) or full path to .app")
	fs.StringVar(&timeLimit, "time", "10s", "Time limit (e.g. 10s, 1m)")
	fs.StringVar(&out, "out", "swiftui.trace", "Output .trace path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if app == "" {
		fmt.Fprintln(os.Stderr, "-app is required")
		return 2
	}

	cli := xctrace.New()
	if err := cli.Record(xctrace.RecordOptions{
		Template:  template,
		Device:    device,
		App:       app,
		TimeLimit: timeLimit,
		OutTrace:  out,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "record failed:", err)
		return 1
	}
	fmt.Println(out)
	return 0
}

func cmdExport(args []string) int {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var inTrace string
	var outDir string
	var format string
	fs.StringVar(&inTrace, "trace", "", "Input .trace path")
	fs.StringVar(&outDir, "out", "exported", "Output directory")
	fs.StringVar(&format, "format", "auto", "Export format: auto|xml|json|csv")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if inTrace == "" {
		fmt.Fprintln(os.Stderr, "-trace is required")
		return 2
	}

	cli := xctrace.New()
	if err := export.ExportTrace(cli, export.Options{TracePath: inTrace, OutDir: outDir, Format: format}); err != nil {
		fmt.Fprintln(os.Stderr, "export failed:", err)
		return 1
	}
	fmt.Println(outDir)
	return 0
}

func cmdSummarize(args []string) int {
	fs := flag.NewFlagSet("summarize", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var input string
	var out string
	var dot string
	fs.StringVar(&input, "in", "", "Input directory (from export) OR a .trace path")
	fs.StringVar(&out, "out", "summary.md", "Summary markdown output")
	fs.StringVar(&dot, "dot", "graph.dot", "Graphviz .dot output")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if input == "" {
		fmt.Fprintln(os.Stderr, "-in is required")
		return 2
	}

	cli := xctrace.New()
	res, err := analyze.Summarize(analyze.Options{Input: input, OutSummary: out, OutDOT: dot, XcTrace: cli})
	if err != nil {
		if errors.Is(err, analyze.ErrNoData) {
			fmt.Fprintln(os.Stderr, "no parseable Cause & Effect data found; see trace/export limitations")
			return 3
		}
		fmt.Fprintln(os.Stderr, "summarize failed:", err)
		return 1
	}
	fmt.Printf("%s\n%s\n", res.SummaryPath, res.DotPath)
	return 0
}
