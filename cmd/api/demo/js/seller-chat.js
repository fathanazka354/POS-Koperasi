'use strict';

/* ─── Seller: Chat ───────────────────────────────────── */
const SellerChat = (() => {
  let _ws = null, _activeId = 0;
  let _convMap = {};
  let _peerPoll = null;
  const skipEcho = {};

  function makeChatWsUrl(tok) {
    return Config.BASE.replace(/^http/,'ws') + '/api/v1/ws/chat?token=' + encodeURIComponent(tok);
  }

  function setWsStatus(s) {
    const el = document.getElementById('seller-ws-status'); if (!el) return;
    const map = {
      connecting: ['Connecting…', 'bg-white/20 text-white/70'],
      online:     ['● Online',    'bg-white/30 text-white'],
      offline:    ['Offline',     'bg-red-400/80 text-white'],
    };
    const [txt, cls] = map[s] || map.connecting;
    el.textContent = txt; el.className = `text-[10px] font-semibold px-2 py-0.5 rounded-full ml-1 ${cls}`;
  }

  function _applyPeerDot(online) {
    const dot = document.getElementById('sc-peer-dot');
    if (!dot) return;
    if (online === undefined) {
      dot.className = 'w-2 h-2 rounded-full shrink-0 bg-gray-300 animate-pulse';
      dot.classList.remove('invisible');
      return;
    }
    if (online) {
      dot.className = 'w-2 h-2 rounded-full shrink-0 bg-emerald-500';
    } else {
      dot.className = 'w-2 h-2 rounded-full shrink-0 invisible';
    }
  }

  function stopPeerPoll() {
    if (_peerPoll) { clearInterval(_peerPoll); _peerPoll = null; }
  }

  async function syncPeerPresence() {
    if (!_activeId) return;
    if (!document.getElementById('sc-peer-dot')) return;
    const res = await API.seller('GET', '/chat/conversations/' + _activeId + '/presence');
    if (!res.success || res.data == null) { _applyPeerDot(false); return; }
    _applyPeerDot(!!res.data.buyer_online);
  }

  function startPeerPoll() {
    stopPeerPoll();
    syncPeerPresence();
    _peerPoll = setInterval(syncPeerPresence, 5000);
  }

  async function init() {
    setWsStatus('connecting');
    await _connect();
  }

  async function _connect() {
    if (!SS.token) return;
    if (_ws) try { _ws.close(); } catch {}
    _ws = new WebSocket(makeChatWsUrl(SS.token));
    _ws.onopen    = () => { setWsStatus('online'); _autoJoinAll(); };
    _ws.onmessage = ev => _onMessage(ev);
    _ws.onclose   = () => { setWsStatus('offline'); if (SS.token) setTimeout(_connect, 4000); };
    _ws.onerror   = () => { setWsStatus('offline'); };
  }

  async function _autoJoinAll() {
    const res  = await API.seller('GET', '/chat/conversations');
    const list = Array.isArray(res.data) ? res.data : [];
    list.forEach(c => {
      if (_ws && _ws.readyState===1) _ws.send(JSON.stringify({ type:'join', conversation_id: c.id }));
    });
    _renderConvList(list);
  }

  /** Badge di nav 💬 Chat — sama konsepnya dengan member-chat-unread-badge pembeli */
  function _maybeBumpSellerNavBadge(cid, m) {
    if (!m || m.sender_role !== 'buyer') return;
    const prodEl = document.getElementById('seller-view-products');
    const chatEl = document.getElementById('seller-view-chat');
    const onProductsTab = prodEl && !prodEl.classList.contains('hidden');
    const chatPanelHidden = chatEl && chatEl.classList.contains('hidden');
    const otherThread = cid !== _activeId;
    if (!onProductsTab && !chatPanelHidden && !otherThread) return;
    const el = document.getElementById('seller-chat-badge');
    if (!el) return;
    const cur = parseInt(el.textContent, 10) || 0;
    el.textContent = cur + 1;
    el.classList.remove('hidden');
  }

  function clearNavBadge() {
    const el = document.getElementById('seller-chat-badge');
    if (!el) return;
    el.textContent = '0';
    el.classList.add('hidden');
  }

  function _setSidebarPreview(cid, text) {
    const item = document.getElementById('scc-' + cid);
    if (!item || text == null || text === '') return;
    const el = item.querySelector('.sc-last-preview');
    if (!el) return;
    const t = String(text);
    el.textContent = t.length > 100 ? t.slice(0, 97) + '…' : t;
  }

  function _renderConvList(list) {
    const el = document.getElementById('sc-conv-list'); if (!el) return;
    if (!list.length) {
      el.innerHTML = `<div class="py-10 flex flex-col items-center gap-2 text-gray-300 px-4"><div class="text-4xl">💬</div><p class="text-xs text-gray-400 text-center">Belum ada percakapan dari pembeli.</p></div>`;
      return;
    }
    el.innerHTML = list.map(c=>{
      _convMap[c.id] = c;
      const active = c.id === _activeId;
      const peer = c.member_name || 'Pembeli';
      const peerEsc = String(peer).replace(/</g, '&lt;');
      const thumb = ChatPeerUI.avHtml(peer, 'w-10 h-10 text-sm', 'seller');
      return `<div id="scc-${c.id}" onclick="SellerChat.openConv(${c.id})"
        class="px-4 py-3.5 flex gap-3 cursor-pointer border-b border-gray-100 transition-colors ${active?'bg-seller-light border-l-4 border-l-seller':'hover:bg-gray-50'}">
        ${thumb}
        <div class="flex-1 min-w-0">
          <p class="text-sm font-semibold truncate">${peerEsc}</p>
          <p class="sc-last-preview text-xs text-gray-400 truncate mt-0.5">${c.last_message||'Belum ada pesan'}</p>
        </div>
        <div class="flex flex-col items-end gap-1 flex-none">
          ${(c.unread_count||0)>0?`<span class="sc-unread bg-seller text-white text-[10px] font-bold rounded-full w-5 h-5 flex items-center justify-center">${c.unread_count}</span>`:''}
        </div>
      </div>`;
    }).join('');

    document.getElementById('sc-search').oninput = e => {
      const q = e.target.value.toLowerCase();
      document.querySelectorAll('#sc-conv-list [id^="scc-"]').forEach(el=>{
        el.style.display = el.textContent.toLowerCase().includes(q)?'':'none';
      });
    };
  }

  async function openConv(id) {
    let conv = _convMap[id];
    if (!conv || conv.product_id == null) {
      const fres = await API.seller('GET', '/chat/conversations');
      const list = Array.isArray(fres.data) ? fres.data : [];
      list.forEach(c => { _convMap[c.id] = c; });
      conv = _convMap[id] || {};
    }

    _activeId = id;
    document.getElementById('sc-empty').classList.add('hidden');
    document.getElementById('sc-active').classList.remove('hidden');

    const buyerNm = conv.member_name ? String(conv.member_name).replace(/</g,'&lt;') : 'Pembeli';
    const avHead = ChatPeerUI.avHtml(buyerNm, 'w-9 h-9 text-sm', 'seller');

    document.getElementById('sc-header').innerHTML = `
      ${avHead}
      <div class="flex-1 min-w-0 flex items-center gap-2">
        <span class="text-sm font-bold text-gray-900 truncate">${buyerNm}</span>
        <span id="sc-peer-dot" class="w-2 h-2 rounded-full shrink-0 bg-gray-300 animate-pulse" aria-hidden="true"></span>
      </div>`;

    // Highlight sidebar
    document.querySelectorAll('#sc-conv-list [id^="scc-"]').forEach(el=>{
      el.className = el.className.replace(/bg-seller-light border-l-4 border-l-seller/g,'').replace(/hover:bg-gray-50/g,'') + ' hover:bg-gray-50';
      if (el.id===`scc-${id}`) {
        el.className = el.className.replace('hover:bg-gray-50','') + ' bg-seller-light border-l-4 border-l-seller';
        const b = el.querySelector('.sc-unread'); if(b) b.remove();
      }
    });

    if (_ws && _ws.readyState===1) _ws.send(JSON.stringify({ type:'join', conversation_id: id }));

    const area = document.getElementById('sc-messages');
    area.innerHTML = `<div class="text-center text-xs text-gray-400 skeleton py-4">Memuat pesan…</div>`;
    const res = await API.seller('GET', `/chat/conversations/${id}/messages`);
    area.innerHTML = '';
    const rows = res.success && Array.isArray(res.data) ? res.data : [];
    for (const m of rows) {
      if (m.product_id) {
        await ChatProductCard.prepend(area, null, { id: m.product_id }, { seller: true });
      } else if (String(m.body || '').trim()) {
        _appendBubble(m.body, m.sender_role, UI.timeFmt(m.created_at));
      }
    }
    const hasContent = area.querySelector('.bubble-me, .bubble-you, button[aria-label]');
    if (!hasContent && !rows.length) {
      area.insertAdjacentHTML('beforeend', `<p class="text-center text-xs text-gray-400 py-4">Belum ada pesan dari pembeli.</p>`);
    }
    startPeerPoll();
  }

  function _appendBubble(text, role, time) {
    const area = document.getElementById('sc-messages'); if (!area) return;
    const isMe = role === 'seller';
    const empty = area.querySelector('p.text-center'); if (empty) empty.remove();
    const wrap  = document.createElement('div');
    wrap.className = `flex ${isMe?'justify-end':'justify-start'} fade-in`;
    wrap.innerHTML = `<div class="max-w-[70%]">
      <div class="px-4 py-2.5 text-sm leading-snug rounded-2xl ${isMe?'rounded-br-sm bubble-me':'rounded-bl-sm bubble-you'}">${text.replace(/</g,'&lt;')}</div>
      ${time?`<p class="text-[10px] text-gray-400 mt-0.5 ${isMe?'text-right':''}">${time}</p>`:''}
    </div>`;
    area.appendChild(wrap);
    area.scrollTop = area.scrollHeight;
  }

  function _onMessage(ev) {
    let msg; try { msg = JSON.parse(ev.data); } catch { return; }
    if (msg?.type !== 'message' || !msg.message) return;
    const m   = msg.message;
    const cid = m.conversation_id;

    // Re-join if new conversation from buyer
    if (!document.getElementById(`scc-${cid}`)) {
      if (_ws && _ws.readyState===1) _ws.send(JSON.stringify({ type:'join', conversation_id: cid }));
      _autoJoinAll(); // Refresh list
      _maybeBumpSellerNavBadge(cid, m);
      return;
    }

    if (cid === _activeId) {
      const area = document.getElementById('sc-messages');
      if (m.product_id && area) {
        void ChatProductCard.prepend(area, null, { id: m.product_id }, { seller: true });
        _setSidebarPreview(cid, '📎 Produk');
        syncPeerPresence();
        _maybeBumpSellerNavBadge(cid, m);
        return;
      }
      const sk = skipEcho[cid];
      if (sk && sk.body===m.body && Date.now()-sk.t<8000) { skipEcho[cid]=null; return; }
      _appendBubble(m.body, m.sender_role, UI.timeFmt(m.created_at));
      _setSidebarPreview(cid, m.body);
      syncPeerPresence();
      _maybeBumpSellerNavBadge(cid, m);
    } else {
      _setSidebarPreview(cid, m.product_id ? '📎 Produk' : m.body);
      // Show badge on conv list item
      const item = document.getElementById(`scc-${cid}`);
      if (item) {
        let badge = item.querySelector('.sc-unread');
        if (!badge) { badge=document.createElement('span'); badge.className='sc-unread bg-seller text-white text-[10px] font-bold rounded-full w-5 h-5 flex items-center justify-center'; const d=item.querySelector('.flex.flex-col'); if(d) d.appendChild(badge); }
        badge.textContent=(parseInt(badge.textContent)||0)+1;
      }
      _maybeBumpSellerNavBadge(cid, m);
      UI.toast(`💬 Pesan baru di percakapan #${cid}`, false, 4000);
    }
  }

  function send() {
    const inp  = document.getElementById('sc-input');
    const body = inp?.value.trim();
    if (!body || !_activeId) return;
    if (!_ws || _ws.readyState!==1) { UI.toast('Koneksi terputus…', true); _connect(); return; }
    skipEcho[_activeId] = { body, t: Date.now() };
    _appendBubble(body, 'seller', new Date().toLocaleTimeString('id-ID',{hour:'2-digit',minute:'2-digit'}));
    _ws.send(JSON.stringify({ type:'message', conversation_id: _activeId, body }));
    _setSidebarPreview(_activeId, body);
    if (inp) inp.value = '';
  }

  function disconnectWS() { stopPeerPoll(); if (_ws) { _ws.close(); _ws = null; } }

  return { init, openConv, send, disconnectWS, clearNavBadge };
})();