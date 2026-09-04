"use strict";

(() => {
  const $ = (id) => document.getElementById(id);
  const svgNS = "http://www.w3.org/2000/svg";
  const requestOrder = new ScenoState.RequestOrder();
  const ui = { state: null, slide: 0, renderedRevision: null, renderedSlide: null, selected: null, issue: null, sourceRevision: null, baseline: "", draftDirty: false, conflict: false, busy: false, proposal: null, proposalSequence: 0, fetching: false, refreshAgain: false, eventSource: null, retry: null, toast: null, transform: { x: 0, y: 0, scale: 1 }, viewport: null, drag: null };
  const tokenURL = new URL(location.href);
  let token = tokenURL.searchParams.get("token") || sessionStorage.getItem("sceno-token") || "";
  if (tokenURL.searchParams.has("token")) {
    sessionStorage.setItem("sceno-token", token);
    tokenURL.searchParams.delete("token");
    history.replaceState(null, "", tokenURL.pathname + tokenURL.search + tokenURL.hash);
  }

  function node(tag, className, text) {
    const el = document.createElement(tag);
    if (className) el.className = className;
    if (text !== undefined && text !== null) el.textContent = String(text);
    return el;
  }
  function icon(name) {
    const svg = document.createElementNS(svgNS, "svg");
    const use = document.createElementNS(svgNS, "use");
    use.setAttribute("href", `#icon-${name}`);
    svg.append(use);
    svg.setAttribute("aria-hidden", "true");
    return svg;
  }
  function finite(value, fallback = 0) { return Number.isFinite(Number(value)) ? Number(value) : fallback; }
  function format(value) { return Number.isFinite(Number(value)) ? Number(value).toLocaleString(undefined, { maximumFractionDigits: 1 }) : "—"; }
  function rect(value) {
    if (!value) return null;
    const out = { x: Number(value.x), y: Number(value.y), w: Number(value.w), h: Number(value.h) };
    return Object.values(out).every(Number.isFinite) && out.w >= 0 && out.h >= 0 ? out : null;
  }
  function currentSlide() { return ui.state?.slides?.[ui.slide] || null; }
  function ready() { return Boolean(ui.state?.valid && ui.state?.render_ready && !ui.state?.file_error); }
  function canInspect() { return Boolean(ui.state?.valid && currentSlide()?.svg && !ui.state?.file_error); }
  function toast(message) {
    clearTimeout(ui.toast);
    $("toast").textContent = message;
    $("toast").hidden = false;
    ui.toast = setTimeout(() => { $("toast").hidden = true; }, 4500);
  }
  function connection(status, message) {
    $("connection-status").className = `live-status ${status}`;
    $("connection-status").lastElementChild.textContent = message;
  }
  async function request(path, body, download = false) {
    const response = await fetch(path, {
      method: body === undefined ? "GET" : "POST",
      headers: { "X-Sceno-Token": token, ...(body === undefined ? {} : { "Content-Type": "application/json" }) },
      ...(body === undefined ? {} : { body: JSON.stringify(body) }),
      cache: "no-store",
    });
    if (!response.ok) {
      const detail = await response.json().catch(() => ({}));
      const error = new Error(detail.message || detail.error || `Request failed (${response.status}).`);
      error.status = response.status;
      error.detail = detail;
      throw error;
    }
    return download ? response : response.json();
  }
  async function refresh() {
    if (ui.fetching || ui.busy) { ui.refreshAgain = true; return; }
    ui.fetching = true;
    const epoch = requestOrder.snapshot();
    try {
      const state = await request("/api/state");
      if (!requestOrder.accepts(epoch)) { ui.refreshAgain = true; return; }
      receive(state);
      connection("connected", "Live · local");
      if (!ui.eventSource) subscribe();
    } catch (error) {
      connection("offline", "Disconnected");
      if (!ui.state) {
        $("canvas-empty").querySelector("h2").textContent = "Let’s reconnect.";
        $("canvas-empty").querySelector("p").textContent = error.status === 401 || error.status === 403 ? "Open the preview link printed by Sceno to authorize this browser." : "The local preview server is unavailable. It will reconnect automatically.";
      }
      clearTimeout(ui.retry);
      ui.retry = setTimeout(refresh, 2500);
    } finally {
      ui.fetching = false;
      if (ui.refreshAgain) { ui.refreshAgain = false; refresh(); }
    }
  }
  function subscribe() {
    const events = new EventSource(`/api/events?token=${encodeURIComponent(token)}`);
    ui.eventSource = events;
    events.addEventListener("change", () => refresh());
    events.onopen = () => connection("connected", "Live · local");
    events.onerror = () => {
      connection("offline", "Reconnecting");
      clearTimeout(ui.retry);
      ui.retry = setTimeout(refresh, 2500);
    };
  }
  function receive(state) {
    if (!state || typeof state !== "object") return;
    const oldRevision = ui.state?.revision;
    ui.state = state;
    const slides = Array.isArray(state.slides) ? state.slides : [];
    ui.slide = Math.min(ui.slide, Math.max(0, slides.length - 1));
    $("document-name").textContent = String(state.filename || "Untitled diagram").split(/[\\/]/).pop();
    $("document-path").textContent = state.filename || "Local workspace";
    document.title = `${$("document-name").textContent} · Sceno`;
    const editor = ScenoState.mergeSource({ value: $("source-editor").value, baseline: ui.baseline, revision: ui.sourceRevision, dirty: ui.draftDirty, conflict: ui.conflict }, state);
    if ($("source-editor").value !== editor.value) $("source-editor").value = editor.value;
    ui.baseline = editor.baseline; ui.sourceRevision = editor.revision; ui.draftDirty = editor.dirty; ui.conflict = editor.conflict;
    if (!ui.draftDirty && oldRevision !== undefined && oldRevision !== state.revision) {
      $("source-feedback").classList.remove("error");
      $("source-feedback").textContent = state.valid ? "Source is up to date." : "Review the current source findings to restore the preview.";
    }
    if (ui.proposal && oldRevision !== state.revision && state.revision !== ui.proposal.revision) {
      $("apply-repair").disabled = true;
      $("repair-error").textContent = "The source changed after this preview. Close this review and request a fresh suggestion.";
    }
    renderControls();
    renderSlideTabs();
    renderHealth();
    renderFindings();
    renderCanvas();
    renderSelection();
  }
  function setSource(source, revision) {
    $("source-editor").value = source;
    ui.baseline = source;
    ui.sourceRevision = revision;
    ui.draftDirty = false;
    ui.conflict = false;
  }
  function renderControls() {
    $("unsaved-dot").hidden = !ui.draftDirty;
    $("source-conflict").hidden = !ui.conflict;
    $("save-source").disabled = ui.busy || !ui.draftDirty || ui.conflict || Boolean(ui.state?.file_error);
    $("source-editor").readOnly = Boolean(ui.state?.file_error);
    $("undo-repair").hidden = !ui.state?.undo_available;
    $("undo-repair").disabled = ui.busy || ui.draftDirty;
    $("export-toggle").disabled = !ready() || ui.busy;
    $("export-toggle").title = ready() ? "Download the current diagram" : "Resolve structural issues before exporting";
  }
  function renderSlideTabs() {
    const tabs = $("slide-tabs");
    tabs.replaceChildren();
    const slides = ui.state?.slides || [];
    tabs.hidden = slides.length < 2;
    slides.forEach((slide, index) => {
      const button = node("button");
      button.type = "button";
      button.setAttribute("aria-current", String(index === ui.slide));
      button.append(node("span", "slide-number", String(index + 1).padStart(2, "0")), node("span", "", slide.title || `Slide ${index + 1}`));
      button.addEventListener("click", () => {
        ui.slide = index;
        ui.selected = ui.issue = null;
        renderSlideTabs(); renderFindings(); renderCanvas(); renderSelection();
      });
      tabs.append(button);
    });
  }
  function renderHealth() {
    const state = ui.state;
    const score = state?.valid ? finite(state.score, 0) : null;
    $("visual-score").textContent = score === null ? "—" : format(score);
    $("quality-progress").style.width = `${Math.min(100, Math.max(0, score || 0))}%`;
    $("readiness").textContent = ready() ? "Export ready" : state?.valid ? "Needs attention" : "Source needs a fix";
    $("readiness").classList.toggle("blocked", !ready());
    $("quality-trend").textContent = score === null ? "Source needs attention." : score >= 90 ? "Ready to share." : score >= 80 ? "A strong foundation." : "A few changes will help.";
    $("quality-summary").textContent = state?.summary || "Checking layout, content, and connections.";
    const issueCount = (state?.issues || []).length;
    $("issue-count").textContent = issueCount;
    $("issue-count").setAttribute("aria-label", `${issueCount} findings across the document`);
  }
  function renderCanvas() {
    const slide = currentSlide();
    $("slide-title").textContent = slide?.title || (ui.state?.valid ? "Your diagram" : "Every good idea starts with a draft.");
    const bounds = rect(slide?.canvas);
    $("canvas-size").textContent = bounds ? `${format(bounds.w)} × ${format(bounds.h)}` : "—";
    if (!canInspect()) {
      const hadPreview = Boolean($("canvas-content").firstElementChild);
      $("preview-error").hidden = false;
      $("preview-error").textContent = ui.state?.file_error || (hadPreview ? "Showing the last valid preview. The current source needs a fix; geometry inspection is paused." : "This source needs a fix before it can be rendered. Open Source or review the findings to continue.");
      $("canvas-hint").textContent = hadPreview ? "Last valid preview · source has changed" : "The preview will appear after validation passes";
      $("canvas-empty").hidden = hadPreview;
      if (!hadPreview) {
        $("canvas-empty").querySelector("h2").textContent = "A small fix, then the big picture.";
        $("canvas-empty").querySelector("p").textContent = "The findings panel will point you in the right direction.";
        $("canvas-empty").querySelector(".loading-orbit").hidden = true;
      }
      $("canvas-content").querySelectorAll(".inspection-overlay").forEach((el) => el.remove());
      ui.selected = ui.issue = null;
      return;
    }
    $("preview-error").hidden = true;
    $("canvas-empty").hidden = true;
    $("canvas-hint").textContent = "Click a shape to explore its geometry";
    if (ui.renderedRevision === ui.state.revision && ui.renderedSlide === ui.slide) { drawOverlay(); return; }
    const parsed = new DOMParser().parseFromString(slide.svg, "image/svg+xml");
    if (parsed.querySelector("parsererror") || parsed.documentElement.localName !== "svg") {
      $("preview-error").hidden = false;
      $("preview-error").textContent = "The server returned an unreadable SVG. Save the source again to refresh the preview.";
      return;
    }
    const svg = document.importNode(parsed.documentElement, true);
    // Only server-produced SVG enters the canvas. Keep executable content out
    // even if a future renderer accidentally starts emitting it.
    svg.querySelectorAll("script,foreignObject").forEach((el) => el.remove());
    [svg, ...svg.querySelectorAll("*")].forEach((el) => {
      Array.from(el.attributes).forEach((attr) => {
        if (/^on/i.test(attr.name) || (/href$/i.test(attr.name) && !attr.value.startsWith("#"))) el.removeAttribute(attr.name);
      });
    });
    const box = svg.viewBox.baseVal;
    ui.viewport = bounds || { x: box.x, y: box.y, w: box.width, h: box.height };
    svg.setAttribute("width", String(ui.viewport.w));
    svg.setAttribute("height", String(ui.viewport.h));
    svg.setAttribute("role", "group");
    svg.setAttribute("aria-label", `${slide.title || "Diagram preview"}. Shapes can be selected for inspection.`);
    const changedSlide = ui.renderedSlide !== ui.slide;
    const first = ui.renderedRevision === null;
    $("canvas-content").replaceChildren(svg);
    ui.renderedRevision = ui.state.revision;
    ui.renderedSlide = ui.slide;
    drawOverlay();
    if (first || changedSlide) requestAnimationFrame(fitCanvas);
    else positionCanvas();
  }
  function drawOverlay() {
    const svg = $("canvas-content").firstElementChild;
    if (!svg) return;
    svg.querySelectorAll(".inspection-overlay").forEach((el) => el.remove());
    if (!canInspect()) return;
    const group = document.createElementNS(svgNS, "g");
    group.classList.add("inspection-overlay");
    // Large containers go first, leaving their smaller children clickable.
    const nodes = [...(currentSlide().nodes || [])].sort((a, b) => finite(b.bounds?.w) * finite(b.bounds?.h) - finite(a.bounds?.w) * finite(a.bounds?.h));
    nodes.forEach((item) => {
      const bounds = rect(item.bounds);
      if (!bounds) return;
      const target = svgRect(bounds, "hit-target");
      target.setAttribute("tabindex", "0");
      target.setAttribute("role", "button");
      target.setAttribute("aria-label", `Inspect ${item.label || item.id}`);
      const title = document.createElementNS(svgNS, "title");
      title.textContent = `${item.label || item.id} · ${format(bounds.w)} × ${format(bounds.h)}`;
      target.append(title);
      target.addEventListener("click", (event) => { event.stopPropagation(); if (!ui.drag?.moved) selectNode(item.id); });
      target.addEventListener("keydown", (event) => { if (event.key === "Enter" || event.key === " ") { event.preventDefault(); selectNode(item.id); } });
      group.append(target);
    });
    if (ui.selected) {
      const selected = nodes.find((item) => item.id === ui.selected);
      if (rect(selected?.bounds)) group.append(svgRect(selected.bounds, "selected-bounds"));
    }
    if (ui.issue) {
      const issue = (currentSlide().issues || ui.state.issues || []).find((item) => item.id === ui.issue);
      const geometry = issue?.geometry || {};
      Object.values(geometry.bounds || {}).forEach((bounds) => { if (rect(bounds)) group.append(svgRect(bounds, "finding-bounds")); });
      if (rect(geometry.overlap)) group.append(svgRect(geometry.overlap, "finding-bounds"));
    }
    svg.append(group);
  }
  function svgRect(bounds, className) {
    const element = document.createElementNS(svgNS, "rect");
    ["x", "y", "w", "h"].forEach((key) => element.setAttribute(key === "w" ? "width" : key === "h" ? "height" : key, String(bounds[key])));
    element.setAttribute("rx", "5");
    element.classList.add(className);
    return element;
  }
  function positionCanvas() {
    const { x, y, scale } = ui.transform;
    $("canvas-content").style.transform = `translate(${x}px,${y}px) scale(${scale})`;
    $("zoom-value").value = `${Math.round(scale * 100)}%`;
    $("zoom-value").textContent = `${Math.round(scale * 100)}%`;
  }
  function fitCanvas() {
    if (!ui.viewport) return;
    const canvas = $("canvas");
    const scale = Math.max(0.03, Math.min((canvas.clientWidth - 70) / ui.viewport.w, (canvas.clientHeight - 54) / ui.viewport.h, 1.35));
    ui.transform = { scale, x: (canvas.clientWidth - ui.viewport.w * scale) / 2, y: (canvas.clientHeight - ui.viewport.h * scale) / 2 };
    positionCanvas();
  }
  function zoom(factor, point) {
    if (!ui.viewport) return;
    const canvas = $("canvas");
    const anchor = point || { x: canvas.clientWidth / 2, y: canvas.clientHeight / 2 };
    const old = ui.transform.scale;
    const scale = Math.min(6, Math.max(0.03, old * factor));
    ui.transform.x = anchor.x - (anchor.x - ui.transform.x) * scale / old;
    ui.transform.y = anchor.y - (anchor.y - ui.transform.y) * scale / old;
    ui.transform.scale = scale;
    positionCanvas();
  }
  function selectNode(id) {
    if (!canInspect()) return;
    ui.selected = id;
    ui.issue = null;
    switchPanel("insights");
    drawOverlay(); renderSelection(); renderFindings();
    $("selection-panel").scrollIntoView({ block: "nearest", behavior: "smooth" });
  }
  function renderSelection() {
    const selected = canInspect() ? (currentSlide()?.nodes || []).find((item) => item.id === ui.selected) : null;
    $("selection-panel").hidden = !selected;
    if (!selected) return;
    $("selection-title").textContent = selected.label || selected.id;
    $("selection-kind").textContent = `${selected.kind} · ${selected.id}`;
    const bounds = rect(selected.bounds);
    const metrics = $("selection-metrics");
    metrics.replaceChildren();
    if (bounds) [["Width", `${format(bounds.w)} px`], ["Height", `${format(bounds.h)} px`], ["X position", `${format(bounds.x)} px`], ["Y position", `${format(bounds.y)} px`]].forEach(([label, value]) => {
      const metric = node("div", "metric"); metric.append(node("span", "", label), node("strong", "", value)); metrics.append(metric);
    });
    const spacing = currentSlide()?.spacing || {};
    const padding = (spacing.nodes || []).find((item) => item.id === selected.id);
    const details = $("selection-spacing");
    details.replaceChildren();
    if (padding?.content_padding) {
      const list = node("div", "spacing-list");
      list.append(node("h3", "", "Content padding · top / right / bottom / left"));
      const p = node("p"); p.append(node("strong", "", ["top", "right", "bottom", "left"].map((side) => format(padding.content_padding[side])).join(" / ") + " px")); list.append(p); details.append(list);
    }
    const nearest = (spacing.pairs || []).filter((pair) => pair.a === selected.id || pair.b === selected.id).sort((a, b) => finite(a.distance) - finite(b.distance)).slice(0, 3);
    if (nearest.length) {
      const list = node("div", "spacing-list");
      list.append(node("h3", "", "Space to nearby shapes · outer bounds"));
      nearest.forEach((pair) => {
        const p = node("p");
        p.append(node("span", "", pair.a === selected.id ? pair.b : pair.a), node("strong", "", pair.overlap ? "Overlapping" : `${format(pair.distance)} px`));
        p.title = `Horizontal gap: ${format(pair.gap_x)} px; vertical gap: ${format(pair.gap_y)} px. Required axis clearance: ${format(spacing.required_clearance)} px.`;
        list.append(p);
      });
      details.append(list);
    }
    $("selection-source").textContent = selected.line ? `View source · line ${selected.line} ↗` : "View source ↗";
    $("selection-excerpt").textContent = selected.source || "Source location unavailable.";
    $("selection-source").onclick = () => { switchPanel("source"); focusSource(selected.line); };
  }
  function renderFindings() {
    const list = $("findings-list");
    list.replaceChildren();
    const issues = currentSlide()?.issues || ui.state?.issues || [];
    if (!issues.length) {
      const clear = node("div", "all-clear");
      clear.append(node("span", "clear-mark", "✓"), node("strong", "", ui.state?.valid ? "Everything has its place." : "Waiting for valid source."), node("p", "", ui.state?.valid ? "No findings on this slide. The layout, labels, and connections are working together." : "Open Source to review and update the diagram."));
      list.append(clear);
      return;
    }
    for (const [category, title] of [["structural", "Structure"], ["readability", "Readability"], ["style", "Polish"]]) {
      const groupIssues = issues.filter((issue) => (issue.category || (issue.severity === "hint" ? "style" : "structural")) === category);
      if (!groupIssues.length) continue;
      const group = node("section", `finding-group ${category}`);
      const heading = node("h3"); heading.append(node("i", "group-dot"), node("span", "", title), node("span", "group-count", groupIssues.length)); group.append(heading);
      groupIssues.forEach((issue) => {
        const card = node("article", `finding-card${ui.issue === issue.id ? " active" : ""}`);
        const main = node("button", "finding-main"); main.type = "button";
        main.append(node("span", "finding-title", issue.message || issue.code), node("span", "finding-detail", issue.fix || "Select to inspect the affected geometry."));
        const meta = node("span", "finding-meta");
        const ids = Object.keys(issue.geometry?.bounds || {});
        ids.slice(0, 3).forEach((id) => meta.append(node("span", "", id)));
        if (!ids.length && issue.code) meta.append(node("span", "", issue.code));
        main.append(meta);
        main.addEventListener("click", () => {
          ui.issue = issue.id; ui.selected = null;
          drawOverlay(); renderSelection(); renderFindings();
          if (!canInspect()) toast("Geometry inspection resumes when the current source validates.");
        });
        card.append(main);
        const repairs = (issue.repairs || []).filter((repair) => repair.properties && Object.keys(repair.properties).length && repair.properties.overlap !== "allow");
        if (repairs.length) {
          const actions = node("div", "finding-actions");
          const button = node("button", "text-button"); button.type = "button"; button.append(icon("spark"), node("span", "", "Review a repair"));
          button.disabled = ui.busy || ui.draftDirty;
          button.title = ui.draftDirty ? "Save or reload your source draft before reviewing a repair" : repairs[0].reason || "Preview and verify this adjustment";
          button.addEventListener("click", () => previewRepair(issue, repairs[0]));
          actions.append(button); card.append(actions);
        }
        group.append(card);
      });
      list.append(group);
    }
  }
  function switchPanel(panel) {
    for (const name of ["insights", "source"]) {
      const active = panel === name;
      $(`${name}-tab`).setAttribute("aria-selected", String(active));
      $(`${name}-tab`).tabIndex = active ? 0 : -1;
      $(`${name}-panel`).hidden = !active;
    }
  }
  function focusSource(line) {
    const editor = $("source-editor");
    const lines = editor.value.split("\n");
    const index = Math.max(0, Math.min(lines.length - 1, finite(line, 1) - 1));
    const start = lines.slice(0, index).reduce((sum, text) => sum + text.length + 1, 0);
    editor.focus(); editor.setSelectionRange(start, start + lines[index].length);
    editor.scrollTop = Math.max(0, index * parseFloat(getComputedStyle(editor).lineHeight) - editor.clientHeight / 3);
  }
  async function saveSource() {
    if (ui.busy || !ui.draftDirty || ui.conflict || ui.state?.file_error) return;
    ui.busy = true; requestOrder.beginWrite(); renderControls(); renderFindings();
    $("source-feedback").classList.remove("error"); $("source-feedback").textContent = "Saving and checking the scene…";
    const source = $("source-editor").value;
    try {
      const state = await request("/api/source", { revision: ui.sourceRevision, source });
      // Preserve any newer typing that happened while the request was pending.
      const newerDraft = $("source-editor").value;
      ui.draftDirty = false;
      receive(state);
      if (newerDraft !== source) { $("source-editor").value = newerDraft; ui.draftDirty = newerDraft !== ui.baseline; }
      $("source-feedback").textContent = ui.draftDirty ? "Saved the submitted edit. Your latest changes are still unsaved." : state.valid ? "Saved. Your preview is up to date." : "Saved. Review the source findings to restore the preview.";
    } catch (error) {
      $("source-feedback").classList.add("error"); $("source-feedback").textContent = error.message;
      if (error.status === 409) { ui.conflict = true; refresh(); }
    } finally { ui.busy = false; requestOrder.finishWrite(); renderControls(); renderFindings(); if (ui.refreshAgain) { ui.refreshAgain = false; refresh(); } }
  }
  function showDiff(before, after) {
    const a = String(before || "").split("\n"), b = String(after || "").split("\n");
    let first = 0, suffix = 0;
    while (first < a.length && first < b.length && a[first] === b[first]) first++;
    while (suffix < a.length - first && suffix < b.length - first && a[a.length - 1 - suffix] === b[b.length - 1 - suffix]) suffix++;
    const from = Math.max(0, first - 3);
    const aEnd = Math.min(a.length, a.length - suffix + 3), bEnd = Math.min(b.length, b.length - suffix + 3);
    [[a, $("diff-before"), aEnd, "removed"], [b, $("diff-after"), bEnd, "added"]].forEach(([lines, pre, end, className]) => {
      pre.replaceChildren();
      for (let i = from; i < end; i++) {
        const changed = i >= first && i < lines.length - suffix;
        const row = node("span", `diff-line${changed ? ` ${className}` : ""}`);
        row.append(node("span", "diff-number", i + 1), document.createTextNode(lines[i] || " "));
        pre.append(row);
      }
    });
  }
  function showMoves(moves) {
    const list = $("repair-moves-list");
    list.replaceChildren();
    $("repair-moves").hidden = !moves.length;
    $("repair-moves").open = moves.length <= 4;
    $("repair-moves-summary").textContent = `${moves.length} ${moves.length === 1 ? "shape changes" : "shapes change"} position or size`;
    const signed = (value) => `${finite(value) > 0 ? "+" : ""}${format(value)}`;
    moves.forEach((move) => {
      const row = node("div", "repair-move");
      const label = node("div");
      label.append(node("strong", "", move.target), node("span", "", `Slide ${move.slide_index || 1}`));
      const geometry = node("div");
      geometry.append(node("strong", "", `Δx ${signed(move.dx)} px · Δy ${signed(move.dy)} px`));
      const before = rect(move.before), after = rect(move.after);
      if (before && after) geometry.append(node("span", "", `${format(before.x)}, ${format(before.y)} → ${format(after.x)}, ${format(after.y)}${before.w !== after.w || before.h !== after.h ? ` · ${format(before.w)} × ${format(before.h)} → ${format(after.w)} × ${format(after.h)} px` : ""}`));
      row.append(label, geometry); list.append(row);
    });
  }
  async function previewRepair(issue, repair) {
    if (ui.busy || ui.draftDirty) return;
    ui.proposal = null;
    const sequence = ++ui.proposalSequence;
    $("repair-title").textContent = "Review suggested repair";
    $("repair-description").textContent = repair.reason || issue.fix || issue.message;
    $("repair-verification").className = "repair-verification";
    $("repair-verification").textContent = "Rebuilding the scene and checking this adjustment…";
    $("repair-error").textContent = "";
    $("apply-repair").disabled = true;
    $("diff-before").textContent = ""; $("diff-after").textContent = "";
    showMoves([]);
    $("repair-dialog").showModal();
    const revision = ui.state.revision;
    try {
      const result = await request("/api/repair/preview", { revision, slide_index: issue.slide_index || currentSlide()?.index || ui.slide + 1, target: repair.target, properties: repair.properties, issue_id: issue.id });
      if (!$("repair-dialog").open || sequence !== ui.proposalSequence) return;
      ui.proposal = { ...result, revision };
      showDiff(result.before_source, result.after_source);
      showMoves(result.moves || []);
      const changed = revision !== ui.state.revision;
      const verified = result.applicable && !changed;
      $("repair-verification").classList.toggle("verified", verified);
      const reasons = (result.reasons || []).map(String).join(" ");
      $("repair-verification").textContent = verified ? `Verified against the scene. Visual score ${format(result.before_score)} → ${format(result.after_score)}.${reasons ? ` ${reasons}` : ""}` : reasons || "This adjustment did not pass verification. Your source remains unchanged.";
      if (changed) $("repair-error").textContent = "The source changed during verification. Close this review and request a fresh suggestion.";
      $("apply-repair").disabled = !verified;
    } catch (error) {
      if (!$("repair-dialog").open || sequence !== ui.proposalSequence) return;
      $("repair-verification").textContent = "We couldn’t verify this adjustment.";
      $("repair-error").textContent = error.message;
      if (error.status === 409) refresh();
    }
  }
  async function applyRepair() {
    if (!ui.proposal?.applicable || ui.busy || ui.proposal.revision !== ui.state.revision) return;
    ui.busy = true; requestOrder.beginWrite(); $("apply-repair").disabled = true; renderControls();
    try {
      const state = await request("/api/repair/apply", { revision: ui.proposal.revision, proposal_id: ui.proposal.id });
      receive(state); $("repair-dialog").close(); toast("Repair applied. You can undo it from Insights.");
    } catch (error) { $("repair-error").textContent = error.message; if (error.status === 409) refresh(); }
    finally { ui.busy = false; requestOrder.finishWrite(); renderControls(); renderFindings(); if (ui.refreshAgain) { ui.refreshAgain = false; refresh(); } }
  }
  async function undo() {
    if (!ui.state?.undo_available || ui.busy || ui.draftDirty) return;
    ui.busy = true; requestOrder.beginWrite(); renderControls();
    try { receive(await request("/api/undo", { revision: ui.state.revision })); toast("Change undone. The previous source is restored."); }
    catch (error) { toast(error.message); if (error.status === 409) refresh(); }
    finally { ui.busy = false; requestOrder.finishWrite(); renderControls(); renderFindings(); if (ui.refreshAgain) { ui.refreshAgain = false; refresh(); } }
  }
  async function download(formatName) {
    if (!ready()) return;
    setExportMenu(false);
    toast(`Preparing your ${formatName === "slides" ? "slide deck" : formatName.toUpperCase()}…`);
    try {
      const response = await request(`/api/export?format=${encodeURIComponent(formatName)}&slide=${ui.slide + 1}&revision=${encodeURIComponent(ui.state.revision)}`, undefined, true);
      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      const disposition = response.headers.get("Content-Disposition") || "";
      const filename = disposition.match(/filename="([^"]+)"/)?.[1] || `${String(ui.state.filename || "diagram").split(/[\\/]/).pop().replace(/\.kdl$/i, "")}.${formatName === "slides" ? "slides.html" : formatName}`;
      const link = node("a"); link.href = url; link.download = filename; document.body.append(link); link.click(); link.remove();
      setTimeout(() => URL.revokeObjectURL(url), 10000);
      toast("Your download is ready.");
    } catch (error) { toast(error.status === 409 ? "The file changed before export. The preview is refreshing; try the download again." : error.message); if (error.status === 409) refresh(); }
  }
  function setExportMenu(open) { $("export-options").hidden = !open; $("export-toggle").setAttribute("aria-expanded", String(open)); }

  const storedTheme = localStorage.getItem("sceno-theme");
  document.documentElement.dataset.theme = storedTheme || (matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light");
  $("theme-toggle").addEventListener("click", () => { const theme = document.documentElement.dataset.theme === "dark" ? "light" : "dark"; document.documentElement.dataset.theme = theme; localStorage.setItem("sceno-theme", theme); });
  $("export-toggle").addEventListener("click", () => setExportMenu($("export-options").hidden));
  document.querySelectorAll("[data-export]").forEach((button) => button.addEventListener("click", () => download(button.dataset.export)));
  document.addEventListener("click", (event) => { if (!event.target.closest(".export-menu")) setExportMenu(false); });
  document.addEventListener("keydown", (event) => { if (event.key === "Escape") setExportMenu(false); });
  $("insights-tab").addEventListener("click", () => switchPanel("insights"));
  $("source-tab").addEventListener("click", () => switchPanel("source"));
  $("open-source").addEventListener("click", () => { switchPanel("source"); $("source-editor").focus(); });
  document.querySelector(".inspector-tabs").addEventListener("keydown", (event) => {
    if (!["ArrowLeft", "ArrowRight", "Home", "End"].includes(event.key)) return;
    event.preventDefault();
    const panel = event.key === "Home" ? "insights" : event.key === "End" ? "source" : $("insights-tab").getAttribute("aria-selected") === "true" ? "source" : "insights";
    switchPanel(panel); $(`${panel}-tab`).focus();
  });
  $("clear-selection").addEventListener("click", () => { ui.selected = null; drawOverlay(); renderSelection(); });
  $("save-source").addEventListener("click", saveSource);
  $("source-editor").addEventListener("input", () => { ui.draftDirty = $("source-editor").value !== ui.baseline; $("source-feedback").classList.remove("error"); $("source-feedback").textContent = ui.draftDirty ? "Unsaved changes." : "Source matches the saved file."; renderControls(); renderFindings(); });
  $("source-editor").addEventListener("keydown", (event) => {
    if (event.key === "Escape") { event.preventDefault(); $("source-tab").focus(); return; }
    if ((event.metaKey || event.ctrlKey) && event.key === "Enter") { event.preventDefault(); saveSource(); }
    if (event.key === "Tab") {
      event.preventDefault(); const editor = event.target; editor.setRangeText("  ", editor.selectionStart, editor.selectionEnd, "end"); editor.dispatchEvent(new Event("input"));
    }
  });
  $("reload-source").addEventListener("click", () => { if (!ui.state) return; setSource(ui.state.source || "", ui.state.revision); renderControls(); renderFindings(); $("source-feedback").textContent = "Latest file loaded. Your previous draft was discarded."; });
  $("undo-repair").addEventListener("click", undo);
  $("close-repair").addEventListener("click", () => $("repair-dialog").close());
  $("cancel-repair").addEventListener("click", () => $("repair-dialog").close());
  $("apply-repair").addEventListener("click", applyRepair);
  $("repair-dialog").addEventListener("close", () => { ui.proposal = null; ui.proposalSequence++; });
  $("zoom-in").addEventListener("click", () => zoom(1.2));
  $("zoom-out").addEventListener("click", () => zoom(1 / 1.2));
  $("zoom-fit").addEventListener("click", fitCanvas);
  $("canvas").addEventListener("wheel", (event) => {
    if (!ui.viewport) return;
    event.preventDefault(); const box = $("canvas").getBoundingClientRect();
    zoom(Math.exp(-event.deltaY * .0015), { x: event.clientX - box.left, y: event.clientY - box.top });
  }, { passive: false });
  $("canvas").addEventListener("keydown", (event) => {
    if (["+", "="].includes(event.key)) { event.preventDefault(); zoom(1.2); }
    if (event.key === "-") { event.preventDefault(); zoom(1 / 1.2); }
    if (event.key === "0") { event.preventDefault(); fitCanvas(); }
  });
  $("canvas").addEventListener("pointerdown", (event) => {
    if (event.button !== 0 || !ui.viewport) return;
    ui.drag = { id: event.pointerId, startX: event.clientX, startY: event.clientY, x: ui.transform.x, y: ui.transform.y, moved: false };
    if (!event.target.closest(".hit-target")) $("canvas").setPointerCapture(event.pointerId);
  });
  $("canvas").addEventListener("pointermove", (event) => {
    const drag = ui.drag;
    if (!drag || drag.id !== event.pointerId) return;
    const dx = event.clientX - drag.startX, dy = event.clientY - drag.startY;
    if (Math.abs(dx) + Math.abs(dy) > 4) drag.moved = true;
    if (drag.moved) { $("canvas").classList.add("panning"); ui.transform.x = drag.x + dx; ui.transform.y = drag.y + dy; positionCanvas(); }
  });
  function endDrag() { $("canvas").classList.remove("panning"); setTimeout(() => { ui.drag = null; }, 0); }
  $("canvas").addEventListener("pointerup", endDrag);
  $("canvas").addEventListener("pointercancel", endDrag);
  new ResizeObserver(() => { if (ui.viewport) fitCanvas(); }).observe($("canvas"));
  window.addEventListener("beforeunload", (event) => { if (ui.draftDirty) { event.preventDefault(); event.returnValue = ""; } });
  refresh();
})();
