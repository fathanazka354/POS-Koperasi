'use strict';

/* ─── Member: boot ───────────────────────────────────── */
const Member = {
  boot() {
    document.getElementById('page-member').classList.remove('hidden');
    document.getElementById('member-username').textContent = '👤 ' + (MS.member?.full_name?.split(' ')[0] || '');
    Shop.init();
    MemberCart.render();
    MemberNotif.connectWS();
    MemberNotif.fetchUnread();
    MemberChat.connectWS();
    Orders.fetchPendingCount();
  },
  logout() { MemberAuth.logout(); },
  goHome() { switchTab('shop'); Shop.showView('home'); },
  switchTab(t) { switchTab(t); },
};

function switchTab(t) {
  const shop = document.getElementById('member-shop');
  const chat = document.getElementById('member-chat');
  if (t === 'shop') { shop.classList.remove('hidden'); chat.classList.add('hidden'); }
  else {
    shop.classList.add('hidden'); chat.classList.remove('hidden');
    const ub = document.getElementById('member-chat-unread-badge');
    if (ub) { ub.textContent = '0'; ub.classList.add('hidden'); }
    MemberChat.onOpen();
  }
  document.querySelectorAll('.member-tab-btn').forEach(b => {
    const a = b.dataset.tab === t;
    b.classList.toggle('bg-white',         a);
    b.classList.toggle('text-brand',       a);
    b.classList.toggle('text-white/80',    !a);
    b.classList.toggle('hover:text-white', !a);
  });
}