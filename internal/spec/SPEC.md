# Sceno — KDL specification

Diagram is defined in **[KDL](https://kdl.dev/)** only. One format, one mental model.

## Quick start

```kdl
diagram title="My Platform" subtitle="Optional" layout=auto gap=32 padding=24 {

  shape box api "API Gateway" icon=api layer=1
  shape cylinder db "PostgreSQL" icon=database layer=2
  shape actor user "Users" at=0,0

  edge user -> api fromSide=right toSide=left
  edge api -> db
}
```

```bash
sceno init -o platform.kdl
sceno render -i platform.kdl -o out --all
sceno validate -i platform.kdl --json
```

## Document structure

| Statement | Example |
|-----------|---------|
| Diagram block | `diagram title="..." layout=auto { ... }` |
| Shape | `shape box id "Label" layer=1` |
| Edge | `edge from -> to` or `edge from=a to=b` |

Top-level properties (on `diagram` or as props): `title`, `subtitle`, `layout`, `style`, `gap`, `padding`.

## Shapes

Syntax: **`shape KIND ID "Label" props...`**

`KIND` can be omitted (defaults to `box`).

| Kind | Aliases | Use |
|------|---------|-----|
| `box` | `card`, `process` | Default card |
| `ellipse` | `actor`, `circle` | People / roles |
| `diamond` | `decision` | Branching |
| `hexagon` | | External / API |
| `octagon` | | Stop / boundary |
| `cylinder` | `database`, `db` | Data store |
| `cloud` | | Cloud service |
| `document` | `doc` | Document / subprocess |
| `parallelogram` | `input`, `output`, `io` | I/O |
| `triangle` | | Merge / split |
| `pill` | `terminal`, `start`, `end` | Start / end |
| `textbox` | | Light annotation |
| `note` | `sticky` | Sticky note (yellow) |
| `infobox` | `callout` | Accent callout + subtitle |
| `lane` | `container` | Dashed swimlane |
| `frame` | `group` | Solid group |

## Node properties

| Property | Example | Description |
|----------|---------|-------------|
| `icon` | `icon=server` | Icon from catalog |
| `fill` | `fill="#dbeafe"` | Background |
| `stroke` | `stroke="#e2e8f0"` | Border |
| `accent` | `accent="#7c3aed"` | Infobox stripe |
| `subtitle` | `subtitle="..."` | Second line |
| `layer` | `layer=2` | Column (auto layout) |
| `row` | `row=1` | Row in column |
| `at` | `at=2,1` | Shorthand `layer,row` |
| `w`, `h` | `w=200 h=72` | Size override |
| `x`, `y` | `x=100 y=50` | Fixed position (`layout free`) |
| `parent` | `parent=lane1` | Container parent |
| `fontSize` | `fontSize=13` | Label size |

Labels support `\n` for line breaks inside quoted strings.

## Edges

```kdl
edge api -> queue
edge policy -> runner fromSide=top toSide=bottom dashed=true color="#e11d48"
edge a -> b from=right to=left   // same as fromSide/toSide
```

| Property | Values |
|----------|--------|
| `fromSide` / `toSide` | `top` `right` `bottom` `left` `auto` |
| `dashed` | `true` |
| `color` | `#hex` |

## Layout

- `layout auto` — grid by `layer` / `row` (default)
- `layout free` — every shape needs `x` and `y`

## Icons

`api`, `cloud`, `database`, `info`, `k8s`, `lock`, `policy`, `queue`, `server`, `shield`, `storage`, `user`, `users`, `workflow`

## Slides

```kdl
diagram title="Talk" slide=16x9 layout=auto gap=36 {
  slide "Overview" {
    shape box api "API" icon=api at=0,0
  }
}
```

Properties: `slide=16x9` or `slide=4x3` on the diagram line.

### Theme

- `theme=dark` or `theme=light` (default)
- `background=transparent` — no canvas fill (good for PNG/SVG overlays)
- Color overrides: `foreground=#fafafa`, `card=#18181b`, `border=#3f3f46`, `muted=#27272a`, `accent=#a78bfa`
- Custom: `var.card=#112233` (any palette key from `sceno guide --json`)

### Code blocks (slides & diagrams)

```kdl
code snippet lang=go source="package main\nfunc main() {}" at=0,0 w=480 h=160
```

Languages: `go`, `json`, `yaml`, `bash`, `kdl`, `text`. In slide HTML exports, code renders as highlighted `<pre>`; in SVG as colored monospace tspans.

## CLI

```
sceno guide [--json]       # AI handbook: workflow, error codes, examples
sceno init [-o file.kdl]
sceno validate -i f --json # always use --json for agent loops
sceno render -i f -o out [--all]
sceno render -i f -o out.slides.html -format slides
sceno describe -i f --json   # textual layout + ascii map (no image)
sceno suggest -i f [--json]
sceno spec | sceno shapes | sceno icons | sceno goals
```

## Text fitting

- Boxes **auto-size** from embedded Inter metrics (label, subtitle, icon).
- `w` / `h` set a **minimum**; the shape grows if text needs more room.
- `sceno validate` reports `text_overflow` if content still does not fit.

## Validation (`sceno validate --json`)

Each issue includes `code`, `message`, `fix`, and often `example` (KDL snippet).

| Code | Severity |
|------|----------|
| `parse_error` | error |
| `missing_node` | error |
| `missing_position` | error |
| `collision` | error |
| `edge_collision` | error if through node; warning if crossing |
| `text_overflow` | error |
| `unknown_icon` | error |
| `suggest_compact` | warning |

Response also includes `agent.next_steps` and `agent.summary` when using `--json`.
