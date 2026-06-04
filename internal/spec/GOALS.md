# Sceno — goals

## Mission

Build the best **local-first** tool for architecture and system diagrams: one KDL file in, polished diagrams out — without a browser editor, without cloud lock-in, without export surprises. Optimized for humans **and** AI agents that edit specs in a validate → describe → render loop.

## Product goals

1. **KDL-only specs** — One human-friendly format; no YAML/JSON/Lua drift. Specs should read like documentation.
2. **Export parity** — SVG, PNG, PDF, HTML, and slides must look like the same diagram (fonts, shapes, icons, spacing, theme).
3. **Boxes fit text** — Node size is driven by real Inter metrics; labels never clip. Explicit `w`/`h` are minimums, not traps.
4. **Trustworthy layout** — Auto grid, border-snapped edges, obstacle-aware routing, collision resolution, optional organic curves (`style=sketch`).
5. **2D scene understanding** — Paint order (lanes → edges → nodes), groups, occlusion, edge visibility, alignment — exposed via `sceno describe` and validate warnings for agents without vision.
6. **AI-ready validation** — `sceno guide --json` is the agent handbook; `sceno validate --json` returns `fix`, `example`, and `agent.next_steps`; `sceno describe --json` returns narrative, `scene`, `ascii_map`, and `visual_problems`.
7. **PowerPoint familiarity** — Shapes, lanes, callouts, icons, actors, dashed policy lines — familiar to platform and product teams.
8. **Slide-ready export** — Declarative `slide "Title" { }` blocks, `slide=16x9`, HTML deck with keyboard nav, per-slide PNG/SVG.
9. **Theming** — `theme=dark`, `background=transparent`, and overridable variables (`foreground`, `card`, `var.*`) for decks and static exports.
10. **Code on slides** — Syntax-highlighted `code` blocks (`lang=go|json|yaml|bash|kdl`) in slides HTML and SVG.

## Best practices (borrowed from the ecosystem)

We adopt what works from other tools and avoid their lock-in:

| Tool | What we take |
|------|----------------|
| **[d2](https://d2lang.com/)** | Declarative text source of truth; layered layouts; themes; sketch vs polished; validate before export; multiple output formats. |
| **[Mermaid](https://mermaid.js.org/)** | Text-first diagrams in docs/repos; familiar `A -> B` edges; dark/light themes; subgraph-style lanes. |
| **[Excalidraw](https://excalidraw.com/)** | Hand-drawn/sketch aesthetic (`style=sketch`); organic connectors; approachable whiteboard feel. |
| **[PlantUML](https://plantuml.com/)** | Precise layout for architecture; code blocks in technical decks. |
| **[Structurizr](https://structurizr.com/)** | Consistent notation, clear layers, separation of concerns in views. |
| **PowerPoint / Keynote** | Slide titles, 16:9 framing, speaker flow, mixed diagram + code slides. |
| **Figma / shadcn** | Design tokens (zinc palette), subtle borders, readable type scale, dark mode. |

### Layout & readability

- **Logical grouping** — Columns/layers and proximity clusters; condense related nodes; avoid orphan columns when possible (`suggest_compact` warnings).
- **Visible connectors** — Edges under nodes in paint order; route around obstacles; set `fromSide`/`toSide` when paths cross nodes; sketch style uses smooth curves.
- **Aligned labels** — Icon column + centered text; column alignment checks in scene analysis.
- **Readable density** — Not too sparse, not overcrowded; aesthetic score in `describe` `scene.aesthetics`.

### Agent workflow

1. `sceno guide --json` once per session.
2. Edit KDL → `sceno validate --json` until `ok: true`.
3. `sceno describe --json` to sanity-check layout without opening images.
4. `sceno render --all` or `-format slides`.

### Authoring conventions

- Quote strings with spaces: `title="My Platform"`.
- Use `\n` in labels: `"API\nGateway"`.
- Define shapes before edges in the same `diagram { }` or `slide { }` block.
- Prefer `layout=auto` with `layer`, `row`, or `at=col,row`.
- Use `theme=dark` for decks; `background=transparent` for overlays on slides/websites.

## Non-goals (for now)

- Real-time collaborative editing
- WYSIWYG drag-and-drop canvas
- Import from Visio/Lucidchart
- Animation timelines inside slides
- Full IDE-grade syntax highlighting (we cover common langs for slides)

## Quality bar

| Area | Target |
|------|--------|
| Typography | Embedded Inter (OFL), measured widths |
| Icons | Crisp in SVG and PNG; consistent stroke weight |
| Text | Auto-size nodes; `text_overflow` validation |
| Arrows | Rounded orthogonal (polished); smooth/wobble (sketch); open chevron markers |
| Scene | `describe` + validate: occlusion, edge visibility, groups, alignment |
| Slides | `slide "Title" { }`; `.slides.html`; code blocks; dark + transparent theme |
| Visual | shadcn/zinc tokens; theme variables; dark mode |
| CLI | Single static binary; fast render; JSON everywhere for agents |

## Principles

- **Spec is source of truth** — The diagram is computed, not hand-tweaked per format.
- **Fail with advice** — Errors tell you what to change in the KDL, with examples.
- **See without pixels** — Agents use `describe` and `scene` when they cannot view PNG/SVG.
- **Boring dependencies** — Go, embedded fonts, KDL; avoid CDN and heavy GUI stacks.
