'use strict';

/* ─── Storage helper ─────────────────────────────────── */
const Store = {
  get(k, fallback = null) { try { const v = localStorage.getItem(k); return v ? JSON.parse(v) : fallback; } catch { return fallback; } },
  set(k, v) { try { localStorage.setItem(k, JSON.stringify(v)); } catch {} },
  del(k)    { try { localStorage.removeItem(k); } catch {} },
};

/* ─── MemberState ────────────────────────────────────── */
const MS = (() => {
  let _token  = Store.get('m_token', '');
  let _member = Store.get('m_member');
  let _cart   = Store.get('m_cart', []);
  return {
    get token()  { return _token; },
    get member() { return _member; },
    get cart()   { return _cart; },
    setAuth(t, m) { _token = t; _member = m; Store.set('m_token', t); Store.set('m_member', m); },
    clearAuth()   { _token = ''; _member = null; Store.del('m_token'); Store.del('m_member'); },
    saveCart(c)   { _cart = c; Store.set('m_cart', c); },
  };
})();

/* ─── SellerState ────────────────────────────────────── */
const SS = (() => {
  let _token    = Store.get('s_token', '');
  let _employee = Store.get('s_employee');
  return {
    get token()    { return _token; },
    get employee() { return _employee; },
    setAuth(t, e) { _token = t; _employee = e; Store.set('s_token', t); Store.set('s_employee', e); },
    clearAuth()   { _token = ''; _employee = null; Store.del('s_token'); Store.del('s_employee'); },
  };
})();