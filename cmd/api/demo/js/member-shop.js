'use strict';

/* ─── Member: Shop / Products ────────────────────────── */
const Shop = {
  _current: null, _qty: 1, _max: 99,

  init() { this.loadProducts(); },

  showView(name) {
    document.querySelectorAll('.member-shop-view').forEach(v=>v.classList.add('hidden'));
    const el = document.getElementById('sv-'+name);
    if (el) el.classList.remove('hidden');
    if (name === 'orders') Orders.load();
  },

  async loadProducts(q='') {
    const grid = document.getElementById('product-grid');
    grid.innerHTML = `<div class="col-span-4 py-12 flex flex-col items-center gap-3 text-gray-300 skeleton">
      <div class="w-12 h-12 bg-gray-200 rounded-full"></div><div class="h-3 bg-gray-200 rounded w-32"></div></div>`;
    const res  = await API.pub('GET', '/shop/products' + (q?`?search=${encodeURIComponent(q)}`:''));
    const list = Array.isArray(res.data) ? res.data : [];
    if (!res.success) { grid.innerHTML = `<div class="col-span-4 text-center py-10 text-gray-400"><div class="text-4xl mb-2">⚠️</div><p class="text-sm">${res.message}</p><button onclick="Shop.loadProducts()" class="mt-3 text-xs text-brand border border-brand px-3 py-1 rounded-lg">Coba Lagi</button></div>`; return; }
    if (!list.length) { grid.innerHTML = `<div class="col-span-4 text-center py-10 text-gray-400"><div class="text-4xl mb-2">📦</div><p class="text-sm">Tidak ditemukan</p></div>`; return; }
    grid.innerHTML = list.map(p=>`
      <div onclick="Shop.openDetail(${p.id})" class="bg-white rounded-xl overflow-hidden cursor-pointer shadow-sm hover:-translate-y-1 hover:shadow-md transition-all">
        <div class="h-28 bg-gray-50 flex items-center justify-center text-5xl">${UI.emoji(p.name)}</div>
        <div class="p-2.5">
          <p class="text-xs font-semibold truncate">${p.name}</p>
          <p class="text-brand font-bold text-sm">Rp ${UI.fmt(p.sell_price)}</p>
          <p class="text-[11px] text-gray-400">${p.stock>0?`Stok: ${p.stock}`:'<span class="text-red-400">Habis</span>'}</p>
        </div>
      </div>`).join('');
  },

  search() { this.loadProducts(document.getElementById('search-input').value.trim()); },

  async openDetail(id) {
    this.showView('detail');
    const el = document.getElementById('detail-content');
    el.innerHTML = `<div class="bg-white rounded-xl p-5 shadow-sm skeleton h-40"></div>`;
    const res = await API.pub('GET', `/shop/products/${id}`);
    if (!res.success||!res.data) { el.innerHTML = `<p class="text-red-600 text-sm">${res.message}</p>`; return; }
    this._current = res.data; this._qty = 1; this._max = res.data.stock;
    this._renderDetail();
  },

  _renderDetail() {
    const p = this._current;
    document.getElementById('detail-content').innerHTML = `
      <div class="bg-white rounded-xl p-5 flex flex-col sm:flex-row gap-5 shadow-sm">
        <div class="w-full sm:w-40 h-40 bg-gray-50 rounded-xl flex items-center justify-center text-7xl flex-none">${UI.emoji(p.name)}</div>
        <div class="flex-1">
          <h2 class="font-bold text-xl mb-1">${p.name}</h2>
          <p class="text-brand font-bold text-2xl mb-1">Rp ${UI.fmt(p.sell_price)}</p>
          <p class="text-xs text-gray-400 mb-3">Satuan: ${p.unit} | Stok: ${p.stock}</p>
          ${p.stock<=0?'<p class="text-red-500 font-semibold text-sm mb-3">Stok Habis</p>':''}
          <div class="flex items-center gap-3 mb-4">
            <button onclick="Shop.changeQty(-1)" class="w-8 h-8 border border-gray-200 rounded-lg font-bold text-lg hover:border-brand hover:text-brand transition-colors flex items-center justify-center">−</button>
            <span id="qty-val" class="font-bold text-base w-8 text-center">${this._qty}</span>
            <button onclick="Shop.changeQty(1)"  class="w-8 h-8 border border-gray-200 rounded-lg font-bold text-lg hover:border-brand hover:text-brand transition-colors flex items-center justify-center">+</button>
            <span class="text-xs text-gray-400">maks. ${p.stock}</span>
          </div>
          <button onclick="Shop.addCurrent()" ${p.stock<=0?'disabled':''}
            class="w-full bg-brand text-white py-3 rounded-xl font-semibold hover:bg-brand-dark transition-colors disabled:opacity-50">+ Tambah ke Keranjang</button>
          ${MS.token
            ? `<button type="button" onclick="MemberChat.startFromCurrentProduct(1)" class="w-full mt-2 border-2 border-brand text-brand py-2.5 rounded-xl font-semibold hover:bg-brand-light transition-colors flex items-center justify-center gap-2">💬 Chat dengan penjual</button>`
            : `<p class="text-xs text-gray-500 mt-2 text-center">Untuk tanya stok &amp; kondisi, <button type="button" class="text-brand font-semibold underline" onclick="Landing.openLoginMember()">login</button> lalu gunakan chat dengan penjual.</p>`}
        </div>
      </div>`;
  },

  changeQty(d) {
    const el = document.getElementById('qty-val'); if (!el) return;
    let v = parseInt(el.textContent)+d;
    v = Math.max(1, Math.min(v, this._max));
    el.textContent = v; this._qty = v;
  },

  addCurrent() {
    if (!this._current) return;
    MemberCart.add(this._current.id, this._current.name, this._current.sell_price, this._current.stock, this._qty);
  },
};