'use strict';

/* ─── Config ─────────────────────────────────────────── */
const Config = (() => {
  function getBase() {
    try {
      const o = location.origin;
      if (!o || o === 'null' || location.protocol === 'file:') return 'http://localhost:8080';
      return o;
    } catch { return 'http://localhost:8080'; }
  }
  return { BASE: getBase() };
})();