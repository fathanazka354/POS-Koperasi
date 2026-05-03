'use strict';

/* ─── INIT ───────────────────────────────────────────── */
document.addEventListener('DOMContentLoaded', () => {
  // Boot berdasar session tersimpan
  if (SS.token) { Landing.hide(); Seller.boot(); }
  else if (MS.token) { Landing.hide(); Member.boot(); }
  else Landing.show();

  // Enter key bindings
  document.getElementById('ml-phone').addEventListener('keydown', e => { if(e.key==='Enter') MemberAuth.login(); });
  document.getElementById('sl-pin').addEventListener('keydown',   e => { if(e.key==='Enter') SellerAuth.login(); });
  document.getElementById('mc-input').addEventListener('keydown',  e => { if(e.key==='Enter') MemberChat.send(); });
  document.getElementById('sc-input').addEventListener('keydown',  e => { if(e.key==='Enter') SellerChat.send(); });

  document.getElementById('btn-mc-send').onclick = () => MemberChat.send();
  document.getElementById('btn-sc-send').onclick = () => SellerChat.send();

  document.getElementById('search-input').addEventListener('keydown', e => { if(e.key==='Enter') Shop.search(); });

  const spSearch = document.getElementById('sp-search');
  if (spSearch) {
    let spTimer;
    spSearch.addEventListener('input', () => {
      clearTimeout(spTimer);
      spTimer = setTimeout(() => { if (document.getElementById('seller-view-products') && !document.getElementById('seller-view-products').classList.contains('hidden')) SellerProducts.load(); }, 350);
    });
    spSearch.addEventListener('keydown', e => { if (e.key === 'Enter') SellerProducts.load(); });
  }
});