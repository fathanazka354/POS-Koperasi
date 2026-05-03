# POS Koperasi Minimarket — Golang + Midtrans + MongoDB

Demo (Shop + Chat): `http://localhost:8080/demo/`
Alihkan: `/demo/shop` dan `/demo/chat` → `/demo/`

Dokumentasi API lengkap (request body, response, error): [`docs/API.md`](docs/API.md).

**Postman:** impor [`docs/openapi.yaml`](docs/openapi.yaml) lewat **Import** → pilih file tersebut; setelah import, sesuaikan **server** / variabel `port` agar sama dengan `APP_PORT` di `.env` (default di file: `8080`).

## Struktur Project

Arsitektur **per modul** (auth, chat, product, transaction): `contract` (interface), `repository/impl`, `service/impl`, `controller` + `controller/dto`, `router`, serta `domain` untuk entity. **Dependency injection** memakai **Uber Fx** (`go.uber.org/fx`: wiring di `cmd/api/inject.go`, lifecycle HTTP server + penutupan DB).

```
pos-koperasi/
├── cmd/api/
│   ├── main.go               ← Entry point: fx.New(...).Run()
│   └── inject.go             ← Fx: Provide / Invoke, router, HTTP server lifecycle
├── cmd/seed/
│   └── main.go               ← Seed DB (per entitas di internal/seed)
├── internal/
│   ├── config/config.go      ← Load .env, OpenDB / NewDB
│   ├── middleware/           ← JWT, member chat token, role
│   ├── model/                ← Model chat (thread / pesan) — sisanya domain per modul
│   ├── modules/
│   │   ├── auth/             ← domain, contract, service/impl, controller+dto, router
│   │   ├── chat/             ← + utility (WebSocket hub), repository/impl, …
│   │   ├── product/
│   │   ├── transaction/
│   │   ├── notification/     ← MongoDB + WebSocket push (notif bell)
│   │   ├── address/          ← Alamat pengiriman member (multi-address)
│   │   ├── voucher/          ← Kode diskon (percent / fixed)
│   │   └── shop/             ← Member checkout, riwayat pesanan
│   ├── seed/                 ← Seeder per entitas (branch, product, …)
│   └── midtrans/             ← Client Midtrans (Core /v2/charge), notification
├── migrations/001_init.sql   ← DDL + seed legacy (opsional)
├── pkg/response/response.go  ← Standard API response
├── .env.example
└── go.mod
```

---

## Setup

### 1. Clone & install

```bash
git clone https://github.com/yourname/pos-koperasi
cd pos-koperasi
cp .env.example .env
go mod tidy
```

### 2. Setup database

```bash
createdb pos_koperasi
make migrate
make seed
```

Perintah DB yang tersedia:

```bash
make migrate  # jalankan schema migration
make seed     # isi data awal: go run ./cmd/seed (per entitas di internal/seed)
make fresh    # drop schema public, migrate, lalu seed
make reset    # jalankan skrip legacy migrations/001_init.sql
```

### 3. Setup Midtrans

1. Buat / cek akun Midtrans lalu ambil **Server Key** (Sandbox/Test Mode).
2. Di Midtrans dashboard, set **Payment Notification URL** ke:
   `https://yourdomain.com/api/v1/midtrans/webhook`
3. Isi `.env`:

```env
MIDTRANS_SERVER_KEY=YOUR_MIDTRANS_SERVER_KEY
MIDTRANS_BASE_URL=https://api.sandbox.midtrans.com
```

> Untuk development lokal, gunakan **ngrok**:
> ```bash
> ngrok http 8070
> # Copy URL ngrok → pasang di Midtrans notification URL
> ```

### 4. Run server

```bash
go run ./cmd/api
# atau: make run
# Server running on :8080 [development]
```

---

## Testing Chat (WebSocket) — Demo Shopee/Tokopedia Style

Fitur chat ini mensimulasikan alur marketplace: **pembeli (member)** chat ke **penjual (karyawan)** dalam 1 thread per **produk**.

### Prasyarat

1. Pastikan DB sudah dimigrasi + seed (termasuk tabel chat):

```bash
make fresh
# atau minimal:
make migrate
make seed
```

2. Jalankan server:

```bash
make run
```

### Buka halaman demo

Buka URL (sesuaikan port dengan `APP_PORT` di `.env` / output terminal):

- `http://localhost:8080/demo/chat`
- atau `http://127.0.0.1:8080/demo/chat`

### Cara uji chat end-to-end

1. **Login member (pembeli)** di panel kiri:
   - `member_code`: `MBR001`
   - `phone`: `081234567890`
2. **Login karyawan (penjual)** di panel kanan:
   - `NIK`: `EMP001`
   - `PIN`: `1234`
3. Di panel member, isi:
   - `product_id`: `1`
   - `seller_employee_id`: `1`
   lalu klik **Buat / ambil percakapan** → akan muncul `conversation_id`.
4. Klik **Connect WebSocket** untuk **pembeli** dan **penjual**.
5. Klik **Join room** di **keduanya** (harus `conversation_id` yang sama).
6. Ketik pesan lalu klik **Kirim** — pesan akan muncul realtime di kedua panel, dan juga tersimpan di DB.

### Catatan penting

- Endpoint WebSocket: `GET /api/v1/ws/chat?token=JWT` (token diambil dari hasil login).
- Jika hanya salah satu side yang **join**, maka broadcast tidak akan terlihat di sisi yang belum join.

---

## Flow Hit API Lengkap

### STEP 1 — Login kasir

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "nik": "EMP001",
  "pin": "1234"
}
```

Response:
```json
{
  "success": true,
  "message": "Login berhasil",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "employee": {
      "id": 1,
      "full_name": "Kasir Satu",
      "role": "cashier",
      "branch_id": 1
    }
  }
}
```

Simpan `token` untuk request selanjutnya.

---

### STEP 2 — Scan barcode produk

```http
GET /api/v1/products/barcode/8999999001
Authorization: Bearer {token}
```

Response:
```json
{
  "success": true,
  "message": "Produk ditemukan",
  "data": {
    "id": 1,
    "barcode": "8999999001",
    "name": "Indomie Goreng",
    "sell_price": 3500,
    "unit": "pcs",
    "stock": 100
  }
}
```

---

### STEP 3A — Buat transaksi (Cash)

```http
POST /api/v1/transactions
Authorization: Bearer {token}
Content-Type: application/json

{
  "branch_id": 1,
  "register_id": 1,
  "member_code": "MBR001",
  "pay_method": "cash",
  "items": [
    { "product_id": 1, "quantity": 3 },
    { "product_id": 2, "quantity": 2 }
  ]
}
```

Response (langsung paid):
```json
{
  "success": true,
  "message": "Pembayaran cash berhasil",
  "data": {
    "transaction": {
      "id": 1,
      "invoice_no": "INV-20240115143022-1",
      "subtotal": 16500,
      "tax": 1815,
      "grand_total": 18315,
      "status": "paid"
    }
  }
}
```

---

### STEP 3B — Buat transaksi (QRIS)

```http
POST /api/v1/transactions
Authorization: Bearer {token}
Content-Type: application/json

{
  "branch_id": 1,
  "register_id": 1,
  "pay_method": "qris",
  "items": [
    { "product_id": 1, "quantity": 2 }
  ]
}
```

Response (pending, tunggu webhook):
```json
{
  "success": true,
  "message": "Scan QR untuk menyelesaikan pembayaran",
  "data": {
    "transaction": {
      "id": 2,
      "invoice_no": "INV-20240115143500-1",
      "grand_total": 7777,
      "status": "pending"
    },
    "qr_string": "00020101021226600014ID.CO.XENDIT.WWW...",
    "midtrans_transaction_id": "qr_123456789"
  }
}
```

→ Tampilkan `qr_string` sebagai QR Code di layar kasir.

---

### STEP 3C — Buat transaksi (Virtual Account)

```http
POST /api/v1/transactions
Authorization: Bearer {token}
Content-Type: application/json

{
  "pay_method": "va",
  "items": [{ "product_id": 2, "quantity": 5 }]
}
```

Response:
```json
{
  "data": {
    "transaction": { "status": "pending", "grand_total": 16650 },
    "va_number": "8808999991234567",
    "midtrans_transaction_id": "va_987654321"
  }
}
```

---

### STEP 3D — Buat transaksi (E-wallet / GoPay - Midtrans)

```http
POST /api/v1/transactions
Authorization: Bearer {token}
Content-Type: application/json

{
  "pay_method": "ewallet",
  "items": [{ "product_id": 3, "quantity": 1 }]
}
```

Response:
```json
{
  "data": {
    "transaction": { "status": "pending", "grand_total": 8325 },
    "payment_url": "payment_url/deeplink dari Midtrans",
    "midtrans_transaction_id": "inv_abcdef123"
  }
}
```

→ Redirect pelanggan ke `payment_url`.

---

### STEP 4 — Midtrans kirim Notification (otomatis setelah bayar)

Midtrans POST ke `POST /api/v1/midtrans/webhook`:
```json
{
  "id": "inv_abcdef123",
  "external_id": "INV-20240115143500-1",
  "status": "PAID",
  "payment_method": "QRIS",
  "amount": 7777,
  "paid_amount": 7777,
  "paid_at": "2024-01-15T14:37:00.000Z"
}
```

Yang terjadi di backend (1 DB transaction atomic):
1. `transactions` → status `pending` → `paid`
2. `payments` → INSERT record baru
3. `stocks` → quantity dikurangi per item
4. `members` → point_balance ditambah (jika ada)
5. `point_histories` → INSERT earn

---

## Penting: Idempotency

Midtrans bisa mengirim notification yang sama lebih dari 1x.
Sistem sudah handle ini dengan cek `reference_no` di tabel `payments` (kolom UNIQUE).

```sql
-- Jika reference_no sudah ada → skip, balas 200
SELECT id FROM payments WHERE reference_no = 'inv_abcdef123';
```

---

## Testing Webhook Lokal (ngrok)

```bash
# Terminal 1: jalankan server
go run ./cmd/api

# Terminal 2: expose ke internet
ngrok http 8080

# Copy URL ngrok: https://abc123.ngrok.io
# Pasang di Midtrans Dashboard → Payment Notification URL
# URL: https://abc123.ngrok.io/api/v1/midtrans/webhook

# Simulate webhook manual
curl -X POST http://localhost:8080/api/v1/midtrans/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "INV-20240115143500-1",
    "transaction_id": "midtrans_test_001",
    "transaction_status": "settlement",
    "payment_type": "qris",
    "status_code": "200",
    "gross_amount": "7777.00",
    "signature_key": "PASTE_SIGNATURE_KEY_SESUAI_MIDTRANS"
  }'
```

---

## Demo Shop (Checkout + Notifikasi Internal)

Buka `http://localhost:8080/demo/shop` di browser.

### Prasyarat tambahan

MongoDB harus berjalan (untuk notifikasi persisten). Jika tidak ada, sistem otomatis menggunakan
in-memory fallback — notifikasi tetap berfungsi selama session server aktif.

```bash
# Jalankan MongoDB lokal (jika belum)
brew services start mongodb-community   # macOS
# atau
docker run -d -p 27017:27017 mongo

# Tambahkan ke .env
MONGO_URI=mongodb://localhost:27017
MONGO_DB_NAME=pos_koperasi

# Jalankan migrasi shop (alamat + voucher)
make migrate

# Jalankan server
make run
```

### Alur demo

1. Buka `http://localhost:8080/demo/shop`
2. Browse produk di halaman utama — tidak perlu login
3. Klik produk → halaman detail → atur jumlah → Tambah ke Keranjang
4. Klik ikon keranjang 🛍️ → Checkout
5. Jika belum login, muncul modal login:
   - Kode Anggota: `MBR001`
   - No. Telepon: `081234567890`
6. Setelah login, masuk halaman checkout 3 langkah:
   - Langkah 1 — Pilih/tambah alamat pengiriman
   - Langkah 2 — Input kode voucher (coba `HEMAT10` atau `HEMAT5K`), lihat ringkasan
   - Langkah 3 — Pilih metode bayar (cash/QRIS/VA/GoPay) → Bayar
7. Cash: langsung confirmed + notifikasi terkirim
   QRIS/VA: tampil QR/nomor VA → klik "Cek Status" atau tunggu webhook Midtrans
8. Klik ikon lonceng 🔔 di header → daftar notifikasi
   - Notifikasi belum dibaca: background lebih gelap (merah muda)
   - Klik notifikasi → diarahkan ke detail transaksi
9. WebSocket real-time: bell badge muncul otomatis tanpa refresh jika notif masuk

### API shop yang tersedia

| Method | Endpoint | Keterangan |
|--------|----------|------------|
| GET | `/api/v1/shop/products` | Daftar produk (public, `?search=`) |
| GET | `/api/v1/shop/products/{id}` | Detail produk (public) |
| POST | `/api/v1/shop/checkout` | Checkout member (JWT member) |
| GET | `/api/v1/shop/orders` | Riwayat pesanan member |
| GET | `/api/v1/shop/orders/{invoiceNo}` | Detail pesanan |
| GET | `/api/v1/shop/orders/{invoiceNo}/status` | Cek status pembayaran |
| GET | `/api/v1/member/addresses` | Daftar alamat member |
| POST | `/api/v1/member/addresses` | Tambah alamat |
| PUT | `/api/v1/member/addresses/{id}/default` | Set alamat default |
| GET | `/api/v1/member/notifications` | Daftar notifikasi (MongoDB) |
| PUT | `/api/v1/member/notifications/{id}/read` | Tandai dibaca |
| PUT | `/api/v1/member/notifications/read-all` | Tandai semua dibaca |
| GET | `/api/v1/ws/notify?token=JWT` | WebSocket push notifikasi |

### Voucher demo yang tersedia

| Kode | Jenis | Nilai | Min. Belanja |
|------|-------|-------|--------------|
| `HEMAT10` | Persen | 10% (maks Rp 50.000) | Rp 10.000 |
| `HEMAT5K` | Nominal | Rp 5.000 | Rp 20.000 |

