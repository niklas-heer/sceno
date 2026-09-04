# Sceno — instructions for AI agents

Use this tool to produce architecture diagrams from a single **KDL** (`.kdl`) file. Outputs: SVG, PNG, PDF, HTML, and slide decks.

## Core objective

Sceno bridges a text-only agent and a visual canvas. The same computed geometry must drive both exported pixels and machine-readable feedback, so an agent can understand what exists, where it is, what overlaps, and how to repair it without guessing from an image.

- Auto layout should be collision-free by default; hybrid/free placement and `overlap=allow` provide deliberate breakout control.
- `validate`, `advise`, and `describe` must expose exact bounds, routes, stack planes, findings, and actionable repairs.
- SVG, PNG, PDF, HTML, and slides must preserve the same semantic shapes, icons, connectors, and paint order.
- Repository examples are the visual regression corpus: every example must score at least 80 and have no structural visual findings.

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
sceno docs architecture --json  # geometry vs semantics SoT, pipeline.Result, paint order
sceno docs visual --json        # measurable composition, spacing, icon, and connector rules
sceno docs stack --json         # stacked-plane validation model + visual rules
sceno docs validation --json    # validate + advise commands, error codes
sceno docs goals --json        # product goals + quality bar
sceno docs practices --json    # workflow + best practices
sceno docs spec --json         # full KDL specification
sceno docs errors --json       # error code repair catalog
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
sceno render -i your.kdl -o output/sceno
```

For interactive work, run `sceno init --list [--json]` to choose a validated starter, then `sceno preview file.kdl`. The local preview supports source editing, measured geometry inspection, and **Review a repair → Apply repair → Undo change**. Repairs must match current engine suggestions and resolve the targeted issue without structural regressions; they never silently change labels, remove shapes, or allow overlap. Other findings can remain after a successful repair. External edits invalidate pending repairs and session Undo history; browser drafts are preserved for conflict resolution.

The preview always uses polished rendering. Invalid source retains a visibly stale last-valid view with geometry inspection disabled, and export waits until structural findings are resolved. Preview SVG/PNG downloads use the selected slide; PDF/slides include the complete deck.

## Stack validation model

Diagrams are validated as **stacked 2D planes** (back → front):

`background → lanes → structure → edges → annotations → nodes → labels → chrome`

Collision and routing checks project onto reduced planes. Full details: `sceno docs stack --json`.

## Describe & advise output

`sceno describe --json` returns:

- `slides[n].narrative` — what the slide communicates
- `slides[n].scene` — 2D analysis (paint order, occlusion, edge visibility, `stack`)
- `slides[n].engine` — stack engine findings, visual score, rules run
- `slides[n].ascii_map` — coarse spatial grid
- `slides[n].visual_problems` — overlaps, hidden/detached arrows, arrow clusters, label/chrome collisions, text overflow, exact geometry, repair candidates
- `slides[n].edges[].route` — step-by-step connector path

`slides[n].engine.scene_stack.planes.*[]` contains exact outer `bounds`, visible `outline`, `internal_lines`, silhouette-safe `writable_bounds`, selected `effective_font_size`, and icon/title/subtitle `content` boxes; `order` is source order within a semantic plane. Text is fitted to that writable region with a 10px readability floor, and `text_overflow` exposes shape/writable/content geometry plus a size repair when it cannot fit.

`slides[n].engine.spacing` reports measured node gaps, content padding, parent insets, and canvas margins in pixels. Clearance uses axis-aligned bounds; negative padding means overflow. Intentional overlap is measured and marked exempt. Check `truncated` before treating the bounded pair list as complete.

`sceno advise --json` returns:

- `visual_score` — 0–100 quality score
- `stack` — plane counts
- `engine.findings` — visual design rule outcomes with `fix`, geometry, and repair candidates
- `recommendations` — prioritized actionable hints
- `ai_review` — when `--ai` and `SCENO_AI_CMD` are set

Advise also retains every slide's full engine in `slides[]`. Indices are 1-based; merged findings carry `slide_index`, and `engine_slide_index` identifies the lowest-score slide supplying top-level geometry. Scope repair targets to their slide because IDs may repeat across slides.

`sceno validate --json` also warns on stack rules: `edge_hidden`, `arrow_detached`, `arrow_hidden`, `arrow_cluster`, `edge_label_chrome_overlap`, `edge_label_overlap`, `text_overflow`, `occluded`, `misaligned`, `dense_layout`, `slide_crowded`, etc. **Arrow checks** use the same math as render: the path ends on the target border, the tip stays within 2px of its anchor, and the target approach reserves a straight 27px run (18px visible shaft + 9px head). If validate passes, arrowheads should meet shapes in export without hooks or stacked tips.

## Rules

1. **KDL only** — `.kdl` files; root block is `diagram { }` in the spec language.
2. **Validate after every edit** — `sceno validate -i file.kdl --json`.
3. **Advise for polish** — `sceno advise -i file.kdl --json` after validate passes.
4. **Read `agent.next_steps`** when `ok` is false; for collisions inspect `geometry` and try one `repairs[]` edit.
5. **Define shapes before edges** in the same `diagram { }` or `slide "Title" { }` block.
6. **Do not invent** shape kinds or icon names — use lists from `sceno docs guide --json`.
7. **Prefer `layout=auto`** with `layer`, `row`, or `at=col,row`; use `dx`/`dy` for small nudges, `layout=hybrid` + `x`/`y` for breakout elements, and `overlap=allow` only intentionally.
8. **Quote labels with spaces** — `title="My Platform"`.
9. **Use `\n` in quoted strings** for line breaks inside labels.
10. **Callouts** — `shape info`, `tip`, `warning`, `infobox`, `note` for annotations; `iconPos=top-left` for icons.
11. **Repository changes** — run `mask verify`; the full KDL corpus must keep score ≥80 with no collision, detour, hidden/detached arrow, label overlap, side mismatch, occlusion, or text overflow.
12. **Bound numeric input** — grid indices are integers from 0 through 10,000; geometry must be finite with absolute magnitude ≤1,000,000. Polished PNG export is capped at 32 million pixels and 32,768 pixels per side; use SVG/PDF or smaller slides when necessary.

## Commands

| Command | Purpose |
|---------|---------|
| `sceno init -o sceno.kdl` | Starter file; `--list` lists templates and `--template NAME` selects one |
| `sceno preview -i f.kdl` | Local live preview, source editing, geometry inspection, verified repairs, Undo, and export |
| `sceno validate -i f --json` | Validate + repair hints + stack warnings |
| `sceno advise -i f --json` | Stack engine + visual score + recommendations |
| `sceno advise -i f --ai` | Optional external AI CLI review (`SCENO_AI_CMD`) |
| `sceno describe -i f --json` | Layout without images (includes engine) |
| `sceno render -i f -o out` | Export PNG (default); `-format svg,pdf` for more; `--all` for every format |
| `sceno render -format slides` | HTML presentation |
| `sceno docs [TOPIC] [--json]` | **Self-doc hub** — guide, spec, goals, shapes, icons, stack, errors, … |
| `sceno docs guide --json` | Full agent handbook (start here) |
| `sceno version [--json]` | Tool version |

## Error codes

Full fixes and examples: `sceno docs errors --json` or `sceno docs guide --json` → `error_codes`.

## Examples in repo

- `examples/how-it-works.kdl` — README pipeline diagram
- `examples/self-service.kdl` — full platform diagram
- `examples/slides-demo.kdl` — slide deck
- `examples/slides-dark.kdl` — dark theme + code
- `examples/shapes-demo.kdl` — shape gallery
