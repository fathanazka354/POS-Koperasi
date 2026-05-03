'use strict';

/* ─── Seller: Produk CRUD ───────────────────────────── */
const SellerProducts = (() => {
  let _editId = null;

  async function load() {
    const q   = document.getElementById('sp-search')?.value.trim() || '';
    const el  = document.getElementById('sp-list');
    el.innerHTML = `<div class="p-8 text-center text-xs text-gray-400 skeleton">Memuat produk…</div>`;
    const res = await API.seller('GET', `/seller/products?search=${encodeURIComponent(q)}`);
    const list = Array.isArray(res.data) ? res.data : [];
    if (!res.success) {
      el.innerHTML = `<div class="p-8 text-center text-red-600 text-sm">${res.message || 'Gagal memuat data'}</div>`;
      return;
    }
    if (!list.length) {
      el.innerHTML = `<div class="p-10 text-center text-gray-400 text-sm">Belum ada produk. Klik <b>Tambah Produk</b>.</div>`;
      return;
    }
    el.innerHTML = `
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-gray-50 border-b border-gray-100">
            <tr class="text-left text-[11px] font-bold text-gray-500 uppercase tracking-wide">
              <th class="px-4 py-3">Produk</th>
              <th class="px-3 py-3 hidden sm:table-cell">Barcode</th>
              <th class="px-3 py-3 text-right">Beli</th>
              <th class="px-3 py-3 text-right">Jual</th>
              <th class="px-3 py-3 text-right">Stok</th>
              <th class="px-4 py-3 text-right w-36">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-50">
            ${list.map(p => `
              <tr class="hover:bg-gray-50/80 transition-colors">
                <td class="px-4 py-3">
                  <p class="font-semibold text-gray-800">${String(p.name||'').replace(/</g,'&lt;')}</p>
                  <p class="text-[11px] text-gray-400 sm:hidden">${String(p.barcode||'').replace(/</g,'&lt;')}</p>
                </td>
                <td class="px-3 py-3 text-xs font-mono text-gray-600 hidden sm:table-cell">${String(p.barcode||'').replace(/</g,'&lt;')}</td>
                <td class="px-3 py-3 text-right text-xs">Rp ${UI.fmt(p.buy_price||0)}</td>
                <td class="px-3 py-3 text-right font-semibold text-seller">Rp ${UI.fmt(p.sell_price||0)}</td>
                <td class="px-3 py-3 text-right">${p.stock ?? 0}</td>
                <td class="px-4 py-3 text-right whitespace-nowrap">
                  <button type="button" onclick="SellerProducts.openModal(${p.id})"
                    class="text-xs font-semibold text-seller hover:underline mr-2">Edit</button>
                  <button type="button" onclick="SellerProducts.remove(${p.id})"
                    class="text-xs font-semibold text-red-500 hover:underline">Hapus</button>
                </td>
              </tr>`).join('')}
          </tbody>
        </table>
      </div>`;
  }

  function openModal(productId) {
    _editId = (typeof productId === 'number' && productId > 0) ? productId : null;
    document.getElementById('spf-err').classList.add('hidden');
    document.getElementById('sp-modal-title').textContent = _editId ? '✏️ Edit Produk' : '➕ Tambah Produk';
    const barcodeEl = document.getElementById('spf-barcode');
    barcodeEl.disabled = false;

    if (_editId) {
      barcodeEl.disabled = false;
      (async () => {
        const res = await API.pub('GET', `/shop/products/${_editId}`);
        const p   = res.data;
        if (!res.success || !p) {
          UI.toast(res.message || 'Produk tidak ditemukan', true);
          return;
        }
        barcodeEl.value = p.barcode || '';
        document.getElementById('spf-name').value    = p.name || '';
        document.getElementById('spf-unit').value      = p.unit || 'pcs';
        document.getElementById('spf-buy').value       = p.buy_price ?? '';
        document.getElementById('spf-sell').value      = p.sell_price ?? '';
        document.getElementById('spf-min').value       = p.min_stock ?? 5;
        document.getElementById('spf-stock').value     = p.stock ?? 0;
        document.getElementById('spf-cat').value       = p.category_id || 1;
        document.getElementById('spf-sup').value       = p.supplier_id || 1;
        document.getElementById('sp-modal').classList.remove('hidden');
      })();
    } else {
      barcodeEl.value = '';
      document.getElementById('spf-name').value    = '';
      document.getElementById('spf-unit').value    = 'pcs';
      document.getElementById('spf-buy').value      = '';
      document.getElementById('spf-sell').value     = '';
      document.getElementById('spf-min').value      = '5';
      document.getElementById('spf-stock').value    = '0';
      document.getElementById('spf-cat').value      = '1';
      document.getElementById('spf-sup').value      = '1';
      document.getElementById('sp-modal').classList.remove('hidden');
    }
  }

  function closeModal() {
    document.getElementById('sp-modal').classList.add('hidden');
    _editId = null;
  }

  async function save() {
    const errEl = document.getElementById('spf-err');
    errEl.classList.add('hidden');
    const body = {
      category_id: Number(document.getElementById('spf-cat').value) || 1,
      supplier_id: Number(document.getElementById('spf-sup').value) || 1,
      barcode:     document.getElementById('spf-barcode').value.trim(),
      name:        document.getElementById('spf-name').value.trim(),
      unit:        document.getElementById('spf-unit').value.trim() || 'pcs',
      buy_price:   Number(document.getElementById('spf-buy').value),
      sell_price:  Number(document.getElementById('spf-sell').value),
      min_stock:   Number(document.getElementById('spf-min').value) || 0,
      stock:       Number(document.getElementById('spf-stock').value) || 0,
    };
    if (!body.barcode || !body.name) {
      errEl.textContent = 'Barcode dan nama wajib diisi';
      errEl.classList.remove('hidden'); return;
    }
    const btn = document.getElementById('btn-spf-save');
    const wasEdit = !!_editId;
    btn.disabled = true; btn.textContent = 'Menyimpan…';
    let res;
    if (_editId) res = await API.seller('PUT', `/seller/products/${_editId}`, body);
    else         res = await API.seller('POST', '/seller/products', body);
    btn.disabled = false; btn.textContent = 'Simpan';
    if (!res.success) {
      errEl.textContent = res.message || 'Gagal menyimpan';
      errEl.classList.remove('hidden'); return;
    }
    closeModal();
    UI.toast(wasEdit ? 'Produk diperbarui' : 'Produk ditambahkan');
    load();
  }

  async function remove(id) {
    if (!confirm('Nonaktifkan produk ini dari katalog toko?')) return;
    const res = await API.seller('DELETE', `/seller/products/${id}`);
    if (!res.success) { UI.toast(res.message || 'Gagal menghapus', true); return; }
    UI.toast('Produk dinonaktifkan');
    load();
  }

  return { load, openModal, closeModal, save, remove };
})();