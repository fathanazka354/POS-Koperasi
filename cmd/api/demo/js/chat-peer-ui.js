'use strict';

/* Avatar & label lawan bicara (bukan produk) */
const ChatPeerUI = {
  initial(name) {
    const s = String(name == null ? '' : name).trim();
    if (!s) return '?';
    return s.charAt(0).toUpperCase();
  },
  avHtml(name, wClass, ring) {
    const ch = String(this.initial(name)).replace(/</g, '&lt;');
    const ringCls = ring === 'seller' ? 'bg-seller/15 text-seller' : 'bg-brand/15 text-brand';
    return `<div class="${wClass} rounded-full ${ringCls} flex items-center justify-center font-bold flex-none">${ch}</div>`;
  },
};
