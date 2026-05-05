/**
 * k6 — healthz, produk publik, dan opsional POST checkout (member JWT).
 *
 * Dasar:
 *   make stress-k6
 *
 * Checkout (pilih salah satu cara token):
 *   K6_SHOP_TOKEN=eyJ... K6_CHECKOUT_RATIO=0.05 make stress-k6
 *   K6_MEMBER_CODE=MBR001 K6_MEMBER_PHONE=081234567890 K6_CHECKOUT_RATIO=0.03 make stress-k6
 *
 * Variabel opsional checkout:
 *   K6_CHECKOUT_RATIO      — fraksi iterasi ke checkout (0 = nonaktif). Contoh: 0.05 = ~5%
 *   K6_CHECKOUT_PRODUCT_ID — default 1
 *   K6_CHECKOUT_QTY        — default 1
 *   K6_CHECKOUT_PAY_METHOD — default qris (pending; tetap memanggil Midtrans). cash = settle + kurangi stok.
 *   K6_CHECKOUT_ADDRESS_ID — default 0
 *   K6_CHECKOUT_VOUCHER_CODE — default kosong
 *
 * Peringatan: qris banyak kali = banyak panggilan Midtrans + transaksi pending (stok belum turun).
 * cash = stok habis cepat. Gunakan DB staging & ratio kecil.
 */

import http from 'k6/http';
import { check, sleep } from 'k6';

const defaultBase = __ENV.BASE_URL || 'http://127.0.0.1:8080';

const profiles = {
  smoke: {
    stages: [
      { duration: '10s', target: 5 },
      { duration: '30s', target: 5 },
      { duration: '5s', target: 0 },
    ],
  },
  medium: {
    stages: [
      { duration: '30s', target: 20 },
      { duration: '2m', target: 20 },
      { duration: '20s', target: 0 },
    ],
  },
  heavy: {
    stages: [
      { duration: '1m', target: 60 },
      { duration: '3m', target: 60 },
      { duration: '30s', target: 0 },
    ],
  },
};

const profileKey = (__ENV.K6_PROFILE || 'medium').toLowerCase();
const stages = profiles[profileKey]?.stages || profiles.medium.stages;

export const options = {
  stages,
  thresholds: {
    http_req_failed: ['rate<0.12'],
    http_req_duration: ['p(95)<3000'],
  },
};

export function setup() {
  const base = defaultBase;
  const preset = __ENV.K6_SHOP_TOKEN || '';
  if (preset) {
    return { base, token: preset };
  }
  const code = __ENV.K6_MEMBER_CODE || '';
  const phone = __ENV.K6_MEMBER_PHONE || '';
  if (!code || !phone) {
    return { base, token: null };
  }
  const res = http.post(
    `${base}/api/v1/auth/member-login`,
    JSON.stringify({ member_code: code, phone: phone }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  if (res.status !== 200) {
    console.warn(`setup member-login: HTTP ${res.status} body=${String(res.body).slice(0, 200)}`);
    return { base, token: null };
  }
  let j;
  try {
    j = res.json();
  } catch {
    console.warn('setup member-login: bad JSON');
    return { base, token: null };
  }
  const tok = j.data && j.data.token;
  if (!tok) {
    console.warn('setup member-login: tidak ada data.token');
    return { base, token: null };
  }
  return { base, token: tok };
}

function checkoutPayload() {
  const productId = parseInt(__ENV.K6_CHECKOUT_PRODUCT_ID || '1', 10);
  const qty = parseInt(__ENV.K6_CHECKOUT_QTY || '1', 10);
  const addressId = parseInt(__ENV.K6_CHECKOUT_ADDRESS_ID || '0', 10);
  return JSON.stringify({
    address_id: addressId,
    voucher_code: __ENV.K6_CHECKOUT_VOUCHER_CODE || '',
    items: [{ product_id: productId, quantity: qty }],
    pay_method: __ENV.K6_CHECKOUT_PAY_METHOD || 'qris',
  });
}

export default function (data) {
  const base = (data && data.base) || defaultBase;
  const token = data && data.token;
  const ratioRaw = __ENV.K6_CHECKOUT_RATIO || '0';
  const ratio = token ? parseFloat(ratioRaw) : 0;
  const checkoutOk = token && !isNaN(ratio) && ratio > 0;

  const u = Math.random();

  if (checkoutOk && u < ratio) {
    const res = http.post(`${base}/api/v1/shop/checkout`, checkoutPayload(), {
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      tags: { name: 'checkout' },
    });
    check(res, {
      'checkout 201': (r) => r.status === 201,
    });
    sleep(0.08 + Math.random() * 0.15);
    return;
  }

  const u2 = checkoutOk ? (u - ratio) / (1 - ratio) : u;
  if (u2 < 0.65) {
    const res = http.get(`${base}/healthz`, { tags: { name: 'healthz' } });
    check(res, { 'healthz 200': (r) => r.status === 200 });
  } else {
    const res = http.get(`${base}/api/v1/shop/products`, { tags: { name: 'shop_products' } });
    check(res, { 'shop products 200': (r) => r.status === 200 });
  }
  sleep(0.05 + Math.random() * 0.12);
}
