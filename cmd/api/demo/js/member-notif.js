'use strict';

/* ─── Member: Notifications ──────────────────────────── */
const MemberNotif = (() => {
  let _ws = null;

  function connectWS() {
    if (!MS.token) return;
    if (_ws) try { _ws.close(); } catch {}
    const url = Config.BASE.replace(/^http/,'ws') + '/api/v1/ws/notify?token=' + MS.token;
    _ws = new WebSocket(url);
    _ws.onmessage = e => {
      try {
        const m = JSON.parse(e.data);
        if (m.event === 'notification') {
          fetchUnread();
          UI.toast('🔔 ' + (m.notification?.title||'Notifikasi baru'), false, 5000);
          if (document.getElementById('notif-drawer').style.right === '0px') loadList();
        }
      } catch {}
    };
    _ws.onclose = () => { if (MS.token) setTimeout(connectWS, 5000); };
  }

  function disconnectWS() { if (_ws) { _ws.close(); _ws = null; } }

  async function fetchUnread() {
    if (!MS.token) return;
    const res = await API.member('GET', '/member/notifications');
    if (res.success) setBadge(res.data?.unread || 0);
  }

  function setBadge(n) {
    const el = document.getElementById('notif-badge');
    if (!el) return;
    el.textContent = n > 9 ? '9+' : String(n);
    el.classList.toggle('hidden', n === 0);
  }

  function openDrawer() { UI.openDrawer('notif-drawer'); loadList(); }

  async function loadList() {
    const el = document.getElementById('notif-list');
    if (!el) return;
    if (!MS.token) { el.innerHTML = `<div class="py-8 text-center text-xs text-gray-400">Silakan login</div>`; return; }
    el.innerHTML = `<div class="py-6 text-center text-xs text-gray-400 skeleton">Memuat…</div>`;
    const res   = await API.member('GET', '/member/notifications');
    const items = Array.isArray(res.data?.items) ? res.data.items : [];
    setBadge(res.data?.unread || 0);
    if (!items.length) { el.innerHTML = `<div class="py-8 text-center text-gray-400"><div class="text-3xl mb-2">🔔</div><p class="text-xs">Belum ada notifikasi</p></div>`; return; }
    el.innerHTML = items.map(n => {
      const inv = n.data?.invoice_no || '';
      return `<div onclick="MemberNotif.click('${n.id}','${inv}')"
        class="px-4 py-3 border-b border-gray-50 cursor-pointer hover:bg-gray-50 transition-colors ${n.is_read?'':'bg-orange-50'}">
        <p class="text-sm font-semibold">${n.title}</p>
        <p class="text-xs text-gray-500 mt-0.5">${n.body}</p>
        <p class="text-[11px] text-gray-400 mt-1">${UI.timeAgo(n.created_at)}</p>
      </div>`;
    }).join('');
  }

  async function click(id, inv) {
    await API.member('PUT', `/member/notifications/${id}/read`);
    if (inv) { UI.closeOverlay(); Orders.openDetail(inv); Shop.showView('order-detail'); }
    else loadList();
  }

  async function markAllRead() {
    if (!MS.token) return;
    await API.member('PUT', '/member/notifications/read-all');
    setBadge(0); loadList(); UI.toast('Semua notifikasi ditandai dibaca');
  }

  return { connectWS, disconnectWS, fetchUnread, openDrawer, loadList, click, markAllRead };
})();