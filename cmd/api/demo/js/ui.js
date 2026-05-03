'use strict';

/* ─── UI helpers ─────────────────────────────────────── */
const UI = (() => {
  let _toastTimer;

  function toast(msg, isErr = false, dur = 3000) {
    const el = document.getElementById('toast');
    el.textContent  = msg;
    el.className = `fixed bottom-5 left-1/2 -translate-x-1/2 z-[999] text-white text-sm px-4 py-2.5 rounded-xl shadow-lg max-w-xs text-center transition-opacity duration-300 opacity-100 ${isErr ? 'bg-red-700' : 'bg-gray-800'}`;
    clearTimeout(_toastTimer);
    _toastTimer = setTimeout(() => { el.classList.replace('opacity-100','opacity-0'); }, dur);
  }

  function openDrawer(id) {
    document.getElementById('overlay').classList.remove('hidden');
    const w = id === 'notif-drawer' ? '-380px' : '-400px';
    const el = document.getElementById(id);
    if (el) el.style.right = '0';
  }

  function closeOverlay() {
    document.getElementById('overlay').classList.add('hidden');
    ['notif-drawer','cart-drawer'].forEach(id => {
      const el = document.getElementById(id);
      if (el) el.style.right = el.id === 'notif-drawer' ? '-380px' : '-400px';
    });
  }

  function openAddrModal()  { document.getElementById('addr-modal').classList.remove('hidden'); }
  function closeAddrModal() { document.getElementById('addr-modal').classList.add('hidden'); }

  function fmt(n)  { return Math.round(n).toLocaleString('id-ID'); }
  function timeAgo(s) {
    const d = Math.floor((Date.now() - new Date(s)) / 60000);
    if (d < 1) return 'Baru saja'; if (d < 60) return `${d} mnt lalu`;
    const h = Math.floor(d/60); if (h < 24) return `${h} jam lalu`;
    return `${Math.floor(h/24)} hari lalu`;
  }
  function timeFmt(s) { return s ? new Date(s).toLocaleTimeString('id-ID',{hour:'2-digit',minute:'2-digit'}) : ''; }
  function statusLabel(s) { return {paid:'Lunas',pending:'Menunggu',cancelled:'Dibatalkan'}[s]||s; }
  function statusClass(s) { return {paid:'bg-green-100 text-green-800',pending:'bg-yellow-100 text-yellow-700',cancelled:'bg-red-100 text-red-700'}[s]||'bg-gray-100 text-gray-600'; }
  function emoji(name='') {
    const n = name.toLowerCase();
    if (n.includes('aqua')||n.includes('air'))  return '💧';
    if (n.includes('indomie')||n.includes('mie')) return '🍜';
    if (n.includes('beras'))  return '🌾'; if (n.includes('gula'))  return '🍬';
    if (n.includes('kopi'))   return '☕'; if (n.includes('teh'))   return '🍵';
    if (n.includes('sabun')||n.includes('sunlight')) return '🧴';
    if (n.includes('susu'))   return '🥛'; if (n.includes('telur')) return '🥚';
    if (n.includes('minyak')) return '🫙'; if (n.includes('roti'))  return '🍞';
    return '🛒';
  }

  return { toast, openDrawer, closeOverlay, openAddrModal, closeAddrModal, fmt, timeAgo, timeFmt, statusLabel, statusClass, emoji };
})();