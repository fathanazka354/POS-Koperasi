'use strict';

/* ─── Member: Address ────────────────────────────────── */
const Address = (() => {
  let _selected = 0;
  const selectedID = () => _selected;

  async function loadCheckout() {
    const el = document.getElementById('addr-list-checkout'); if (!el) return;
    el.innerHTML = `<div class="py-2 text-xs text-gray-400 skeleton">Memuat…</div>`;
    const res  = await API.member('GET', '/member/addresses');
    const list = Array.isArray(res.data) ? res.data : [];
    if (!list.length) { el.innerHTML = `<div class="py-2 text-xs text-gray-400">Belum ada alamat.</div>`; return; }
    el.innerHTML = list.map(a=>`
      <div id="ac-${a.id}" onclick="Address.select(${a.id})"
        class="border-2 rounded-xl p-3 mb-2 cursor-pointer transition-all ${a.id===_selected?'border-brand bg-brand-light':'border-gray-200 hover:border-brand/50'}">
        <p class="text-xs font-bold">${a.label}</p>
        <p class="text-xs text-gray-500"><b>${a.recipient}</b> · ${a.phone}</p>
        <p class="text-xs text-gray-400">${a.address_line}, ${a.city}</p>
      </div>`).join('');
    if (!_selected) { const d=list.find(a=>a.is_default)||list[0]; if(d) select(d.id); }
  }

  function select(id) {
    document.querySelectorAll('[id^="ac-"]').forEach(el=>{
      const active = el.id === `ac-${id}`;
      el.className = `border-2 rounded-xl p-3 mb-2 cursor-pointer transition-all ${active?'border-brand bg-brand-light':'border-gray-200 hover:border-brand/50'}`;
    });
    _selected = id;
  }

  async function save() {
    const errEl = document.getElementById('addr-err'); errEl.classList.add('hidden');
    const inp = { label:document.getElementById('al').value||'Rumah', recipient:document.getElementById('ar').value.trim(), phone:document.getElementById('ap').value.trim(), address_line:document.getElementById('aa').value.trim(), city:document.getElementById('ac').value.trim(), province:document.getElementById('apv').value.trim(), postal_code:document.getElementById('az').value.trim(), is_default:false };
    if (!inp.recipient||!inp.address_line||!inp.city) { errEl.textContent='Nama penerima, alamat, kota wajib diisi'; errEl.classList.remove('hidden'); return; }
    const res = await API.member('POST', '/member/addresses', inp);
    if (!res.success) { errEl.textContent=res.message||'Gagal'; errEl.classList.remove('hidden'); return; }
    UI.closeAddrModal(); UI.toast('Alamat berhasil ditambahkan!'); loadCheckout();
  }

  return { selectedID, loadCheckout, select, save };
})();