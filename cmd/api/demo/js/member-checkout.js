'use strict';

/* ─── Member: Checkout ───────────────────────────────── */
const Checkout = (() => {
  let _method = 'cash';

  function start() {
    if (!MS.token) { Landing.openLoginMember(); UI.toast('Login dulu untuk checkout'); return; }
    if (!MS.cart.length) { UI.toast('Keranjang masih kosong'); return; }
    UI.closeOverlay(); Vouchers.remove(); _method = 'cash'; method('cash');
    goStep(1); Shop.showView('checkout'); Address.loadCheckout();
  }

  function goStep(n) {
    [1,2,3].forEach(i=>{
      document.getElementById(`co-step-${i}`).classList.toggle('hidden', i!==n);
      const t = document.getElementById(`co-tab-${i}`);
      t.className=`flex-1 text-center py-2 rounded-lg text-xs font-bold transition-colors ${i<n?'bg-green-500 text-white':i===n?'bg-brand text-white':'bg-gray-200 text-gray-500'}`;
    });
  }

  function step2() { if (!Address.selectedID()) { UI.toast('Pilih alamat dulu'); return; } goStep(2); renderSummary(); Vouchers.loadGrid(); }
  function step3() { goStep(3); document.getElementById('final-summary').innerHTML = document.getElementById('order-summary').innerHTML; }

  function method(m) {
    _method = m;
    document.querySelectorAll('.pay-opt').forEach(el=>{
      const a = el.dataset.method===m;
      el.className=`pay-opt border-2 rounded-xl p-3 cursor-pointer text-center transition-all ${a?'border-brand bg-brand-light':'border-gray-200 hover:border-brand/50 transition-colors'}`;
    });
  }

  function renderSummary() {
    const sub  = MemberCart.subtotal();
    const disc = Vouchers.disc();
    const tax  = Math.round((sub-disc)*0.11);
    const grand= sub-disc+tax;
    const el   = document.getElementById('order-summary'); if (!el) return;
    el.innerHTML = MS.cart.map(i=>`<div class="flex justify-between"><span>${i.name} ×${i.qty}</span><span>Rp ${UI.fmt(i.price*i.qty)}</span></div>`).join('')
      +`<div class="flex justify-between text-gray-500 mt-1"><span>Subtotal</span><span>Rp ${UI.fmt(sub)}</span></div>`
      +(disc>0?`<div class="flex justify-between text-green-700"><span>Diskon Voucher</span><span>−Rp ${UI.fmt(disc)}</span></div>`:'')
      +`<div class="flex justify-between text-gray-500"><span>PPN (11%)</span><span>Rp ${UI.fmt(tax)}</span></div>`
      +`<div class="flex justify-between font-bold border-t border-gray-100 pt-2 mt-1"><span>Total</span><span class="text-brand">Rp ${UI.fmt(grand)}</span></div>`;
  }

  async function submit() {
    const btn = document.getElementById('btn-pay');
    btn.disabled=true; btn.textContent='Memproses…';
    const res = await API.member('POST', '/shop/checkout', { address_id:Address.selectedID(), voucher_code:Vouchers.code(), items:MS.cart.map(i=>({product_id:i.id,quantity:i.qty})), pay_method:_method });
    btn.disabled=false; btn.textContent='🔒 Bayar Sekarang';
    if (!res.success) { UI.toast('Gagal: '+(res.message||''), true); return; }
    MS.saveCart([]); MemberCart.render();
    _showResult(res.data);
  }

  function _showResult(data) {
    Shop.showView('payment');
    const inv = data?.transaction?.invoice_no||'';
    let html = '';
    if (data?.transaction?.status==='paid') {
      html = `<div class="text-5xl mb-3">✅</div><h3 class="font-bold text-xl mb-2">Pembayaran Berhasil!</h3><p class="text-gray-500 text-sm mb-5">Invoice <b>${inv}</b> dikonfirmasi.</p><button onclick="Shop.showView('orders')" class="bg-brand text-white px-6 py-2.5 rounded-xl font-semibold hover:bg-brand-dark transition-colors mb-2">Lihat Riwayat →</button><br/><button onclick="Shop.showView('home')" class="mt-2 border-2 border-brand text-brand px-6 py-2 rounded-xl font-semibold hover:bg-brand-light">Belanja Lagi</button>`;
    } else if (data?.qr_string) {
      const qr = `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(data.qr_string)}`;
      html = `<div class="text-5xl mb-3">📱</div><h3 class="font-bold text-xl mb-2">Scan QRIS</h3><p class="text-sm text-gray-500 mb-3">Total: <b>Rp ${UI.fmt(data.transaction?.grand_total||0)}</b> · Invoice: <b>${inv}</b></p><div class="border-2 border-gray-200 rounded-2xl p-4 inline-block mb-4"><img src="${qr}" class="w-44 h-44 mx-auto" /></div><br/><button onclick="Orders.checkStatus('${inv}')" class="bg-brand text-white px-6 py-2.5 rounded-xl font-semibold hover:bg-brand-dark transition-colors">🔄 Cek Status</button>`;
    } else if (data?.va_number) {
      html = `<div class="text-5xl mb-3">🏦</div><h3 class="font-bold text-xl mb-2">Transfer VA</h3><div class="border-2 border-gray-200 rounded-2xl px-8 py-4 mb-4"><p class="text-xs text-gray-400 mb-1">No. VA BCA</p><p class="font-bold text-2xl tracking-widest">${data.va_number}</p></div><button onclick="Orders.checkStatus('${inv}')" class="bg-brand text-white px-6 py-2.5 rounded-xl font-semibold hover:bg-brand-dark">🔄 Cek Status</button>`;
    } else {
      html = `<div class="text-5xl mb-3">⏳</div><h3 class="font-bold text-xl mb-2">Menunggu Pembayaran</h3><p class="text-gray-500 text-sm mb-4">Invoice: <b>${inv}</b></p><button onclick="Orders.checkStatus('${inv}')" class="bg-brand text-white px-6 py-2.5 rounded-xl font-semibold hover:bg-brand-dark">🔄 Cek Status</button>`;
    }
    document.getElementById('payment-result').innerHTML = html;
  }

  return { start, goStep, step2, step3, method, renderSummary, submit };
})();