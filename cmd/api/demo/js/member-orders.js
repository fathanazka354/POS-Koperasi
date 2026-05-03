'use strict';

/* ─── Member: Orders ─────────────────────────────────── */
const Orders = {
  async fetchPendingCount() {
    if (!MS.token) return;
    const res  = await API.member('GET', '/shop/orders');
    const list = Array.isArray(res.data) ? res.data : [];
    const n    = list.filter(t=>t.status==='pending').length;
    const el   = document.getElementById('member-pending-badge');
    if (el) { el.textContent=n; el.classList.toggle('hidden',n===0); }
  },

  async load() {
    const el = document.getElementById('orders-list'); if (!el) return;
    el.innerHTML = `<div class="py-8 text-center text-xs text-gray-400 skeleton">Memuat…</div>`;
    const res  = await API.member('GET', '/shop/orders');
    const list = Array.isArray(res.data) ? res.data : [];
    if (!list.length) { el.innerHTML = `<div class="py-12 text-center text-gray-400"><div class="text-4xl mb-2">📦</div><p class="text-sm">Belum ada transaksi</p></div>`; return; }
    const pending = list.filter(t=>t.status==='pending');
    const others  = list.filter(t=>t.status!=='pending');
    el.innerHTML = (pending.length?`<div class="bg-yellow-50 border border-yellow-200 rounded-xl p-3 mb-3 text-xs text-yellow-800">⏳ <b>${pending.length} transaksi</b> menunggu pembayaran.</div>`:'')
      + [...pending,...others].map(t=>`
        <div onclick="Orders.openDetail('${t.invoice_no}')" class="bg-white rounded-xl p-3.5 mb-2 cursor-pointer shadow-sm hover:shadow-md flex items-start gap-3 ${t.status==='pending'?'border-l-4 border-orange-400':''}">
          <div class="flex-1">
            <p class="text-sm font-bold truncate">${t.invoice_no}</p>
            <p class="text-xs text-gray-400">${new Date(t.created_at).toLocaleString('id-ID')}</p>
            <p class="text-sm font-semibold text-gray-700">Rp ${UI.fmt(t.grand_total)}</p>
            ${t.status==='pending'?`<button onclick="event.stopPropagation();Orders.checkStatus('${t.invoice_no}',true)" class="mt-1 text-[11px] font-semibold text-brand border border-brand px-2 py-0.5 rounded-lg hover:bg-brand-light">🔄 Cek Status</button>`:''}
          </div>
          <span class="text-[11px] font-bold px-2.5 py-1 rounded-full ${UI.statusClass(t.status)}">${UI.statusLabel(t.status)}</span>
        </div>`).join('');
  },

  async openDetail(inv) {
    Shop.showView('order-detail');
    const el = document.getElementById('order-detail-content'); if (!el) return;
    el.innerHTML = `<div class="py-8 text-center text-xs text-gray-400 skeleton">Memuat…</div>`;
    const res = await API.member('GET', `/shop/orders/${encodeURIComponent(inv)}`);
    if (!res.success||!res.data) { el.innerHTML = `<p class="text-red-600 text-sm">${res.message||'Tidak ditemukan'}</p>`; return; }
    const {transaction:t, items} = res.data;
    el.innerHTML = `
      <div class="bg-white rounded-xl p-4 shadow-sm mb-3">
        <h4 class="font-bold text-sm mb-3">📋 Detail Pesanan</h4>
        <div class="space-y-1.5 text-sm">
          <div class="flex justify-between"><span class="text-gray-500">Invoice</span><span class="font-bold">${t.invoice_no}</span></div>
          <div class="flex justify-between"><span class="text-gray-500">Tanggal</span><span>${new Date(t.created_at).toLocaleString('id-ID')}</span></div>
          <div class="flex justify-between items-center"><span class="text-gray-500">Status</span><span class="text-[11px] font-bold px-2.5 py-0.5 rounded-full ${UI.statusClass(t.status)}">${UI.statusLabel(t.status)}</span></div>
        </div>
      </div>
      <div class="bg-white rounded-xl p-4 shadow-sm">
        <h4 class="font-bold text-sm mb-3">🛍️ Item</h4>
        <div class="space-y-1 text-sm">
          ${(items||[]).map(i=>`<div class="flex justify-between"><span>Produk #${i.product_id} ×${i.quantity}</span><span>Rp ${UI.fmt(i.subtotal)}</span></div>`).join('')}
          <div class="flex justify-between border-t border-gray-100 pt-2 mt-1 text-gray-500"><span>Subtotal</span><span>Rp ${UI.fmt(t.subtotal)}</span></div>
          ${t.voucher_discount>0?`<div class="flex justify-between text-green-700"><span>Diskon Voucher</span><span>−Rp ${UI.fmt(t.voucher_discount)}</span></div>`:''}
          <div class="flex justify-between text-gray-500"><span>PPN</span><span>Rp ${UI.fmt(t.tax)}</span></div>
          <div class="flex justify-between font-bold border-t border-gray-100 pt-2 mt-1"><span>Total Bayar</span><span class="text-brand">Rp ${UI.fmt(t.grand_total)}</span></div>
        </div>
      </div>`;
  },

  async checkStatus(inv, fromOrders=false) {
    UI.toast('Mengecek status…');
    const res = await API.member('GET', `/shop/orders/${encodeURIComponent(inv)}/status`);
    if (!res.success) { UI.toast('Gagal: '+(res.message||''), true); return; }
    const st = res.data?.status;
    if (st==='paid') {
      MemberNotif.fetchUnread(); this.fetchPendingCount();
      if (fromOrders) { UI.toast('✅ Pembayaran dikonfirmasi!'); this.load(); }
      else { Shop.showView('payment'); document.getElementById('payment-result').innerHTML = `<div class="text-5xl mb-3">✅</div><h3 class="font-bold text-xl">Pembayaran Berhasil!</h3><p class="text-gray-500 text-sm mt-2">Invoice <b>${inv}</b> telah lunas.</p><button onclick="Shop.showView('orders')" class="mt-5 bg-brand text-white px-6 py-2.5 rounded-xl font-semibold">Lihat Riwayat</button>`; }
    } else if (st==='cancelled') {
      UI.toast('❌ Transaksi dibatalkan', true); if (fromOrders) this.load();
    } else {
      UI.toast(`⏳ Status: ${st||'pending'} — belum terkonfirmasi`);
    }
  },
};