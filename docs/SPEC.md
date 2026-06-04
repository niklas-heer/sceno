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

## CLI

```
sceno init [-o file.kdl]
sceno render -i spec.kdl -o out [--all]
sceno validate -i spec.kdl [--json]
sceno suggest -i spec.kdl
sceno spec
sceno shapes
sceno icons
```

## Validation

- **Errors**: parse issues, unknown shape/icon, node overlap, edge through node
- **Warnings**: edge crossings, sparse layout suggestions
