'use strict';

/* ─── Member: Chat ───────────────────────────────────── */
const MemberChat = (() => {
  let _ws = null, _activeId = 0, _convMap = {};
  /** Draft dari Shop: belum ada POST — kartu produk hanya di UI sampai kirim pesan pertama */
  let _draftChat = null;
  let _peerPoll = null;
  const skipEcho = {};

  function makeChatWsUrl(tok) {
    return Config.BASE.replace(/^http/,'ws') + '/api/v1/ws/chat?token=' + encodeURIComponent(tok);
  }

  function _applyPeerDot(online) {
    const dot = document.getElementById('mc-peer-dot');
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
    if (!document.getElementById('mc-peer-dot')) return;
    const res = await API.member('GET', '/chat/conversations/' + _activeId + '/presence');
    if (!res.success || res.data == null) { _applyPeerDot(false); return; }
    _applyPeerDot(!!res.data.seller_online);
  }

  function startPeerPoll() {
    stopPeerPoll();
    syncPeerPresence();
    _peerPoll = setInterval(syncPeerPresence, 5000);
  }

  function connectWS() {
    if (!MS.token) return;
    if (_ws) try { _ws.close(); } catch {}
    _ws = new WebSocket(makeChatWsUrl(MS.token));
    _ws.onopen    = () => { _autoJoinAll(); };
    _ws.onmessage = ev => _onMessage(ev);
    _ws.onclose   = () => { if (MS.token) setTimeout(connectWS, 4000); };
  }

  function disconnectWS() { stopPeerPoll(); if (_ws) { _ws.close(); _ws = null; } }

  /** Badge tab 💬 Chat — unread pesan penjual (pisah dari badge pending order di Shop) */
  function _maybeBumpMemberNavUnread(cid, m) {
    if (!m || m.sender_role !== 'seller') return;
    const shop = document.getElementById('member-shop');
    const onShopTab = shop && !shop.classList.contains('hidden');
    const otherConv = cid !== _activeId;
    if (!onShopTab && !otherConv) return;
    const el = document.getElementById('member-chat-unread-badge');
    if (!el) return;
    const cur = parseInt(el.textContent, 10) || 0;
    el.textContent = cur + 1;
    el.classList.remove('hidden');
  }

  async function _autoJoinAll() {
    const res  = await API.member('GET', '/chat/conversations');
    const list = Array.isArray(res.data) ? res.data : [];
    list.forEach(c => {
      _convMap[c.id] = c;
      if (_ws && _ws.readyState === 1) _ws.send(JSON.stringify({ type:'join', conversation_id: c.id }));
    });
  }

  function _onMessage(ev) {
    let msg; try { msg = JSON.parse(ev.data); } catch { return; }
    if (msg?.type !== 'message' || !msg.message) return;
    const m   = msg.message;
    const cid = m.conversation_id;
    if (cid === _activeId) {
      const area = document.getElementById('mc-messages');
      if (m.product_id && area) {
        void ChatProductCard.prepend(area, null, { id: m.product_id }, {});
        syncPeerPresence();
        _maybeBumpMemberNavUnread(cid, m);
        return;
      }
      const sk = skipEcho[cid];
      if (sk && sk.body === m.body && Date.now()-sk.t < 8000) { skipEcho[cid]=null; return; }
      _appendBubble(m.body, m.sender_role, UI.timeFmt(m.created_at));
      syncPeerPresence();
      _maybeBumpMemberNavUnread(cid, m);
      return;
    }
    // cid !== _activeId
    const item = document.getElementById(`mcc-${cid}`);
    if (item) {
      let badge = item.querySelector('.mc-unread');
      if (!badge) { badge = document.createElement('span'); badge.className = 'mc-unread bg-brand text-white text-[10px] font-bold rounded-full w-5 h-5 flex items-center justify-center flex-none mt-1'; item.appendChild(badge); }
      badge.textContent = (parseInt(badge.textContent)||0)+1;
    }
    _maybeBumpMemberNavUnread(cid, m);
  }

  async function onOpen() {
    if (!MS.token) return;
    await loadConvList();
    if (_ws && _ws.readyState !== 1) connectWS();
    if (_activeId === 0 && !_draftChat) {
      stopPeerPoll();
      document.getElementById('mc-empty').classList.remove('hidden');
      document.getElementById('mc-active').classList.add('hidden');
    }
  }

  async function loadConvList() {
    const el = document.getElementById('mc-conv-list'); if (!el) return;
    el.innerHTML = `<div class="py-6 text-center text-xs text-gray-400 skeleton">Memuat…</div>`;
    const res  = await API.member('GET', '/chat/conversations');
    const list = Array.isArray(res.data) ? res.data : [];
    if (!list.length) {
      el.innerHTML = `<div class="py-10 flex flex-col items-center gap-2 text-gray-300 px-4"><div class="text-4xl">💬</div><p class="text-xs text-center text-gray-400">Belum ada percakapan. Buka <b>Shop</b>, pilih produk, lalu <b>Chat dengan penjual</b>.</p></div>`;
      return;
    }
    el.innerHTML = list.map(c => {
      _convMap[c.id] = c;
      const active = _activeId > 0 && c.id === _activeId;
      const peer = c.seller_name || 'Penjual';
      const peerEsc = String(peer).replace(/</g, '&lt;');
      const thumb = ChatPeerUI.avHtml(peer, 'w-10 h-10 text-sm', 'member');
      return `<div id="mcc-${c.id}" onclick="MemberChat.openConv(${c.id})"
        class="px-4 py-3 flex gap-3 cursor-pointer border-b border-gray-50 transition-colors ${active?'bg-brand-light border-l-4 border-l-brand':'hover:bg-gray-50'}">
        ${thumb}
        <div class="flex-1 min-w-0">
          <p class="text-sm font-semibold truncate">${peerEsc}</p>
          <p class="text-xs text-gray-400 truncate mt-0.5">${c.last_message||'Belum ada pesan'}</p>
        </div>
        ${(c.unread_count||0)>0?`<span class="mc-unread bg-brand text-white text-[10px] font-bold rounded-full w-5 h-5 flex items-center justify-center flex-none mt-1">${c.unread_count}</span>`:''}
      </div>`;
    }).join('');

    // search
    document.getElementById('mc-search').oninput = e => {
      const q = e.target.value.toLowerCase();
      document.querySelectorAll('#mc-conv-list [id^="mcc-"]').forEach(el=>{
        el.style.display = el.textContent.toLowerCase().includes(q)?'':'none';
      });
    };
  }

  async function openConv(id) {
    _draftChat = null;
    _activeId = id;
    document.getElementById('mc-empty').classList.add('hidden');
    document.getElementById('mc-active').classList.remove('hidden');

    let conv = _convMap[id];
    if (!conv || conv.seller_name == null || conv.seller_name === '') {
      const fres = await API.member('GET', '/chat/conversations');
      const list = Array.isArray(fres.data) ? fres.data : [];
      list.forEach(c => { _convMap[c.id] = { ..._convMap[c.id], ...c }; });
      conv = _convMap[id] || conv || {};
    }
    const sellerNm = (conv && conv.seller_name) ? String(conv.seller_name).replace(/</g,'&lt;') : 'Penjual';
    const avHead = ChatPeerUI.avHtml(sellerNm, 'w-9 h-9 text-sm', 'member');

    document.getElementById('mc-header').innerHTML = `
      ${avHead}
      <div class="flex-1 min-w-0 flex items-center gap-2">
        <span class="text-sm font-bold text-gray-900 truncate">${sellerNm}</span>
        <span id="mc-peer-dot" class="w-2 h-2 rounded-full shrink-0 bg-gray-300 animate-pulse" title="" aria-hidden="true"></span>
      </div>`;

    // Highlight sidebar
    document.querySelectorAll('#mc-conv-list [id^="mcc-"]').forEach(el=>{
      el.className = el.className.replace('bg-brand-light border-l-4 border-l-brand','hover:bg-gray-50').replace('hover:bg-gray-50 hover:bg-gray-50','hover:bg-gray-50');
      if (el.id===`mcc-${id}`) el.className = el.className.replace('hover:bg-gray-50','') + ' bg-brand-light border-l-4 border-l-brand';
      if (el.id===`mcc-${id}`) { const b=el.querySelector('.mc-unread'); if(b) b.remove(); }
    });

    if (_ws && _ws.readyState === 1) _ws.send(JSON.stringify({ type:'join', conversation_id: id }));

    const area = document.getElementById('mc-messages');
    area.innerHTML = `<div class="text-center text-xs text-gray-400 skeleton py-4">Memuat pesan…</div>`;
    const res = await API.member('GET', `/chat/conversations/${id}/messages`);
    area.innerHTML = '';
    const rows = res.success && Array.isArray(res.data) ? res.data : [];
    for (const m of rows) {
      if (m.product_id) {
        await ChatProductCard.prepend(area, null, { id: m.product_id }, {});
      } else if (String(m.body || '').trim()) {
        _appendBubble(m.body, m.sender_role, UI.timeFmt(m.created_at));
      }
    }
    if (!rows.length) {
      area.insertAdjacentHTML('beforeend', `<p class="text-center text-xs text-gray-400 py-4">Belum ada pesan. Mulai percakapan!</p>`);
    }
    startPeerPoll();
  }

  async function startFromCurrentProduct(sellerEmployeeId) {
    const emp = Number(sellerEmployeeId) || 1;
    if (!MS.token) { Landing.openLoginMember(); UI.toast('Login dulu untuk chat dengan penjual'); return; }
    const p = Shop._current;
    if (!p || !p.id) { UI.toast('Buka detail produk dulu', true); return; }
    _draftChat = { product_id: p.id, seller_employee_id: emp, snap: { id: p.id, name: p.name, sell_price: p.sell_price } };
    _activeId = 0;
    stopPeerPoll();
    switchTab('chat');
    document.getElementById('mc-empty').classList.add('hidden');
    document.getElementById('mc-active').classList.remove('hidden');
    const sellerNm = 'Penjual';
    const avHead = ChatPeerUI.avHtml(sellerNm, 'w-9 h-9 text-sm', 'member');
    document.getElementById('mc-header').innerHTML = `
      ${avHead}
      <div class="flex-1 min-w-0 flex items-center gap-2">
        <span class="text-sm font-bold text-gray-900 truncate">${sellerNm}</span>
        <span id="mc-peer-dot" class="w-2 h-2 rounded-full shrink-0 bg-gray-300 invisible" title="" aria-hidden="true"></span>
      </div>`;
    const area = document.getElementById('mc-messages');
    area.innerHTML = '';
    await ChatProductCard.prepend(area, null, _draftChat.snap, {});
    area.insertAdjacentHTML('beforeend', `<p class="text-center text-xs text-gray-400 py-3 px-4">Tulis pesan di bawah lalu kirim untuk menyimpan percakapan ke penjual.</p>`);
  }

  function _appendBubble(text, role, time) {
    const area  = document.getElementById('mc-messages'); if (!area) return;
    const isMe  = role === 'buyer';
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

  async function send() {
    const inp  = document.getElementById('mc-input');
    const body = inp?.value.trim();
    if (!body) return;

    if (_draftChat) {
      const res = await API.member('POST', '/chat/conversations', {
        product_id: _draftChat.product_id,
        seller_employee_id: _draftChat.seller_employee_id,
        first_message: body
      });
      if (!res.success || !res.data) { UI.toast(res.message || 'Gagal mengirim', true); return; }
      const conv = res.data;
      _draftChat = null;
      _convMap[conv.id] = { ..._convMap[conv.id], ...conv };
      _activeId = conv.id;
      if (inp) inp.value = '';
      if (_ws && _ws.readyState === 1) _ws.send(JSON.stringify({ type:'join', conversation_id: conv.id }));
      await loadConvList();
      await openConv(conv.id);
      return;
    }

    if (!_activeId) return;
    if (!_ws || _ws.readyState !== 1) { UI.toast('Koneksi terputus, coba lagi…', true); connectWS(); return; }
    skipEcho[_activeId] = { body, t: Date.now() };
    _appendBubble(body, 'buyer', new Date().toLocaleTimeString('id-ID',{hour:'2-digit',minute:'2-digit'}));
    _ws.send(JSON.stringify({ type:'message', conversation_id: _activeId, body }));
    if (inp) inp.value = '';
  }

  return { connectWS, disconnectWS, onOpen, loadConvList, openConv, startFromCurrentProduct, send };
})();