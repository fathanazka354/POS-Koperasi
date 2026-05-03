'use strict';

/* ─── API ────────────────────────────────────────────── */
const API = (() => {
  let _memberRefreshPromise = null;
  let _sellerRefreshPromise = null;
  let _last401Toast = 0;

  async function reqOnce(method, path, body, token) {
    const ctrl = new AbortController();
    const tid  = setTimeout(() => ctrl.abort(), 10000);
    try {
      const h = { 'Content-Type': 'application/json' };
      if (token) h['Authorization'] = 'Bearer ' + token;
      const opts = { method, headers: h, signal: ctrl.signal };
      if (body && method !== 'GET') opts.body = JSON.stringify(body);
      const r = await fetch(Config.BASE + '/api/v1' + path, opts);
      const t = await r.text();
      let json;
      try {
        json = t ? JSON.parse(t) : { success: false, message: 'Respons kosong' };
      } catch {
        json = { success: false, message: 'HTTP ' + r.status };
      }
      return { json, status: r.status };
    } catch (e) {
      return {
        json: { success: false, message: e?.name === 'AbortError' ? 'Timeout' : 'Gagal terhubung ke API' },
        status: 0,
      };
    } finally { clearTimeout(tid); }
  }

  function toastSessionExpiredOnce() {
    if (typeof UI === 'undefined') return;
    const now = Date.now();
    if (now - _last401Toast < 2000) return;
    _last401Toast = now;
    UI.toast('Sesi berakhir, silakan login lagi', true);
  }

  async function doMemberRefresh() {
    if (_memberRefreshPromise) return _memberRefreshPromise;
    const tok = MS.token;
    if (!tok) return false;
    _memberRefreshPromise = (async () => {
      const { json, status } = await reqOnce('POST', '/auth/member-refresh', null, tok);
      if (status !== 200 || !json.success || !json.data?.token) return false;
      MS.setAuth(json.data.token, json.data.member || MS.member);
      if (typeof MemberChat !== 'undefined' && typeof MemberChat.connectWS === 'function') MemberChat.connectWS();
      if (typeof MemberNotif !== 'undefined' && typeof MemberNotif.connectWS === 'function') MemberNotif.connectWS();
      return true;
    })();
    try {
      return await _memberRefreshPromise;
    } finally {
      _memberRefreshPromise = null;
    }
  }

  async function doSellerRefresh() {
    if (_sellerRefreshPromise) return _sellerRefreshPromise;
    const tok = SS.token;
    if (!tok) return false;
    _sellerRefreshPromise = (async () => {
      const { json, status } = await reqOnce('POST', '/auth/employee-refresh', null, tok);
      if (status !== 200 || !json.success || !json.data?.token) return false;
      SS.setAuth(json.data.token, json.data.employee || SS.employee);
      if (typeof SellerChat !== 'undefined' && typeof SellerChat.init === 'function') void SellerChat.init();
      return true;
    })();
    try {
      return await _sellerRefreshPromise;
    } finally {
      _sellerRefreshPromise = null;
    }
  }

  async function authedReq(role, method, path, body) {
    const getTok = () => (role === 'member' ? MS.token : SS.token);
    let token = getTok();
    let { json, status } = await reqOnce(method, path, body, token);
    if (status === 401 && token) {
      const ok = role === 'member' ? await doMemberRefresh() : await doSellerRefresh();
      if (ok) {
        token = getTok();
        const second = await reqOnce(method, path, body, token);
        json = second.json;
        status = second.status;
      }
    }
    if (status === 401) {
      toastSessionExpiredOnce();
      if (role === 'member' && typeof MemberAuth !== 'undefined') MemberAuth.logout();
      else if (role === 'seller' && typeof SellerAuth !== 'undefined') SellerAuth.logout();
    }
    return json;
  }

  return {
    pub: async (m, p, b) => {
      const { json } = await reqOnce(m, p, b, null);
      return json;
    },
    member: (m, p, b) => authedReq('member', m, p, b),
    seller: (m, p, b) => authedReq('seller', m, p, b),
    custom: async (m, p, b, tk) => {
      const { json } = await reqOnce(m, p, b, tk);
      return json;
    },
  };
})();
