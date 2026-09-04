# Render backend consolidation: `tdewolff/canvas`

Date: 2026-07-19

Status update (2026-09-05): this remains a historical migration proposal, not the current implementation plan. Sceno now rasterizes the canonical SVG for polished PNG, outlines embedded Inter text before rasterization, and keeps code and connectors inside the same SVG scene for HTML slides. The separate PNG icon overlay and silent fallback to a second polished drawing path have been removed. PDF still has its own backend. These changes reduce the original parity gap without adopting `tdewolff/canvas`; any future migration must be assessed against the updated corpus and current runtime architecture documentation (`sceno docs architecture --json`).

## Recommendation

Prototype `tdewolff/canvas` behind an internal draw-list adapter, but do not replace the three production backends yet. It is a strong architectural fit: one canvas can emit SVG, PDF, and raster formats, and its renderer contract accepts the same paths, text, images, and transforms across targets. That directly addresses Sceno's backend-parity burden. The project also supports cubic/arc paths, font embedding and subsetting, and in-memory font loading. Sources: [project README](https://github.com/tdewolff/canvas), [renderer API](https://pkg.go.dev/github.com/tdewolff/canvas/renderers), [font API](https://pkg.go.dev/github.com/tdewolff/canvas).

The main reservation is maturity and migration risk. The module still uses pseudo-versions rather than stable tagged releases, its recent API notes show active breaking changes, and its text stack adds a materially larger dependency surface. Its README also calls out FriBidi's LGPL license, which needs a distribution review before adoption. Sceno should not trade known backend duplication for unpinned output or typography drift.

## Carry-over design

- Keep `internal/geom`, `internal/layout`, and measured scene data as the source of truth. Build a small ordered draw list from those numbers; `canvas` is only a consumer.
- Load the four embedded Inter byte slices into one `canvas.FontFamily` as regular, medium, semibold, and bold. The API supports loading font bytes directly, while the PDF renderer supports font subsetting. During migration, compare `FontFace.TextWidth` with `internal/measure` for every corpus string; switch measurement and drawing together only after they agree.
- Convert the icon catalog's existing 24×24 SVG fragments to canvas paths once and cache them. `canvas.ParseSVG` can bootstrap this, but a normalized internal icon path avoids reparsing and preserves vector icons in PDF instead of the current PNG overlay. Apply the existing icon transform and `currentColor` replacement before recording it.
- Preserve Sceno's top-left pixel coordinate system through one explicit transform. Canvas/PDF units and Y-axis conventions must not leak back into geometry or diagnostics.
- Continue generating HTML and slides from the resulting SVG. They do not need a separate canvas renderer.

## Safe migration sequence

1. Record one representative shape, connector, label, Inter text run, and icon into a backend-neutral draw list.
2. Render that list through canvas to SVG, PDF, and PNG; add silhouette, text-bounds, paint-order, and deterministic-output tests.
3. Run the full corpus side-by-side. Require identical semantic bounds and zero new structural findings; use perceptual image diffs for antialiasing changes.
4. Switch PNG first, then PDF. Switch SVG last because it is the current reference backend and HTML/slides consume it.
5. Remove `gg`, `go-pdf/fpdf`, `oksvg`, and `rasterx` only after the new path passes `mask verify` on all supported platforms.

Decision gate: proceed only if embedded Inter metrics stay within 0.5px per line, vector icons survive SVG/PDF, exports are byte-deterministic for identical inputs, and the dependency/license review is acceptable.
