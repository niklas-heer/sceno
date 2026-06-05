# Sceno — KDL specification

Diagram is defined in **[KDL](https://kdl.dev/)** only. One format, one mental model.

## Quick start

```kdl
diagram title="My Platform" subtitle="Optional" layout=auto gap=32 padding=24 {

  shape box api "API Gateway" icon=api iconPos=top-left layer=1
  shape cylinder db "PostgreSQL" icon=database layer=2
  shape info note "Context" icon=info subtitle="Read left to right" at=0,2

  edge api -> db fromSide=right toSide=left label="SQL"
}
```

```bash
sceno init -o platform.kdl
sceno validate -i platform.kdl --json
sceno advise -i platform.kdl --json
sceno render -i platform.kdl -o out --all
```

## Document structure

| Statement | Example |
|-----------|---------|
| Diagram block | `diagram title="..." layout=auto { ... }` |
| Shape | `shape box id "Label" layer=1` |
| Edge | `edge from -> to` or `edge from=a to=b` |
| Slide | `slide "Title" { ... }` inside diagram |

Top-level properties (on `diagram` or as props): `title`, `subtitle`, `layout`, `style`, `gap`, `padding`, `slide`, `theme`, `background`.

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
| `note` | `sticky`, `postit` | Sticky note (yellow) |
| `infobox` | `callout` | Accent callout + subtitle |
| `info` | | Blue infobox (default accent) |
| `warning` | `warn` | Amber infobox |
| `tip` | `hint` | Green infobox |
| `lane` | `container` | Dashed swimlane |
| `frame` | `group` | Solid group |
| `code` | `codeblock` | Syntax-highlighted block |

## Node properties

| Property | Example | Description |
|----------|---------|-------------|
| `icon` | `icon=server` | Icon from catalog |
| `iconPos` | `iconPos=top-left` | Icon placement (see below) |
| `fill` | `fill="#dbeafe"` | Background |
| `stroke` | `stroke="#e2e8f0"` | Border |
| `accent` | `accent="#7c3aed"` | Infobox left stripe |
| `subtitle` | `subtitle="..."` | Second line |
| `layer` | `layer=2` | Column (auto layout) |
| `row` | `row=1` | Row in column |
| `at` | `at=2,1` | Shorthand `layer,row` |
| `w`, `h` | `w=200 h=72` | Minimum size (auto-expands for text) |
| `x`, `y` | `x=100 y=50` | Fixed position (`layout free`) |
| `parent` | `parent=lane1` | Container parent |
| `fontSize` | `fontSize=13` | Label size |
| `lang`, `source` | on `code` shapes | Code language and body |

### Icon placement (`iconPos`)

| Value | Position |
|-------|----------|
| `top-left` | Default — icon top-left, label below/right |
| `top` | Centered on top edge |
| `top-right` | Top-right corner |
| `center` | Center of node |
| `bottom-left`, `bottom`, `bottom-right` | Bottom positions |

Labels support `\n` for line breaks inside quoted strings.

## Edges

```kdl
edge api -> queue
edge api -> queue "async jobs"
edge policy -> runner fromSide=top toSide=bottom dashed=true color="#e11d48" label="enforce"
edge a -> b from=right to=left   // same as fromSide/toSide
```

| Property | Values |
|----------|--------|
| `label` | Text on the arrow (quoted string after `->` or `label="..."`) |
| `fromSide` / `toSide` | `top` `right` `bottom` `left` |
| `dashed` | `true` |
| `color` | `#hex` |

Edge labels render above horizontal segments and to the right of vertical segments.

## Layout

- `layout auto` — grid by `layer` / `row` / `at=col,row` (default)
- `layout free` — every shape needs `x` and `y`

Single-row diagrams (all nodes share one `row`) vertically center within the row band so horizontal pipelines stay straight.

## Icons

Run `sceno icons` or `sceno docs icons --json` for the full catalog.

Common: `api`, `cloud`, `database`, `info`, `k8s`, `lock`, `policy`, `queue`, `server`, `shield`, `storage`, `user`, `users`, `workflow`

## Slides

```kdl
diagram title="Talk" slide=16x9 theme=dark layout=auto gap=36 {

  slide "Overview" {
    shape info summary "Key point" icon=info at=0,0
  }

  slide "Architecture" {
    shape box api "API" icon=api layer=1
    shape box db "DB" icon=database layer=2
    edge api -> db
  }
}
```

Properties: `slide=16x9` or `slide=4x3` on the diagram line.

### Theme

- `theme=dark` or `theme=light` (default)
- `background=transparent` — no canvas fill (good for PNG/SVG overlays)
- Color overrides: `foreground`, `card`, `border`, `muted`, `accent`
- Custom: `var.card=#112233`

### Code blocks

```kdl
shape code snippet lang=go source="package main\nfunc main() {}" at=0,0 w=480 h=160
```

Languages: `go`, `json`, `yaml`, `bash`, `kdl`, `text`.

## Stack validation

Sceno validates diagrams as **stacked 2D planes** (background → lanes → edges → structure → annotations → nodes → labels → chrome). See `sceno docs stack` for the full model.

```bash
sceno advise -i file.kdl --json    # visual score + stack + rules
sceno describe -i file.kdl --json  # includes scene.stack and engine
sceno validate -i file.kdl --json  # errors + stack rule warnings
```

Optional AI review: `SCENO_AI_CMD="codex exec -" sceno advise -i file.kdl --ai`

## CLI

```
sceno docs [TOPIC] [--json]   # guide, spec, goals, practices, stack, errors, shapes, icons
sceno guide [--json]          # agent handbook (alias)
sceno init [-o file.kdl]
sceno validate -i f --json    # always use --json for agent loops
sceno advise -i f --json      # stack engine + visual rules + recommendations
sceno suggest -i f --json     # prioritized layout hints
sceno describe -i f --json    # layout + ascii map + engine (no image)
sceno render -i f -o out [--all]
sceno render -i f -o out.slides.html -format slides
sceno spec | sceno goals | sceno shapes | sceno icons | sceno version
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
| `occluded`, `edge_hidden`, `misaligned` | warning |
| `dense_layout`, `slide_crowded`, `too_many_elements`, `annotation_blocks` | warning |
| `suggest_compact`, `sparse_layout`, `weak_hierarchy`, `suggest_annotation` | hint |

Response also includes `recommendations`, `agent.next_steps`, and `agent.summary` when using `--json`.
