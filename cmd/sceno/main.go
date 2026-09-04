package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/niklas-heer/sceno/internal/advise"
	"github.com/niklas-heer/sceno/internal/docs"
	"github.com/niklas-heer/sceno/internal/export"
	"github.com/niklas-heer/sceno/internal/guide"
	"github.com/niklas-heer/sceno/internal/inspect"
	"github.com/niklas-heer/sceno/internal/preview"
	"github.com/niklas-heer/sceno/internal/spec"
	"github.com/niklas-heer/sceno/internal/starter"
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
	case "preview":
		cmdPreview(args)
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
	if err := runInit(args, os.Stdout); err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func runInit(args []string, output io.Writer) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(output)
	path := fs.String("o", "sceno.kdl", "output spec path")
	name := fs.String("template", "service-architecture", "starter template name (use --list to browse)")
	list := fs.Bool("list", false, "list available starter templates")
	jsonOut := fs.Bool("json", false, "JSON template metadata or creation result")
	force := fs.Bool("force", false, "overwrite an existing output file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("init: unexpected argument %q; use -o for the output path", fs.Arg(0))
	}
	if *list {
		if *jsonOut {
			enc := json.NewEncoder(output)
			enc.SetIndent("", "  ")
			return enc.Encode(starter.List())
		}
		for _, template := range starter.List() {
			if _, err := fmt.Fprintf(output, "%s — %s\n  %s\n", template.Name, template.Title, template.Description); err != nil {
				return err
			}
		}
		return nil
	}
	source, err := starter.Source(*name)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*path) == "" {
		return fmt.Errorf("init: output path must not be empty")
	}
	if !strings.HasSuffix(strings.ToLower(*path), ".kdl") {
		*path += ".kdl"
	}
	if err := os.MkdirAll(filepath.Dir(*path), 0o755); err != nil {
		return err
	}
	mode := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if *force {
		mode = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	f, err := os.OpenFile(*path, mode, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("init: %s already exists; use --force to overwrite it", *path)
		}
		return err
	}
	_, writeErr := io.WriteString(f, source)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	quotedPath := "'" + strings.ReplaceAll(*path, "'", "'\"'\"'") + "'"
	next := []string{"sceno preview " + quotedPath, "sceno validate -i " + quotedPath + " --json"}
	if *jsonOut {
		canonical := strings.TrimSpace(*name)
		if canonical == "" || canonical == "default" {
			canonical = "service-architecture"
		}
		var selected starter.Template
		for _, template := range starter.List() {
			if template.Name == canonical {
				selected = template
				break
			}
		}
		enc := json.NewEncoder(output)
		enc.SetIndent("", "  ")
		return enc.Encode(struct {
			Path      string           `json:"path"`
			Template  starter.Template `json:"template"`
			NextSteps []string         `json:"next_steps"`
		}{*path, selected, next})
	}
	_, err = fmt.Fprintf(output, "wrote %s\nnext: %s\nthen: %s\n", *path, next[0], next[1])
	return err
}

func cmdPreview(args []string) {
	input, port, openBrowser, err := parsePreviewArgs(args, os.Stdout)
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := preview.Serve(ctx, input, port, openBrowser); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func parsePreviewArgs(args []string, output io.Writer) (string, int, bool, error) {
	fs := flag.NewFlagSet("preview", flag.ContinueOnError)
	fs.SetOutput(output)
	input := fs.String("i", "", "input .kdl file (or pass its path as a positional argument)")
	port := fs.Int("port", 0, "local server port (0 chooses an available port)")
	noOpen := fs.Bool("no-open", false, "print the preview URL without opening a browser")
	// Accept `preview diagram.kdl --no-open` as well as flags before the path.
	var flags, positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			positional = append(positional, arg)
			continue
		}
		flags = append(flags, arg)
		if arg == "-i" || arg == "--i" || arg == "-port" || arg == "--port" {
			if i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
		}
	}
	if err := fs.Parse(flags); err != nil {
		return "", 0, false, err
	}
	if len(positional) > 1 || len(positional) == 1 && *input != "" {
		return "", 0, false, fmt.Errorf("preview: specify one input file, using -i or a positional path")
	}
	if len(positional) == 1 {
		*input = positional[0]
	}
	if strings.TrimSpace(*input) == "" {
		return "", 0, false, fmt.Errorf("preview: input .kdl file required; use sceno preview diagram.kdl")
	}
	if *port < 0 || *port > 65535 {
		return "", 0, false, fmt.Errorf("preview: port must be between 0 and 65535")
	}
	return *input, *port, !*noOpen, nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "sceno %s — declarative diagrams in KDL (https://kdl.dev)\n\n", version.Version)
	fmt.Fprintf(os.Stderr, `Workflow:  init → validate → advise → describe → render

  sceno init [-o sceno.kdl] [--template NAME]   create a starter (use --force to replace)
  sceno init --list [--json]       browse starter templates
  sceno preview file.kdl [--port N] [--no-open]   edit and preview locally
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
