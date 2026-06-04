# Sceno

Declarative architecture diagrams in **[KDL](https://kdl.dev/)** — one readable spec, polished SVG/PNG/PDF/HTML/slide exports. Built for humans and **AI agents** that iterate until the spec validates.

## For AI agents

**Start every session with:**

```bash
sceno guide --json
```

**After every KDL edit:**

```bash
sceno validate -i sceno.kdl --json
```

The JSON report includes `ok`, `errors` (with `fix` + `example`), `warnings`, and `agent.next_steps`. Only render when `ok` is true.

See [AGENTS.md](AGENTS.md) for the full agent playbook.

## Why KDL?

- **Readable** — `edge api -> queue`, `shape actor devs "Developers"`, `at=1,2`
- **Declarative** — like Mermaid/d2, but with PowerPoint-familiar shapes, icons, and slides
- **Single format** — the CLI only accepts `.kdl` (no YAML/JSON drift)
- **Self-documenting** — `sceno guide`, `sceno spec`, `sceno goals`, structured validation
- **Agent-friendly** — `validate` + `describe` (2D scene, ASCII map) without viewing images
- **Themed slides** — `theme=dark`, `background=transparent`, syntax-highlighted `code` blocks

Run `sceno goals` for the full product goals and ecosystem best practices.

## Install

### macOS (Apple Silicon or Intel)

Download the latest release for your Mac from [GitHub Releases](https://github.com/niklas-heer/sceno/releases), or install with Go:

```bash
go install github.com/niklas-heer/sceno/cmd/sceno@latest
```

Ensure `$(go env GOPATH)/bin` is on your `PATH`.

### Linux

```bash
go install github.com/niklas-heer/sceno/cmd/sceno@latest
```

Or download `sceno_linux_arm64.tar.gz` / `sceno_linux_amd64.tar.gz` from [Releases](https://github.com/niklas-heer/sceno/releases).

### Build from source

```bash
git clone https://github.com/niklas-heer/sceno.git
cd sceno
make build    # produces ./sceno
make install  # go install ./cmd/sceno
```

Requires **Go 1.23+**.

## Commands

| Command | Description |
|---------|-------------|
| `sceno guide [--json]` | **Agent handbook** — workflow, error codes, examples, shapes, icons |
| `sceno init [-o sceno.kdl]` | Create a starter spec |
| `sceno validate -i f --json` | Validate + actionable repair hints |
| `sceno check -i f --json` | Alias for `validate` |
| `sceno render -i f -o out --all` | Export svg, png, pdf, html, slides.html |
| `sceno render -format slides` | 16:9 HTML presentation |
| `sceno describe -i f --json` | **Visual feedback without images** — narrative, ASCII map, positions, problems |
| `sceno feedback` | Alias for `describe` |
| `sceno suggest -i f [--json]` | Layout warnings (sparse columns, etc.) |
| `sceno spec` | Full KDL specification |
| `sceno goals` | Product goals and quality bar |
| `sceno shapes` / `sceno icons` | Quick lists |

## Quick start

```bash
sceno init -o platform.kdl
# edit platform.kdl
sceno validate -i platform.kdl --json
sceno render -i platform.kdl -o output/sceno --all
```

## Spec example

```kdl
diagram title="My Platform" layout=auto gap=32 padding=24 {

  shape box api "API Gateway" icon=api layer=1
  shape cylinder db "Database" icon=database layer=2
  shape actor ops "Operators" at=0,0

  edge ops -> api fromSide=right toSide=left
  edge api -> db
}
```

> The root block keyword in KDL specs is `diagram { }` — that is the file format, not the CLI name.

## Describe layout (no vision required)

Agents that cannot view PNG/SVG can still sanity-check layout:

```bash
sceno describe -i examples/self-service.kdl --json
```

Example fields:

- `slides[0].narrative` — plain-language overview + scene summary
- `slides[0].scene` — paint order, groups, occlusion, edge visibility, aesthetic score
- `slides[0].ascii_map` — coarse character grid of node positions and edge paths
- `slides[0].visual_problems` — overlaps, hidden edges, misalignment (not layout hints)
- `slides[0].edges[].route` — `from (x,y), down 120px then right 200px, to (x,y)`

## Validation (AI-ready)

`sceno validate --json` returns machine-readable issues:

```json
{
  "ok": false,
  "errors": [
    {
      "code": "missing_node",
      "message": "edge references unknown node \"queue\"",
      "fix": "Add: shape box queue \"Label\" before the edge.",
      "example": "shape box queue \"queue\"\nedge api -> queue"
    }
  ],
  "agent": {
    "summary": "1 error(s) — fix errors before render.",
    "next_steps": ["Fix error 1 ...", "Run: sceno validate -i ..."]
  }
}
```

| Code | Blocks render? |
|------|----------------|
| `parse_error`, `missing_node`, `collision`, `text_overflow`, … | Yes |
| `edge_collision` (through node) | Yes |
| `edge_collision` (crossing) | Warning only |
| `suggest_compact` | Warning only |

## Theme & code (slides)

```kdl
diagram title="Talk" theme=dark background=transparent slide=16x9 layout=auto gap=36 {
  slide "Snippet" {
    code main lang=go source="package main\nfunc main() {}" at=0,0 w=480 h=140
  }
}
```

Override colors: `foreground=#fafafa`, `card=#18181b`, or `var.border=#3f3f46`.

## Slides (declarative decks)

```kdl
diagram title="Talk" slide=16x9 layout=auto gap=36 {
  slide "Problem" {
    shape callout note "Pain point" icon=info at=0,0
  }
  slide "Solution" {
    shape box api "API" icon=api layer=1
    shape box db "DB" icon=database layer=2
    edge api -> db
  }
}
```

```bash
sceno render -i examples/slides-demo.kdl -o output/talk.slides.html -format slides
```

Open `.slides.html` in a browser — **← / → / Space** to navigate. Use `--all` to also get `sceno.slides.html` alongside svg/png/pdf when `-o output/sceno`.

## Shapes & icons

Run `sceno shapes` and `sceno icons`, or see `examples/shapes-demo.kdl`.

Highlights: `box`, `actor`, `cylinder`, `cloud`, `callout`, `lane`, `hexagon`, `note`, …

## Export formats

| Format | Use |
|--------|-----|
| SVG | Reference vector (rounded connectors, embedded Inter) |
| PNG | Rasterized from SVG |
| PDF | Vector + Inter |
| HTML | Self-contained page (shadcn/zinc styling) |
| slides | 16:9 HTML deck for presentations |

## Examples

| File | Description |
|------|-------------|
| [examples/self-service.kdl](examples/self-service.kdl) | Platform architecture |
| [examples/slides-demo.kdl](examples/slides-demo.kdl) | Three-slide deck |
| [examples/slides-dark.kdl](examples/slides-dark.kdl) | Dark theme + Go code slide |
| [examples/shapes-demo.kdl](examples/shapes-demo.kdl) | Shape gallery |

## Goals

```bash
sceno goals
```

## Development

```bash
make test
make build
```

## License

MIT — see [LICENSE](LICENSE).
