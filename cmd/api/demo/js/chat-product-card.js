'use strict';

/* ─── Shared: kartu produk di chat (nama/harga dari GET /shop/products/:id) ─ */
const ChatProductCard = (() => {
  async function prepend(area, conv, productSnap, opts) {
    if (!area) return;
    const seller = opts && opts.seller;
    const pid = (productSnap && productSnap.id) || (conv && conv.product_id);
    let name = (productSnap && productSnap.name) || (conv && conv.product_name);
    let price = productSnap && productSnap.sell_price;
    if (pid) {
      try {
        const pr = await API.pub('GET', `/shop/products/${pid}`);
        if (pr.success && pr.data) {
          if (price == null) price = pr.data.sell_price;
          if (!name) name = pr.data.name;
        }
      } catch {}
    }
    if (!pid && !name) return;
    const dispName = name || ('Produk #' + pid);
    const em = UI.emoji(dispName);
    const priceClass = seller ? 'text-seller' : 'text-brand';
    const footer = seller ? 'Lampiran produk · dari pembeli — ketuk untuk ubah' : 'Lampiran produk · chat dengan penjual — ketuk untuk detail';
    const aria = seller ? 'Ubah produk' : 'Lihat detail produk';
    // Pembeli yang membawa konteks produk → kanan di app member, kiri di app penjual (sisi lawan bicara)
    const rowAlign = seller ? 'justify-start' : 'justify-end';
    const navClick = seller
      ? `event.stopPropagation();ChatNav.toProductEdit(${Number(pid)||0})`
      : `event.stopPropagation();ChatNav.toProductDetail(${Number(pid)||0})`;
    const sticky = document.createElement('div');
    sticky.className = 'mb-3 w-full';
    sticky.innerHTML = `
      <div class="flex ${rowAlign} fade-in w-full">
        <button type="button" aria-label="${aria.replace(/"/g,'&quot;')}"
          class="border border-gray-200 p-3 bg-white shadow-sm flex gap-3 max-w-[70%] rounded-2xl ${seller ? 'rounded-bl-sm' : 'rounded-br-sm'} cursor-pointer hover:shadow-md active:scale-[0.99] transition-all text-left font-sans w-full"
          onclick="${navClick}">
          <div class="w-14 h-14 rounded-lg bg-gray-50 flex items-center justify-center text-3xl flex-none pointer-events-none">${em}</div>
          <div class="flex-1 min-w-0 pointer-events-none">
            <p class="font-semibold text-sm text-gray-900 truncate">${String(dispName).replace(/</g,'&lt;')}</p>
            ${price != null ? `<p class="${priceClass} font-bold text-sm">Rp ${UI.fmt(price)}</p>` : ''}
            <p class="text-[10px] text-gray-400 mt-0.5">${footer}</p>
          </div>
        </button>
      </div>`;
    area.appendChild(sticky);
  }
  return { prepend };
})();