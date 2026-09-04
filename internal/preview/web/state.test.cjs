const test = require("node:test");
const assert = require("node:assert/strict");
const { mergeSource, RequestOrder } = require("./state.js");

test("external file changes preserve the complete unsaved draft and its baseline", () => {
  const draft = { value: "my unfinished edit", baseline: "original", revision: "r1", dirty: true, conflict: false };
  const result = mergeSource(draft, { source: "external change", revision: "r2" });
  assert.deepEqual(result, { ...draft, conflict: true });
  assert.equal(draft.conflict, false, "transition must not mutate its input");
});

test("unchanged snapshot does not conflict with typing and matching revisions clear stale conflict", () => {
  const draft = { value: "my edit", baseline: "original", revision: "r1", dirty: true, conflict: true };
  assert.deepEqual(mergeSource(draft, { source: "original", revision: "r1" }), { ...draft, conflict: false });
});

test("clean editors follow save, undo, and external changes", () => {
  let editor = { value: "", baseline: "", revision: null, dirty: false, conflict: false };
  for (const source of ["original", "saved edit", "original"]) {
    editor = mergeSource(editor, { source, revision: source });
    assert.equal(editor.value, source);
    assert.equal(editor.baseline, source);
    assert.equal(editor.revision, source);
    assert.equal(editor.dirty, false);
    assert.equal(editor.conflict, false);
  }
});

test("a GET started before a write cannot replace its newer saved state", () => {
  const order = new RequestOrder();
  const oldRequest = order.snapshot();
  order.beginWrite();
  assert.equal(order.accepts(oldRequest), false);
  order.finishWrite();
  assert.equal(order.accepts(oldRequest), false);
  assert.equal(order.accepts(order.snapshot()), true);
});

test("snapshots during a pending mutation are rejected and successive writes invalidate pending reads", () => {
  const order = new RequestOrder();
  order.beginWrite();
  assert.equal(order.accepts(order.snapshot()), false);
  order.finishWrite();
  const afterFirstWrite = order.snapshot();
  order.beginWrite();
  order.finishWrite();
  assert.equal(order.accepts(afterFirstWrite), false);
});
