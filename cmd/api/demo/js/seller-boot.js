'use strict';

/* ─── Seller: boot ───────────────────────────────────── */
const Seller = {
  boot() {
    document.getElementById('page-seller').classList.remove('hidden');
    const name = SS.employee?.full_name || SS.employee?.nik || '';
    document.getElementById('seller-username').textContent = '👤 ' + name;
    SellerChat.init();
    Seller.switchView('chat');
  },
  logout() { SellerAuth.logout(); },

  switchView(which) {
    const chat    = document.getElementById('seller-view-chat');
    const prod    = document.getElementById('seller-view-products');
    const btnChat = document.getElementById('sn-chat');
    const btnProd = document.getElementById('sn-prod');
    const active  = 'seller-nav px-3 py-1 rounded-full text-xs font-semibold bg-white text-seller shadow-sm transition-all';
    const idle    = 'seller-nav px-3 py-1 rounded-full text-xs font-semibold text-white/90 hover:bg-white/15 transition-all';
    if (which === 'products') {
      chat.classList.add('hidden');
      prod.classList.remove('hidden');
      if (btnChat) btnChat.className = idle;
      if (btnProd) btnProd.className = active;
      SellerProducts.load();
    } else {
      chat.classList.remove('hidden');
      prod.classList.add('hidden');
      if (btnChat) btnChat.className = active;
      if (btnProd) btnProd.className = idle;
      if (typeof SellerChat.clearNavBadge === 'function') SellerChat.clearNavBadge();
    }
  },
};