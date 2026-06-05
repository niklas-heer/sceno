# Sceno — stack validation model

Sceno validates diagrams as **stacked 2D planes** (back → front), not as a single flat canvas. Think of it like a lightweight 3D stack where each layer is a 2D plane — collision and routing checks **project** onto reduced planes when needed.

## Plane order (back → front)

| Plane | Contents | Purpose |
|-------|----------|---------|
| `background` | Canvas bounds | Whitespace / density rules |
| `lanes` | `lane`, `container` swimlanes | Grouping backdrop |
| `edges` | Connector paths | Routing plane checks |
| `structure` | `frame`, `group` | Structural grouping |
| `annotations` | `infobox`, `info`, `tip`, `warning`, `note`, `textbox` | Callouts without blocking flow |
| `nodes` | Primary flow shapes (`box`, `cloud`, …) | Main diagram content |
| `labels` | Edge label boxes | Label placement (horizontal / vertical) |
| `chrome` | Title / subtitle band | Visual hierarchy |

Paint order at render time: **lanes → edges → nodes** (annotations render as nodes with accent stripe).

## Reduced 2D projections

| Check | Planes used | What it catches |
|-------|-------------|-----------------|
| Node collision | `nodes` + `annotations` | Overlapping boxes on the node plane |
| Edge routing | `edges` vs `nodes` | Connectors crossing through nodes |
| Flow blocking | `annotations` vs `edges` | Callouts sitting on the main left→right path |
| Scene occlusion | paint order + overlap area | Nodes drawn on top of each other |

## Visual design rules

The stack engine runs these rules (see `sceno docs stack --json` for the full catalog):

| Rule ID | Source | Checks |
|---------|--------|--------|
| `whitespace` | Gestalt / IxDF | Canvas density — not too crowded or empty |
| `hierarchy` | NN/g | Title when many nodes; clear focal point |
| `alignment` | PowerPoint grids | Column alignment, icon + label balance |
| `edge_clarity` | d2 / Mermaid | Edge visibility fraction ≥ ~82% |
| `element_budget` | C4 | ≤15 primary nodes per view |
| `slide_focus` | 10/20/30, Visme | ≤9 shapes per slide (one idea per slide) |
| `annotations` | PowerPoint callouts | Suggest infobox/note; warn if blocking flow |
| `collision_2d` | Sceno stack | Node-plane overlaps |
| `routing_plane` | Sceno stack | Edge–node crossings |

Each finding includes `severity` (`error`, `warning`, `hint`), `fix`, and often `example` KDL.

## Commands

```bash
# Stack summary + visual score + recommendations
sceno advise -i file.kdl --json

# Optional: pipe structured analysis to an external AI CLI
export SCENO_AI_CMD="codex exec -"
sceno advise -i file.kdl --json --ai

# Describe includes scene.stack + slides[n].engine
sceno describe -i file.kdl --json

# Validate runs stack engine rules as warnings/hints
sceno validate -i file.kdl --json
```

## JSON fields (`sceno advise --json`)

| Field | Description |
|-------|-------------|
| `validation_ok` | Same as `sceno validate` `ok` |
| `visual_score` | 0–100 quality score from rules + aesthetics |
| `stack` | Plane counts and canvas size |
| `engine.findings` | Rule outcomes with fix hints |
| `engine.rules_run` | Which rules executed |
| `visual_rules` | Full rule catalog with descriptions |
| `recommendations` | Prioritized actionable hints |
| `ai_review` | Present when `--ai` and `SCENO_AI_CMD` set |

## Agent workflow

1. `sceno docs guide --json` — handbook + `stack_model` + `visual_rules`
2. Edit KDL → `sceno validate --json` until `ok: true`
3. `sceno advise --json` — visual polish and design recommendations
4. `sceno describe --json` — spatial layout without images
5. `sceno render --all`

## Annotation shapes

Semantic callout kinds (all render as infobox with accent stripe):

```kdl
shape info ctx "Context" icon=info subtitle="Supporting detail" at=0,2
shape warning alert "Caution" icon=shield at=1,2
shape tip hint "Pro tip" icon=info at=2,2
shape infobox note "Note" icon=info accent="#7c3aed" at=3,2
shape note sticky "Sticky" at=0,3
```

Default accents: `info` → blue, `warning` → amber, `tip` → green. Override with `accent="#hex"`.

## Icon placement

```kdl
shape box api "API" icon=api iconPos=top-left
```

Values: `top-left` (default), `top`, `top-right`, `center`, `bottom-left`, `bottom`, `bottom-right`.

Single-row pipelines (e.g. left→right flow) auto-center nodes vertically within the row band for straight connectors.
