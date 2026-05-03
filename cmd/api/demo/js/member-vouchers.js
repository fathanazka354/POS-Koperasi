'use strict';

/* ─── Member: Vouchers ───────────────────────────────── */
const Vouchers = (() => {
  let _applied = null;
  const get = () => _applied;
  const disc = () => _applied?.discount || 0;
  const code = () => _applied?.code || '';

  function tab(t) {
    const a = 'px-3 py-1 rounded-full text-xs font-semibold bg-brand text-white';
    const b = 'px-3 py-1 rounded-full text-xs font-semibold border border-gray-300 text-gray-600 hover:border-brand hover:text-brand';
    document.getElementById('vtab-list').className  = t==='list'  ? a : b;
    document.getElementById('vtab-input').className = t==='input' ? a : b;
    document.getElementById('vpanel-list').classList.toggle('hidden', t!=='list');
    document.getElementById('vpanel-input').classList.toggle('hidden', t!=='input');
  }

  async function loadGrid() {
    const el = document.getElementById('voucher-grid');
    el.innerHTML = `<div class="col-span-2 py-3 text-xs text-gray-400 skeleton">Memuat voucher…</div>`;
    const res  = await API.pub('GET', '/shop/vouchers');
    const list = Array.isArray(res.data) ? res.data : [];
    const sub  = MemberCart.subtotal();
    if (!list.length) { el.innerHTML = `<div class="col-span-2 text-xs text-gray-400">Tidak ada voucher tersedia</div>`; return; }
    el.innerHTML = list.map(v => {
      const ok  = sub >= v.min_purchase;
      const sel = _applied?.code === v.code;
      const lbl = v.discount_type==='percent' ? `Hemat ${v.value}%` : `Hemat Rp ${UI.fmt(v.value)}`;
      return `<div onclick="${ok?`Vouchers.pick('${v.code}')`:`UI.toast('Min belanja Rp ${UI.fmt(v.min_purchase)}',true)`}"
        class="border-2 rounded-xl p-3 cursor-pointer transition-all
          ${sel?'border-brand bg-brand-light':ok?'border-gray-200 hover:border-brand/50':'border-gray-100 opacity-50 bg-gray-50'}">
        <p class="text-[10px] font-bold text-brand uppercase mb-0.5">${v.discount_type==='ongkir'?'🚚 Ongkir':v.discount_type==='percent'?'% Diskon':'Rp Diskon'}</p>
        <p class="text-xs font-bold">${v.code}</p>
        <p class="text-[11px] text-gray-500">${v.name}</p>
        <p class="text-sm font-bold text-brand mt-1">${lbl}</p>
        <p class="text-[10px] text-gray-400">Min. Rp ${UI.fmt(v.min_purchase)}</p>
      </div>`;
    }).join('');
  }

  async function pick(code) {
    const res = await API.pub('POST', '/shop/voucher/validate', { code, subtotal: MemberCart.subtotal() });
    if (!res.success) { UI.toast('❌ '+(res.message||'Tidak valid'), true); return; }
    _apply(code, res.data.discount_amount, res.data.label);
    loadGrid();
  }

  async function applyManual() {
    const code  = document.getElementById('voucher-code-input').value.trim().toUpperCase();
    const errEl = document.getElementById('voucher-err');
    errEl.classList.add('hidden'); if (!code) return;
    const res = await API.pub('POST', '/shop/voucher/validate', { code, subtotal: MemberCart.subtotal() });
    if (!res.success) { errEl.textContent = '❌ '+(res.message||'Kode tidak valid'); errEl.classList.remove('hidden'); return; }
    _apply(code, res.data.discount_amount, res.data.label);
  }

  function _apply(code, amount, label) {
    _applied = { code, discount: amount, label };
    document.getElementById('voucher-applied').innerHTML = `
      <div class="inline-flex items-center gap-2 bg-green-50 border border-green-200 text-green-800 text-xs font-semibold px-3 py-1.5 rounded-lg">
        🏷️ <b>${code}</b> — ${label||`Hemat Rp ${UI.fmt(amount)}`}
        <button onclick="Vouchers.remove()" class="ml-1 text-green-600 font-bold text-sm leading-none hover:text-green-800">×</button>
      </div>`;
    Checkout.renderSummary(); UI.toast(`Voucher ${code} diterapkan!`);
  }

  function remove() {
    _applied = null;
    const inp = document.getElementById('voucher-code-input'); if (inp) inp.value='';
    const err = document.getElementById('voucher-err'); if (err) err.classList.add('hidden');
    document.getElementById('voucher-applied').innerHTML = '';
    loadGrid(); Checkout.renderSummary(); UI.toast('Voucher dihapus');
  }

  return { get, disc, code, tab, loadGrid, pick, applyManual, remove };
})();