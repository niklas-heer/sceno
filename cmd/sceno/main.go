package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/niklas-heer/sceno/internal/export"
	"github.com/niklas-heer/sceno/internal/guide"
	"github.com/niklas-heer/sceno/internal/inspect"
	"github.com/niklas-heer/sceno/internal/model"
	"github.com/niklas-heer/sceno/internal/pipeline"
	"github.com/niklas-heer/sceno/internal/spec"
	"github.com/niklas-heer/sceno/internal/validate"
)

// version is set at link time (Makefile, CI, release workflow).
var version = "dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "render":
		cmdRender(os.Args[2:])
	case "validate", "check":
		cmdValidate(os.Args[2:])
	case "suggest":
		cmdSuggest(os.Args[2:])
	case "guide", "agent":
		cmdGuide(os.Args[2:])
	case "describe", "feedback", "inspect":
		cmdDescribe(os.Args[2:])
	case "icons":
		cmdIcons()
	case "shapes":
		cmdShapes()
	case "init":
		cmdInit(os.Args[2:])
	case "spec":
		cmdSpec(os.Args[2:])
	case "goals":
		cmdGoals()
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func cmdRender(args []string) {
	fs := flag.NewFlagSet("render", flag.ExitOnError)
	in := fs.String("i", "", "input .kdl spec")
	out := fs.String("o", "", "output path")
	style := fs.String("style", "polished", "sketch or polished")
	format := fs.String("format", "svg", "svg|png|pdf|html|slides|all")
	all := fs.Bool("all", false, "write all formats")
	noFix := fs.Bool("no-fix", false, "skip collision resolution")
	jsonErr := fs.Bool("json-errors", false, "on failure print validate JSON to stderr")
	_ = fs.Parse(args)

	if *in == "" {
		fmt.Fprintln(os.Stderr, "render: -i required (.kdl)")
		os.Exit(2)
	}

	report, _, err := validate.Run(*in, validate.Options{FixCollisions: !*noFix})
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

	s, err := spec.LoadFile(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	popt := pipeline.DefaultOptions()
	popt.ResolveCollision = !*noFix
	deck, _, err := pipeline.BuildDeck(s, popt)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if len(deck.Slides) == 0 {
		fmt.Fprintln(os.Stderr, "no slides in diagram")
		os.Exit(2)
	}

	opt := export.Options{Style: export.RenderStyle(*style), Scale: 2}
	f := strings.ToLower(*format)
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
	if f == "slides" || f == "slide" {
		if err := export.WriteDeck(deck, *out, export.FormatSlides, opt); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	} else if len(deck.Slides) == 1 {
		if err := export.Write(deck.Slides[0], *out, export.Format(f), opt); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	} else {
		if err := export.WriteDeck(deck, *out, export.Format(f), opt); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}
	fmt.Println("wrote", *out)
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

func cmdSuggest(args []string) {
	fs := flag.NewFlagSet("suggest", flag.ExitOnError)
	in := fs.String("i", "", "input .kdl spec")
	jsonOut := fs.Bool("json", false, "JSON output")
	_ = fs.Parse(args)
	if *in == "" {
		os.Exit(2)
	}
	report, _, _ := validate.Run(*in, validate.Options{FixCollisions: true})
	if *jsonOut {
		out := struct {
			OK       bool        `json:"ok"`
			Warnings interface{} `json:"warnings"`
		}{OK: report.OK, Warnings: report.Warnings}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
		return
	}
	if len(report.Warnings) == 0 {
		fmt.Println("no suggestions")
		return
	}
	for _, w := range report.Warnings {
		fmt.Println("suggest:", w.Message)
		if w.Fix != "" {
			fmt.Println("  fix:", w.Fix)
		}
	}
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
	fmt.Print(spec.SpecMarkdown)
}

func cmdGoals() {
	fmt.Print(spec.GoalsMarkdown)
}

func cmdShapes() {
	fmt.Println("Shapes (use as: shape KIND id \"Label\" ...):")
	for _, s := range model.AllShapes() {
		fmt.Println(" ", s)
	}
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

  edge start -> end fromSide=right toSide=left
}
`
	if err := os.WriteFile(*out, []byte(tpl), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Println("wrote", *out)
	fmt.Println("next: sceno validate -i", *out, "--json")
}

func cmdIcons() {
	fmt.Println("Icons (use as: icon=name):")
	for _, name := range []string{"api", "cloud", "database", "info", "k8s", "lock", "policy", "queue", "server", "shield", "storage", "user", "users", "workflow"} {
		fmt.Println(" ", name)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "sceno %s — declarative diagrams in KDL (https://kdl.dev)\n\n", version)
	fmt.Fprintf(os.Stderr, `For AI agents: start with  sceno guide --json

  sceno guide [--json]          self-doc: workflow, error codes, examples
  sceno init [-o sceno.kdl]   starter spec
  sceno validate|check -i f --json   validate + repair hints (use every edit)
  sceno render -i f -o out [--all]
  sceno render -i f -o deck.slides.html -format slides
  sceno describe|feedback -i f --json   how it looks (no image needed)
  sceno suggest -i f [--json]   layout warnings only
  sceno spec                    full KDL specification
  sceno goals                   product goals
  sceno shapes | sceno icons

`)
}
