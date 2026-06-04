# Sceno — instructions for AI agents

Use this tool to produce architecture diagrams from a single **KDL** (`.kdl`) file. Outputs: SVG, PNG, PDF, HTML, and slide decks.

## Start here

```bash
sceno docs guide --json
```

Or browse all documentation topics:

```bash
sceno docs --json
```

Optional context:

```bash
sceno docs goals        # product goals + quality bar
sceno docs practices --json   # workflow + best practices
sceno docs spec         # full KDL specification
sceno docs errors --json      # error code repair catalog
```

## Edit loop (every change)

```bash
sceno validate -i your.kdl --json
```

Do **not** render until `ok` is true.

Optional layout check (no image needed):

```bash
sceno describe -i your.kdl --json   # positions, ascii map, scene, problems
sceno render -i your.kdl -o output/sceno --all
```

## Describe output

`sceno describe --json` returns:

- `slides[n].narrative` — what the slide communicates
- `slides[n].scene` — 2D analysis (paint order, occlusion, edge visibility)
- `slides[n].ascii_map` — coarse spatial grid
- `slides[n].visual_problems` — overlaps, hidden edges, misalignment
- `slides[n].edges[].route` — step-by-step connector path

`sceno validate --json` also warns on `edge_hidden`, `occluded`, and `misaligned` via the same scene model.

Use after validate passes to confirm the diagram reads well before shipping PNG/SVG.

## Rules

1. **KDL only** — `.kdl` files; root block is `diagram { }` in the spec language.
2. **Validate after every edit** — `sceno validate -i file.kdl --json`.
3. **Read `agent.next_steps`** when `ok` is false; apply `errors[].fix` and `errors[].example`.
4. **Define shapes before edges** in the same `diagram { }` or `slide "Title" { }` block.
5. **Do not invent** shape kinds or icon names — use lists from `sceno guide --json`.
6. **Prefer `layout=auto`** with `layer`, `row`, or `at=col,row` unless you need exact `x`/`y`.
7. **Quote labels with spaces** — `title="My Platform"`.
8. **Use `\n` in quoted strings** for line breaks inside labels.

## Commands

| Command | Purpose |
|---------|---------|
| `sceno docs [TOPIC] [--json]` | **Self-doc hub** — guide, spec, goals, practices, errors, shapes, icons |
| `sceno docs guide --json` | Full agent handbook (start here) |
| `sceno guide [--json]` | Alias for `docs guide` |
| `sceno init -o sceno.kdl` | Starter file |
| `sceno validate -i f --json` | Validate + repair hints + recommendations |
| `sceno suggest -i f --json` | Prioritized layout recommendations |
| `sceno render -i f -o out --all` | Export everything |
| `sceno render -format slides` | HTML presentation |
| `sceno describe -i f --json` | Layout without images |
| `sceno spec` | KDL spec (alias: `docs spec`) |
| `sceno goals` | Product goals (alias: `docs goals`) |
| `sceno version [--json]` | Tool version |

## Error codes

Full fixes and examples: `sceno guide --json` → `error_codes`.

## Examples in repo

- `examples/self-service.kdl` — full platform diagram
- `examples/slides-demo.kdl` — slide deck
- `examples/slides-dark.kdl` — dark theme + code
- `examples/shapes-demo.kdl` — shape gallery
