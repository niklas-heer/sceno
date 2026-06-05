package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/niklas-heer/sceno/internal/advise"
	"github.com/niklas-heer/sceno/internal/docs"
	"github.com/niklas-heer/sceno/internal/export"
	"github.com/niklas-heer/sceno/internal/guide"
	"github.com/niklas-heer/sceno/internal/inspect"
	"github.com/niklas-heer/sceno/internal/spec"
	"github.com/niklas-heer/sceno/internal/validate"
	"github.com/niklas-heer/sceno/internal/version"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "-V", "--version", "version":
		cmdVersion(args)
	case "init":
		cmdInit(args)
	case "validate":
		cmdValidate(args)
	case "advise":
		cmdAdvise(args)
	case "describe":
		cmdDescribe(args)
	case "render":
		cmdRender(args)
	case "docs":
		cmdDocs(args)
	case "help", "-h", "--help":
		usage()
	default:
		if runLegacy(cmd, args) {
			return
		}
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		usage()
		os.Exit(2)
	}
}

// runLegacy handles deprecated aliases and prints a short redirect hint.
func runLegacy(cmd string, args []string) bool {
	switch cmd {
	case "check":
		legacyHint("validate", "sceno validate -i FILE --json")
		cmdValidate(args)
		return true
	case "feedback":
		legacyHint("describe", "sceno describe -i FILE --json")
		cmdDescribe(args)
		return true
	case "guide", "agent":
		legacyHint("docs guide", "sceno docs guide --json")
		cmdGuide(args)
		return true
	case "spec":
		legacyHint("docs spec", "sceno docs spec")
		cmdSpec(args)
		return true
	case "goals":
		legacyHint("docs goals", "sceno docs goals [--json]")
		cmdGoals(args)
		return true
	case "shapes":
		legacyHint("docs shapes", "sceno docs shapes")
		cmdDocsShapes()
		return true
	case "icons":
		legacyHint("docs icons", "sceno docs icons")
		cmdDocsIcons()
		return true
	case "suggest":
		legacyHint("advise", "sceno advise -i FILE --json")
		cmdAdvise(args)
		return true
	case "inspect":
		legacyHint("describe", "sceno describe -i FILE --json")
		cmdDescribe(args)
		return true
	default:
		return false
	}
}

func legacyHint(preferred, example string) {
	fmt.Fprintf(os.Stderr, "note: %q is an alias — prefer %s (e.g. %s)\n", os.Args[1], preferred, example)
}

func cmdRender(args []string) {
	fs := flag.NewFlagSet("render", flag.ExitOnError)
	in := fs.String("i", "", "input .kdl spec")
	out := fs.String("o", "", "output path (base name when writing multiple formats)")
	style := fs.String("style", "polished", "sketch or polished")
	format := fs.String("format", "png", "output format(s): png, svg, pdf, html, slides, all (comma-separated for multiple)")
	all := fs.Bool("all", false, "write all formats (svg, png, pdf, html, slides.html)")
	noFix := fs.Bool("no-fix", false, "skip collision resolution")
	jsonErr := fs.Bool("json-errors", false, "on failure print validate JSON to stderr")
	_ = fs.Parse(args)

	if *in == "" {
		fmt.Fprintln(os.Stderr, "render: -i required (.kdl)")
		os.Exit(2)
	}

	result, report, err := validate.LoadAndEvaluate(*in, validate.Options{FixCollisions: !*noFix})
	if !report.OK {
		if *jsonErr {
			_ = report.WriteJSON(os.Stderr)
		} else {
			_ = report.WriteHuman(os.Stderr)
		}
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	deck := result.Deck
	if len(deck.Slides) == 0 {
		fmt.Fprintln(os.Stderr, "no slides in diagram")
		os.Exit(2)
	}

	opt := export.Options{Style: export.RenderStyle(*style), Scale: 2}
	f := strings.ToLower(strings.TrimSpace(*format))
	if *all || f == "all" {
		base := *out
		if base == "" {
			base = "sceno"
		}
		paths, err := export.WriteAllDeck(deck, base, opt)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		for _, p := range paths {
			fmt.Println("wrote", p)
		}
		return
	}
	if *out == "" {
		fmt.Fprintln(os.Stderr, "render: -o required")
		os.Exit(2)
	}

	formats, err := export.ParseFormats(*format)
	if err != nil {
		fmt.Fprintln(os.Stderr, "render:", err)
		os.Exit(2)
	}
	paths, err := export.WriteFormatsDeck(deck, *out, formats, opt)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	for _, p := range paths {
		fmt.Println("wrote", p)
	}
}

func cmdValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	in := fs.String("i", "", "input .kdl spec")
	jsonOut := fs.Bool("json", false, "JSON report for AI (recommended)")
	fix := fs.Bool("fix", true, "resolve node collisions when checking")
	_ = fs.Parse(args)
	if *in == "" {
		fmt.Fprintln(os.Stderr, "validate: -i required (.kdl)")
		os.Exit(2)
	}
	report, _, _ := validate.Run(*in, validate.Options{FixCollisions: *fix})
	if *jsonOut {
		_ = report.WriteJSON(os.Stdout)
	} else {
		_ = report.WriteHuman(os.Stdout)
	}
	os.Exit(report.ExitCode())
}

func cmdAdvise(args []string) {
	fs := flag.NewFlagSet("advise", flag.ExitOnError)
	in := fs.String("i", "", "input .kdl spec")
	jsonOut := fs.Bool("json", false, "JSON output with stack engine + recommendations")
	useAI := fs.Bool("ai", false, "invoke external AI CLI for intelligent review")
	aiCmd := fs.String("ai-cmd", "", "AI command (default: SCENO_AI_CMD env)")
	noFix := fs.Bool("no-fix", false, "skip collision resolution before analysis")
	_ = fs.Parse(args)
	if *in == "" {
		fmt.Fprintln(os.Stderr, "advise: -i required (.kdl)")
		os.Exit(2)
	}
	report, err := advise.Run(*in, advise.Options{
		FixCollisions: !*noFix,
		UseAI:         *useAI,
		AICmd:         *aiCmd,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if *jsonOut {
		_ = report.WriteJSON(os.Stdout)
		return
	}
	_ = report.WriteHuman(os.Stdout)
}

func cmdDescribe(args []string) {
	fs := flag.NewFlagSet("describe", flag.ExitOnError)
	in := fs.String("i", "", "input .kdl spec")
	jsonOut := fs.Bool("json", false, "JSON visual description for AI")
	noFix := fs.Bool("no-fix", false, "skip collision resolution before describe")
	_ = fs.Parse(args)
	if *in == "" {
		fmt.Fprintln(os.Stderr, "describe: -i required (.kdl)")
		os.Exit(2)
	}
	report, err := inspect.Run(*in, inspect.Options{FixCollisions: !*noFix})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if *jsonOut {
		if err := report.WriteJSON(os.Stdout); err != nil {
			os.Exit(2)
		}
		return
	}
	_ = report.WriteHuman(os.Stdout)
}

func cmdDocs(args []string) {
	jsonOut, rest := parseJSONFlag(args)
	topic := ""
	if len(rest) > 0 {
		topic = rest[0]
	}
	if err := docs.Run(topic, jsonOut, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func cmdDocsShapes() {
	if err := docs.Run("shapes", false, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func cmdDocsIcons() {
	if err := docs.Run("icons", false, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

// parseJSONFlag extracts --json anywhere in args (Go flag stops at first positional).
func parseJSONFlag(args []string) (jsonOut bool, rest []string) {
	for _, a := range args {
		if a == "--json" {
			jsonOut = true
			continue
		}
		rest = append(rest, a)
	}
	return jsonOut, rest
}

func cmdGuide(args []string) {
	fs := flag.NewFlagSet("guide", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "machine-readable guide for AI agents")
	_ = fs.Parse(args)
	if *jsonOut {
		if err := guide.JSON(os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		return
	}
	if err := guide.Markdown(os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func cmdSpec(args []string) {
	fs := flag.NewFlagSet("spec", flag.ExitOnError)
	_ = fs.Parse(args)
	fmt.Print(spec.SpecMarkdown())
}

func cmdGoals(args []string) {
	jsonOut, _ := parseJSONFlag(args)
	if jsonOut {
		if err := docs.Run("goals", true, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		return
	}
	fmt.Print(spec.GoalsMarkdown())
}

func cmdVersion(args []string) {
	fs := flag.NewFlagSet("version", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	_ = fs.Parse(args)
	if *jsonOut {
		if err := version.WriteJSON(os.Stdout); err != nil {
			os.Exit(2)
		}
		return
	}
	version.WriteHuman(os.Stdout)
}

func cmdInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	out := fs.String("o", "sceno.kdl", "output spec path")
	_ = fs.Parse(args)
	if !strings.HasSuffix(strings.ToLower(*out), ".kdl") {
		*out += ".kdl"
	}
	tpl := `// Edit this file, then: sceno validate -i sceno.kdl --json
diagram title="My Diagram" subtitle="Optional subtitle" layout=auto style=polished gap=32 padding=24 {

  shape box start "Start" icon=server at=0,0
  shape box end "End" icon=server at=1,0

  edge start -> end fromSide=right toSide=left label="flow"
}
`
	if err := os.WriteFile(*out, []byte(tpl), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Println("wrote", *out)
	fmt.Println("next: sceno validate -i", *out, "--json")
}

func usage() {
	fmt.Fprintf(os.Stderr, "sceno %s — declarative diagrams in KDL (https://kdl.dev)\n\n", version.Version)
	fmt.Fprintf(os.Stderr, `Workflow:  init → validate → advise → describe → render

  sceno init [-o sceno.kdl]        create a starter spec
  sceno validate -i f --json       check spec + layout (run after every edit)
  sceno advise -i f --json         visual rules, score, recommendations
  sceno describe -i f --json       layout feedback without viewing images
  sceno render -i f -o out              export PNG (default)
  sceno render -i f -o out -format svg,pdf   export selected formats
  sceno render -i f -o out --all        export svg, png, pdf, html, slides.html
  sceno docs [TOPIC] [--json]      self-doc: guide, spec, goals, shapes, icons, …
  sceno version [--json]           version, commit, build date

Docs topics: guide, spec, goals, practices, stack, validation, shapes, icons, errors
Agents: start with  sceno docs guide --json

`)
}
