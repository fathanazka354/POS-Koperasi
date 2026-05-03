'use strict';

/* ─── Seller: Auth ───────────────────────────────────── */
const SellerAuth = {
  async login() {
    const nik = document.getElementById('sl-nik').value.trim();
    const pin = document.getElementById('sl-pin').value;
    const err = document.getElementById('sl-err');
    err.classList.add('hidden');
    if (!nik||!pin) { err.textContent='NIK dan PIN wajib diisi'; err.classList.remove('hidden'); return; }
    const btn = document.getElementById('btn-sl-submit');
    btn.disabled=true; btn.textContent='Memproses…';
    const res = await API.pub('POST', '/auth/login', { nik, pin });
    btn.disabled=false; btn.textContent='Masuk sebagai Penjual';
    if (!res.success||!res.data?.token) { err.textContent=res.message||'Login gagal'; err.classList.remove('hidden'); return; }
    SS.setAuth(res.data.token, res.data.employee);
    Landing.closeAll(); Landing.hide();
    Seller.boot();
  },
  logout() {
    SS.clearAuth();
    SellerChat.disconnectWS();
    document.getElementById('page-seller').classList.add('hidden');
    Landing.show();
  },
};