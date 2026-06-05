# Sceno — instructions for AI agents

Use this tool to produce architecture diagrams from a single **KDL** (`.kdl`) file. Outputs: SVG, PNG, PDF, HTML, and slide decks.

## Start here

```bash
sceno docs guide --json
```

Browse all documentation topics:

```bash
sceno docs --json
```

Key topics (all **generated from code** at runtime — no separate markdown to maintain):

```bash
sceno docs stack --json       # stacked-plane validation model + visual rules
sceno docs validation --json  # validate + advise commands, error codes
sceno docs goals              # product goals + quality bar
sceno docs practices --json   # workflow + best practices
sceno docs spec               # full KDL specification
sceno docs errors --json      # error code repair catalog
```

## Edit loop (every change)

```bash
sceno validate -i your.kdl --json
```

Do **not** render until `ok` is true.

Visual quality check (no image needed):

```bash
sceno advise -i your.kdl --json   # stack engine, visual score, recommendations
sceno describe -i your.kdl --json  # positions, ascii map, scene, engine
sceno render -i your.kdl -o output/sceno --all
```

## Stack validation model

Diagrams are validated as **stacked 2D planes** (back → front):

`background → lanes → edges → structure → annotations → nodes → labels → chrome`

Collision and routing checks project onto reduced planes. Full details: `sceno docs stack --json`.

## Describe & advise output

`sceno describe --json` returns:

- `slides[n].narrative` — what the slide communicates
- `slides[n].scene` — 2D analysis (paint order, occlusion, edge visibility, `stack`)
- `slides[n].engine` — stack engine findings, visual score, rules run
- `slides[n].ascii_map` — coarse spatial grid
- `slides[n].visual_problems` — overlaps, hidden edges, misalignment
- `slides[n].edges[].route` — step-by-step connector path

`sceno advise --json` returns:

- `visual_score` — 0–100 quality score
- `stack` — plane counts
- `engine.findings` — visual design rule outcomes with `fix` hints
- `recommendations` — prioritized actionable hints
- `ai_review` — when `--ai` and `SCENO_AI_CMD` are set

`sceno validate --json` also warns on stack rules: `edge_hidden`, `occluded`, `misaligned`, `dense_layout`, `slide_crowded`, etc.

## Rules

1. **KDL only** — `.kdl` files; root block is `diagram { }` in the spec language.
2. **Validate after every edit** — `sceno validate -i file.kdl --json`.
3. **Advise for polish** — `sceno advise -i file.kdl --json` after validate passes.
4. **Read `agent.next_steps`** when `ok` is false; apply `errors[].fix` and `errors[].example`.
5. **Define shapes before edges** in the same `diagram { }` or `slide "Title" { }` block.
6. **Do not invent** shape kinds or icon names — use lists from `sceno docs guide --json`.
7. **Prefer `layout=auto`** with `layer`, `row`, or `at=col,row`; use `layout=free` + `x`/`y` for free placement.
8. **Quote labels with spaces** — `title="My Platform"`.
9. **Use `\n` in quoted strings** for line breaks inside labels.
10. **Callouts** — `shape info`, `tip`, `warning`, `infobox`, `note` for annotations; `iconPos=top-left` for icons.

## Commands

| Command | Purpose |
|---------|---------|
| `sceno init -o sceno.kdl` | Starter file |
| `sceno validate -i f --json` | Validate + repair hints + stack warnings |
| `sceno advise -i f --json` | Stack engine + visual score + recommendations |
| `sceno advise -i f --ai` | Optional external AI CLI review (`SCENO_AI_CMD`) |
| `sceno describe -i f --json` | Layout without images (includes engine) |
| `sceno render -i f -o out --all` | Export everything |
| `sceno render -format slides` | HTML presentation |
| `sceno docs [TOPIC] [--json]` | **Self-doc hub** — guide, spec, goals, shapes, icons, stack, errors, … |
| `sceno docs guide --json` | Full agent handbook (start here) |
| `sceno version [--json]` | Tool version |

## Error codes

Full fixes and examples: `sceno docs errors --json` or `sceno guide --json` → `error_codes`.

## Examples in repo

- `examples/how-it-works.kdl` — README pipeline diagram
- `examples/self-service.kdl` — full platform diagram
- `examples/slides-demo.kdl` — slide deck
- `examples/slides-dark.kdl` — dark theme + code
- `examples/shapes-demo.kdl` — shape gallery
