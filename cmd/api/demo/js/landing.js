'use strict';

/* ─── Landing ────────────────────────────────────────── */
const Landing = {
  show() { document.getElementById('page-landing').classList.remove('hidden'); },
  hide() { document.getElementById('page-landing').classList.add('hidden'); },
  openLoginMember() { document.getElementById('modal-login-member').classList.remove('hidden'); },
  openLoginSeller() { document.getElementById('modal-login-seller').classList.remove('hidden'); },
  closeAll() {
    ['modal-login-member','modal-login-seller'].forEach(id => document.getElementById(id).classList.add('hidden'));
    ['ml-err','sl-err'].forEach(id => { const e=document.getElementById(id); if(e){e.classList.add('hidden');} });
  },
};