# Architecture Contracts (Regression Checklist)

Dokumen ini berisi kontrak publik/internal yang **tidak boleh berubah** saat refactor arsitektur (clean-layer, by-layer, migrasi GORM).

## HTTP routes (public contract)

Sumber kebenaran saat ini: `cmd/api/routing.go`.

- `GET /healthz`
- `GET /` → redirect ke `/demo/`
- `GET /demo` → redirect ke `/demo/`
- `GET /demo/chat` → redirect ke `/demo/` (backward compat)
- `GET /demo/shop` → redirect ke `/demo/` (backward compat)
- static: `GET /demo/*`

Semua API berada di prefix:

- `/api/v1/*`

Catatan grouping (harus tetap):

- Public:
  - `auth` routes
  - transaction public routes (webhook Midtrans)
  - chat routes (membutuhkan JWT secret untuk WS token)
  - product public shop routes
  - notification WS + member routes
  - address member routes
  - shop checkout + orders
- Protected group (`middleware.JWTAuth(cfg.JWTSecret)`):
  - transaction protected routes
  - product protected routes
  - seller routes
  - supervisor/admin-only: `GET /products/low-stock/detail`

## Redis Streams (internal wire contract)

Sumber kebenaran saat ini: `internal/gateway/queue/*`.

### Midtrans webhook stream

- **Stream**: `midtrans:webhooks` (`queue.StreamMidtransWebhooks`)
- **Consumer group**: `midtrans-webhook-workers` (`queue.GroupMidtransWorkers`)
- **Field**: `payload` (string JSON)
- **Ack policy**: selalu `XACK` setelah handler dipanggil (error hanya di-log).

### Member notification stream

- **Stream**: `notifications:member` (`queue.StreamMemberNotifications`)
- **Consumer group**: `member-notify-workers` (`queue.GroupMemberNotifyWorkers`)
- **Field**: `payload` (string JSON)
- **Ack policy**: selalu `XACK` setelah `Notify` dipanggil (error hanya di-log).

## Fallback behaviour (availability contracts)

Sumber kebenaran saat ini: `cmd/api/inject.go`.

- **MongoDB gagal**:
  - notification repository harus fallback ke in-memory
  - payment outbox harus fallback ke direct dispatch (sinkron / via Redis jika ada)
- **Redis tidak diset atau ping gagal**:
  - webhook Midtrans diproses sinkron (queue `nil`)
  - member notifications tetap bisa jalan via direct dispatch (tanpa Redis)

## Outbox semantics (internal behaviour contract)

Sumber kebenaran saat ini: `internal/gateway/outbox/mongo_payment_store.go`.

- koleksi Mongo: `outbox_events`
- claim mechanism:
  - event dianggap available jika `published_at` nil/tidak ada
  - event bisa di-claim jika `claimed_at` nil/tidak ada atau stale (>15 menit)
- `MarkPublished` harus mengisi `published_at` dan reset `claimed_at`

