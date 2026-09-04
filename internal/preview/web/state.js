"use strict";

// Small state transitions kept independent of the DOM so preservation of local
// drafts and ordering of server responses can be regression-tested directly.
(() => {
  function mergeSource(current, incoming) {
    if (current.revision === null || !current.dirty) {
      return { value: incoming.source || "", baseline: incoming.source || "", revision: incoming.revision, dirty: false, conflict: false };
    }
    return { ...current, conflict: incoming.revision !== current.revision };
  }
  class RequestOrder {
    constructor() { this.epoch = 0; this.writing = false; }
    beginWrite() { this.epoch++; this.writing = true; }
    finishWrite() { this.writing = false; }
    snapshot() { return this.epoch; }
    accepts(epoch) { return !this.writing && epoch === this.epoch; }
  }
  const api = { mergeSource, RequestOrder };
  if (typeof module !== "undefined" && module.exports) module.exports = api;
  else globalThis.ScenoState = api;
})();
