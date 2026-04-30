-- =========================================================
-- Shop: member addresses, vouchers, dan nullable employee_id
-- =========================================================

-- Izinkan transaksi dari member checkout (tanpa employee / system)
ALTER TABLE transactions ALTER COLUMN employee_id DROP NOT NULL;

-- Alamat pengiriman member
CREATE TABLE IF NOT EXISTS member_addresses (
  id           SERIAL PRIMARY KEY,
  member_id    INT NOT NULL REFERENCES members(id),
  label        VARCHAR(100) NOT NULL DEFAULT 'Rumah',
  recipient    VARCHAR(150) NOT NULL,
  phone        VARCHAR(20) NOT NULL,
  address_line TEXT NOT NULL,
  city         VARCHAR(100) NOT NULL,
  province     VARCHAR(100) NOT NULL DEFAULT '',
  postal_code  VARCHAR(10) NOT NULL DEFAULT '',
  is_default   BOOLEAN DEFAULT false,
  created_at   TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_member_addresses_member ON member_addresses(member_id);

-- Voucher / promo code
CREATE TABLE IF NOT EXISTS vouchers (
  id            SERIAL PRIMARY KEY,
  code          VARCHAR(50) UNIQUE NOT NULL,
  name          VARCHAR(150) NOT NULL,
  discount_type VARCHAR(20) NOT NULL DEFAULT 'percent',
  value         NUMERIC(10,2) NOT NULL,
  min_purchase  NUMERIC(15,2) DEFAULT 0,
  max_discount  NUMERIC(15,2),
  quota         INT DEFAULT 100,
  used_count    INT DEFAULT 0,
  start_date    DATE NOT NULL,
  end_date      DATE NOT NULL,
  is_active     BOOLEAN DEFAULT true
);

-- Riwayat pemakaian voucher
CREATE TABLE IF NOT EXISTS voucher_usages (
  id             SERIAL PRIMARY KEY,
  voucher_id     INT REFERENCES vouchers(id),
  member_id      INT REFERENCES members(id),
  transaction_id BIGINT REFERENCES transactions(id),
  created_at     TIMESTAMP DEFAULT NOW()
);

-- Seed voucher demo
INSERT INTO vouchers (code, name, discount_type, value, min_purchase, max_discount, quota, start_date, end_date)
VALUES
  ('HEMAT10', 'Diskon 10%', 'percent', 10, 10000, 50000, 100,
   CURRENT_DATE, CURRENT_DATE + INTERVAL '30 days'),
  ('HEMAT5K', 'Diskon Rp 5.000', 'fixed', 5000, 20000, NULL, 50,
   CURRENT_DATE, CURRENT_DATE + INTERVAL '30 days')
ON CONFLICT (code) DO NOTHING;

-- Tambah kolom address di transactions (opsional, untuk referensi)
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS address_id INT REFERENCES member_addresses(id);
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS voucher_id INT REFERENCES vouchers(id);
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS voucher_discount NUMERIC(15,2) DEFAULT 0;
