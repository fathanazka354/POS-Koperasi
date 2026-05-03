'use strict';

/* ─── Member: Auth ───────────────────────────────────── */
const MemberAuth = {
  async login() {
    const code  = document.getElementById('ml-code').value.trim();
    const phone = document.getElementById('ml-phone').value.trim();
    const errEl = document.getElementById('ml-err');
    errEl.classList.add('hidden');
    if (!code || !phone) { errEl.textContent = 'Isi kode anggota dan telepon'; errEl.classList.remove('hidden'); return; }
    const btn = document.getElementById('btn-ml-submit');
    btn.disabled = true; btn.textContent = 'Memproses…';
    const res = await API.pub('POST', '/auth/member-login', { member_code: code, phone });
    btn.disabled = false; btn.textContent = 'Masuk sebagai Pembeli';
    if (!res.success || !res.data?.token) {
      errEl.textContent = res.message || 'Login gagal'; errEl.classList.remove('hidden'); return;
    }
    MS.setAuth(res.data.token, res.data.member);
    Landing.closeAll(); Landing.hide();
    Member.boot();
  },
  logout() {
    MS.clearAuth(); MS.saveCart([]);
    MemberNotif.disconnectWS(); MemberChat.disconnectWS();
    document.getElementById('page-member').classList.add('hidden');
    Landing.show();
  },
};