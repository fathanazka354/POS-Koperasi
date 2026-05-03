'use strict';

/* ─── Member: Cart ───────────────────────────────────── */
const MemberCart = {
  render() {
    const count = MS.cart.reduce((s,i)=>s+i.qty,0);
    const badge = document.getElementById('cart-badge');
    if (badge) badge.textContent = count;
    const el = document.getElementById('cart-list');
    if (!el) return;
    if (!MS.cart.length) { el.innerHTML = `<div class="py-10 text-center text-gray-400"><div class="text-4xl mb-2">🛒</div><p class="text-sm">Keranjang kosong</p></div>`; document.getElementById('cart-total').textContent = 'Rp 0'; return; }
    el.innerHTML = MS.cart.map((item,i) => `
      <div class="flex gap-3 py-3 border-b border-gray-50">
        <div class="flex-1">
          <p class="text-sm font-semibold">${item.name}</p>
          <p class="text-xs text-gray-400">Rp ${UI.fmt(item.price)}</p>
          <div class="flex items-center gap-2 mt-1">
            <button onclick="MemberCart.qty(${i},-1)" class="w-6 h-6 border border-gray-200 rounded font-bold text-xs hover:border-brand hover:text-brand transition-colors">−</button>
            <span class="text-sm font-semibold w-5 text-center">${item.qty}</span>
            <button onclick="MemberCart.qty(${i},1)"  class="w-6 h-6 border border-gray-200 rounded font-bold text-xs hover:border-brand hover:text-brand transition-colors">+</button>
            <button onclick="MemberCart.del(${i})" class="ml-1 text-red-400 hover:text-red-600">🗑</button>
          </div>
        </div>
        <p class="font-bold text-brand text-sm">Rp ${UI.fmt(item.price*item.qty)}</p>
      </div>`).join('');
    const total = MS.cart.reduce((s,i)=>s+i.price*i.qty,0);
    document.getElementById('cart-total').textContent = `Rp ${UI.fmt(total)}`;
  },
  add(id,name,price,stock,qty=1) {
    const c = [...MS.cart]; const idx = c.findIndex(i=>i.id===id);
    if (idx>=0) c[idx].qty = Math.min(c[idx].qty+qty,stock);
    else c.push({id,name,price,qty,stock});
    MS.saveCart(c); this.render(); UI.toast(`${name} ditambahkan 🛍️`);
  },
  qty(i,d) {
    const c=[...MS.cart]; c[i].qty+=d;
    if (c[i].qty<=0) c.splice(i,1); else if(c[i].qty>c[i].stock) c[i].qty=c[i].stock;
    MS.saveCart(c); this.render();
  },
  del(i) { const c=[...MS.cart]; c.splice(i,1); MS.saveCart(c); this.render(); },
  openDrawer() { UI.openDrawer('cart-drawer'); },
  subtotal()   { return MS.cart.reduce((s,i)=>s+i.price*i.qty,0); },
};