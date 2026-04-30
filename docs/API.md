# Dokumentasi API — POS Koperasi

Dokumen ini menggambarkan perilaku API **sesuai kode saat ini** (handler, service, `pkg/response`).

## Informasi umum

| Item | Nilai |
|------|--------|
| Base URL (lokal) | `http://localhost:{APP_PORT}` — default `APP_PORT=8080` |
| Prefix API | `/api/v1` |
| Format body | `application/json` (kecuali catatan khusus di webhook) |
| Autentikasi | `Authorization: Bearer <JWT>` untuk endpoint yang dilindungi |

---

## Format respons standar (JSON)

Mayoritas endpoint memakai struktur berikut.

### Sukses (`response.Success` / `response.Created`)

HTTP status **200** (`Success`) atau **201** (`Created`).

```json
{
  "success": true,
  "message": "string",
  "data": {}
}
```

Field `data` bisa berupa objek, array, atau `null` jika tidak ada data.

### Gagal (`response.BadRequest`, `Unauthorized`, `NotFound`)

HTTP status **400**, **401**, atau **404**. Payload JSON:

```json
{
  "success": false,
  "message": "string"
}
```

**Catatan:** Struct `APIResponse` memiliki field opsional `error`, tetapi handler saat ini **tidak mengisinya** — detail kesalahan ada di `message`.

### Respons yang bukan format di atas

- **Webhook Midtrans** (`POST /midtrans/webhook`) selalu mengembalikan JSON sederhana `{"status":"..."}` dengan status HTTP **200** (untuk menghindari retry berulang).

---

## Auth — JWT

### Klaim JWT (setelah login)

Token ditandatangani HS256. Klaim kustom (disematkan di context sebagai `EmployeeClaims`):

| Klaim | Tipe | Keterangan |
|-------|------|------------|
| `employee_id` | number | ID karyawan |
| `branch_id` | number | ID cabang |
| `role` | string | Peran, mis. `cashier`, `supervisor`, `admin` |

Middleware juga mengisi `RegisteredClaims` JWT standar (mis. `exp`, `iat`).

### Error dari middleware JWT (endpoint dilindungi)

Tanpa header atau format salah:

**401 Unauthorized**

```json
{ "success": false, "message": "Authorization header required" }
```

```json
{ "success": false, "message": "Invalid authorization format" }
```

Token tidak valid / kedaluwarsa:

```json
{ "success": false, "message": "Invalid or expired token" }
```

### Middleware peran (`RequireRole`)

Jika klaim tidak ada di context:

```json
{ "success": false, "message": "Unauthorized" }
```

Jika peran tidak termasuk daftar yang diizinkan:

```json
{ "success": false, "message": "Insufficient permissions" }
```

---

## Endpoint

### 1. Login

**POST** `/api/v1/auth/login`  
**Auth:** tidak perlu

#### Request body

| Field | Tipe | Wajib | Keterangan |
|-------|------|-------|------------|
| `nik` | string | ya | NIK karyawan |
| `pin` | string | ya | PIN |

Contoh:

```json
{
  "nik": "EMP001",
  "pin": "1234"
}
```

#### Respons sukses — 200 OK

```json
{
  "success": true,
  "message": "Login berhasil",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "employee": {
      "id": 1,
      "branch_id": 1,
      "nik": "EMP001",
      "full_name": "Kasir Satu",
      "role": "cashier",
      "is_active": true,
      "created_at": "2024-01-15T10:00:00Z"
    }
  }
}
```

Field `pin_hash` tidak dikirim ke klien (`json:"-"`).

#### Error

| HTTP | Kondisi | Body JSON |
|------|---------|-----------|
| 400 | Body bukan JSON valid | `{ "success": false, "message": "Invalid request body" }` |
| 400 | `nik` atau `pin` kosong | `{ "success": false, "message": "NIK dan PIN wajib diisi" }` |
| 401 | NIK tidak ditemukan / tidak aktif, atau PIN salah | `{ "success": false, "message": "NIK atau PIN salah" }` |
| 401 | Gagal generate token | `{ "success": false, "message": "gagal generate token" }` |

---

### 2. Cari produk menurut barcode

**GET** `/api/v1/products/barcode/{barcode}`  
**Auth:** Bearer JWT

#### Path parameter

| Parameter | Keterangan |
|-----------|------------|
| `barcode` | Kode barcode produk |

#### Respons sukses — 200 OK

Stok > 0:

```json
{
  "success": true,
  "message": "Produk ditemukan",
  "data": {
    "id": 1,
    "category_id": 1,
    "supplier_id": 2,
    "barcode": "8999999001",
    "name": "Indomie Goreng",
    "unit": "pcs",
    "buy_price": 2500,
    "sell_price": 3500,
    "min_stock": 10,
    "is_active": true,
    "created_at": "2024-01-15T10:00:00Z",
    "stock": 100
  }
}
```

Stok = 0 (produk tetap dikembalikan di `data`, `success` **false**):

```json
{
  "success": false,
  "message": "Stok produk habis",
  "data": {
    "id": 1,
    "category_id": 1,
    "supplier_id": 2,
    "barcode": "8999999001",
    "name": "Indomie Goreng",
    "unit": "pcs",
    "buy_price": 2500,
    "sell_price": 3500,
    "min_stock": 10,
    "is_active": true,
    "created_at": "2024-01-15T10:00:00Z",
    "stock": 0
  }
}
```

#### Error

| HTTP | Kondisi | Body |
|------|---------|------|
| 400 | Parameter `barcode` kosong | `{ "success": false, "message": "Barcode diperlukan" }` |
| 401 | Tanpa token / klaim tidak ada (jalur defensif) | `{ "success": false, "message": "Unauthorized" }` |
| 401 | Lihat tabel error JWT di atas | (sama) |
| 404 | Produk tidak ada atau tidak aktif | `{ "success": false, "message": "Produk tidak ditemukan" }` |

---

### 3. Produk stok rendah (alert)

**GET** `/api/v1/products/low-stock`  
**Auth:** Bearer JWT

**Catatan implementasi:** Handler saat ini selalu mengembalikan sukses dengan `data: null` (belum ada query low stock).

#### Respons — 200 OK

```json
{
  "success": true,
  "message": "Low stock products",
  "data": null
}
```

#### Error

| HTTP | Kondisi | Body |
|------|---------|------|
| 401 | Tanpa token / tidak sah | Sama seperti error JWT |

---

### 4. Detail low stock (supervisor / admin)

**GET** `/api/v1/products/low-stock/detail`  
**Auth:** Bearer JWT + peran `supervisor` atau `admin`

Perilaku handler sama dengan endpoint #3 (stub yang sama).

#### Error tambahan

| HTTP | Kondisi | Body |
|------|---------|------|
| 401 | Bukan supervisor/admin | `{ "success": false, "message": "Insufficient permissions" }` |

---

### 5. Buat transaksi

**POST** `/api/v1/transactions`  
**Auth:** Bearer JWT

#### Request body

| Field | Tipe | Wajib | Keterangan |
|-------|------|-------|------------|
| `branch_id` | number | tidak | Jika `0` atau tidak diisi, diganti dari JWT `branch_id` |
| `register_id` | number | tidak | ID mesin kasir |
| `member_code` | string | tidak | Kode member; jika kosong, transaksi tanpa member |
| `items` | array | ya | Minimal satu item |
| `promo_code` | string | tidak | Ada di model; belum diproses di service |
| `pay_method` | string | ya | `cash` \| `qris` \| `va` \| `ewallet` |

**Item:**

| Field | Tipe | Wajib |
|-------|------|-------|
| `product_id` | number | ya |
| `quantity` | number | ya (harus > 0 secara logika bisnis; tidak divalidasi eksplisit di handler) |

Contoh minimum:

```json
{
  "branch_id": 1,
  "register_id": 1,
  "member_code": "MBR001",
  "pay_method": "cash",
  "items": [
    { "product_id": 1, "quantity": 2 }
  ]
}
```

#### Perhitungan (service)

- Subtotal = jumlah (`sell_price * quantity`) per baris.
- PPN **11%** dari subtotal (`tax`), dibulatkan.
- `grand_total` = subtotal + tax.

#### Respons sukses — 201 Created

Field `data` adalah objek `CreateTransactionResponse`:

```json
{
  "success": true,
  "message": "<sama dengan message di dalam data.message>",
  "data": {
    "transaction": {
      "id": 1,
      "branch_id": 0,
      "register_id": 0,
      "employee_id": 0,
      "member_id": null,
      "promo_id": null,
      "invoice_no": "INV-20240115143022-1",
      "subtotal": 7000,
      "discount": 0,
      "tax": 770,
      "grand_total": 7770,
      "status": "paid",
      "created_at": "0001-01-01T00:00:00Z"
    },
    "payment_url": "",
    "qr_string": "",
    "va_number": "",
    "midtrans_transaction_id": "",
    "message": "Pembayaran cash berhasil"
  }
}
```

**Catatan:** Untuk alur **cash**, service mengisi subset field `transaction` dan mengubah status menjadi `paid`. Beberapa field numerik/waktu bisa nol/default karena tidak di-load ulang dari database setelah settle.

Variasi menurut `pay_method`:

| `pay_method` | `status` transaksi (awal/sesuai response) | Field tambahan di `data` |
|--------------|------------------------------------------|---------------------------|
| `cash` | `paid` | — |
| `qris` | `pending` | `qr_string`, `midtrans_transaction_id` |
| `va` | `pending` | `va_number`, `midtrans_transaction_id` |
| `ewallet` | `pending` | `payment_url`, `midtrans_transaction_id` |

Contoh `message` sukses (nilai pasti tergantung metode):

- Cash: `"Pembayaran cash berhasil"`
- QRIS: `"Scan QR untuk menyelesaikan pembayaran"`
- VA: `"Transfer ke VA untuk menyelesaikan pembayaran"`
- E-wallet: `"Buka link untuk menyelesaikan pembayaran"`

#### Error — 400 Bad Request (handler memetakan semua error service ke 400)

| `message` (contoh / pola) | Penyebab |
|---------------------------|----------|
| `Invalid request body` | Body bukan JSON |
| `Items tidak boleh kosong` | `items` kosong |
| `Metode pembayaran wajib diisi (cash/qris/va/ewallet)` | `pay_method` kosong |
| `Unauthorized` | Klaim JWT tidak ada (defensif) |
| `product {id}: product not found: ...` | `product_id` tidak ditemukan / tidak aktif |
| `stok {nama} tidak cukup (tersedia: {n})` | Stok cabang tidak mencukupi |
| `gagal menyimpan transaksi: ...` | Gagal simpan transaksi/item DB |
| `gagal settle cash: ...` | Gagal selesaikan pembayaran tunai |
| `gagal membuat QRIS: ...` | Error API Midtrans QRIS |
| `gagal membuat Virtual Account: ...` | Error API Midtrans VA |
| `gagal membuat invoice Midtrans: ...` | Error API Midtrans Invoice |
| `metode pembayaran tidak valid: {pay_method}` | Nilai selain empat metode di atas |

#### Error — 401

Sama seperti error JWT untuk endpoint dilindungi.

---

### 6. Midtrans Webhook

**POST** `/api/v1/midtrans/webhook`

**Auth:** tidak pakai JWT. Signature diverifikasi server-side menggunakan `signature_key` di dalam request body.

Ini **bukan** respons standar `APIResponse` untuk semua kasus.

#### Request body (payload)

Subset field yang dipakai aplikasi:

| Field | Tipe | Keterangan |
|-------|------|------------|
| `order_id` | string | Harus sama dengan `invoice_no` transaksi |
| `transaction_id` | string | Midtrans transaction id (dipakai untuk idempotency) |
| `transaction_status` | string | Contoh: `settlement` (paid), `pending`, `expire`, `deny` |
| `payment_type` | string | Contoh: `qris`, `bank_transfer`, `gopay` |
| `status_code` | string | Dipakai untuk perhitungan signature |
| `gross_amount` | string | Dipakai untuk perhitungan signature |
| `signature_key` | string | Hasil signature dari Midtrans |

#### Respons

Handler akan selalu membalas **200 OK** dengan salah satu status berikut:

Sukses memproses:

```json
{ "status": "received" }
```

Signature valid tapi proses internal gagal:

```json
{ "status": "error_logged" }
```

Payload rusak (bukan JSON valid):

```json
{ "status": "invalid_payload" }
```

Signature tidak valid:

```json
{ "status": "invalid_signature" }
```

---

## Ringkasan kode HTTP

| Kode | Penggunaan |
|------|------------|
| 200 | Sukses umum, GET produk, webhook OK |
| 201 | Transaksi berhasil dibuat (endpoint create transaction) |
| 400 | Validasi / bad request / error bisnis transaksi (dipetakan ke 400) |
| 401 | Login gagal, JWT gagal, webhook token salah |
| 404 | Produk tidak ditemukan |

**Tidak digunakan oleh handler saat ini untuk endpoint terdokumentasi:** `500` via `response.InternalError` (tersedia di `pkg/response` tetapi belum dipanggil dari handler yang ada).

---

## Referensi cepat path

| Method | Path | Auth |
|--------|------|------|
| POST | `/api/v1/auth/login` | Tidak |
| GET | `/api/v1/products/barcode/{barcode}` | JWT |
| GET | `/api/v1/products/low-stock` | JWT |
| GET | `/api/v1/products/low-stock/detail` | JWT + supervisor/admin |
| POST | `/api/v1/transactions` | JWT |
| POST | `/api/v1/midtrans/webhook` | Verifikasi via `signature_key` di body |
