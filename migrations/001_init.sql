-- =========================================================
-- POS Koperasi Minimarket — Database Migration
-- Run: psql -U postgres -d pos_koperasi -f 001_init.sql
-- =========================================================

CREATE TABLE IF NOT EXISTS branches (
  id         SERIAL PRIMARY KEY,
  name       VARCHAR(150) NOT NULL,
  address    TEXT,
  city       VARCHAR(100),
  phone      VARCHAR(20),
  is_active  BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS employees (
  id          SERIAL PRIMARY KEY,
  branch_id   INT REFERENCES branches(id),
  nik         VARCHAR(20) UNIQUE NOT NULL,
  full_name   VARCHAR(150) NOT NULL,
  role        VARCHAR(30) DEFAULT 'cashier',
  pin_hash    VARCHAR(255),
  is_active   BOOLEAN DEFAULT true,
  created_at  TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS cashier_registers (
  id            SERIAL PRIMARY KEY,
  branch_id     INT REFERENCES branches(id),
  register_code VARCHAR(20) UNIQUE NOT NULL,
  status        VARCHAR(20) DEFAULT 'active'
);

CREATE TABLE IF NOT EXISTS members (
  id            SERIAL PRIMARY KEY,
  member_code   VARCHAR(30) UNIQUE NOT NULL,
  full_name     VARCHAR(150) NOT NULL,
  phone         VARCHAR(20) UNIQUE,
  point_balance INT DEFAULT 0,
  joined_at     DATE DEFAULT CURRENT_DATE,
  is_active     BOOLEAN DEFAULT true
);

CREATE TABLE IF NOT EXISTS categories (
  id        SERIAL PRIMARY KEY,
  parent_id INT REFERENCES categories(id),
  name      VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS suppliers (
  id           SERIAL PRIMARY KEY,
  name         VARCHAR(150) NOT NULL,
  contact_name VARCHAR(100),
  phone        VARCHAR(20),
  address      TEXT,
  is_active    BOOLEAN DEFAULT true
);

CREATE TABLE IF NOT EXISTS products (
  id          SERIAL PRIMARY KEY,
  category_id INT REFERENCES categories(id),
  supplier_id INT REFERENCES suppliers(id),
  barcode     VARCHAR(50) UNIQUE NOT NULL,
  name        VARCHAR(200) NOT NULL,
  unit        VARCHAR(20) DEFAULT 'pcs',
  buy_price   NUMERIC(15,2) NOT NULL,
  sell_price  NUMERIC(15,2) NOT NULL,
  min_stock   INT DEFAULT 5,
  is_active   BOOLEAN DEFAULT true,
  created_at  TIMESTAMP DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_products_barcode ON products(barcode);

CREATE TABLE IF NOT EXISTS stocks (
  id         SERIAL PRIMARY KEY,
  product_id INT REFERENCES products(id),
  branch_id  INT REFERENCES branches(id),
  quantity   INT NOT NULL DEFAULT 0,
  updated_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(product_id, branch_id)
);
CREATE INDEX IF NOT EXISTS idx_stocks_branch ON stocks(branch_id, product_id);

CREATE TABLE IF NOT EXISTS promos (
  id         SERIAL PRIMARY KEY,
  name       VARCHAR(150) NOT NULL,
  type       VARCHAR(30) NOT NULL,
  value      NUMERIC(10,2) NOT NULL,
  min_qty    INT DEFAULT 1,
  start_date DATE NOT NULL,
  end_date   DATE NOT NULL,
  is_active  BOOLEAN DEFAULT true
);

CREATE TABLE IF NOT EXISTS transactions (
  id          BIGSERIAL PRIMARY KEY,
  branch_id   INT REFERENCES branches(id),
  register_id INT REFERENCES cashier_registers(id),
  employee_id INT REFERENCES employees(id),
  member_id   INT REFERENCES members(id),
  promo_id    INT REFERENCES promos(id),
  invoice_no  VARCHAR(50) UNIQUE NOT NULL,
  subtotal    NUMERIC(15,2) NOT NULL,
  discount    NUMERIC(15,2) DEFAULT 0,
  tax         NUMERIC(15,2) DEFAULT 0,
  grand_total NUMERIC(15,2) NOT NULL,
  status      VARCHAR(20) DEFAULT 'pending',
  created_at  TIMESTAMP DEFAULT NOW(),
  updated_at  TIMESTAMP DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_trx_branch_date ON transactions(branch_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_trx_member      ON transactions(member_id);
CREATE INDEX IF NOT EXISTS idx_trx_invoice     ON transactions(invoice_no);

CREATE TABLE IF NOT EXISTS transaction_items (
  id             SERIAL PRIMARY KEY,
  transaction_id BIGINT REFERENCES transactions(id),
  product_id     INT REFERENCES products(id),
  quantity       INT NOT NULL,
  unit_price     NUMERIC(15,2) NOT NULL,
  discount       NUMERIC(15,2) DEFAULT 0,
  subtotal       NUMERIC(15,2) NOT NULL
);

CREATE TABLE IF NOT EXISTS payments (
  id             SERIAL PRIMARY KEY,
  transaction_id BIGINT REFERENCES transactions(id),
  method         VARCHAR(30) NOT NULL,
  amount         NUMERIC(15,2) NOT NULL,
  change_amount  NUMERIC(15,2) DEFAULT 0,
  reference_no   VARCHAR(100) UNIQUE,
  paid_at        TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_payments_ref ON payments(reference_no);

CREATE TABLE IF NOT EXISTS point_histories (
  id             SERIAL PRIMARY KEY,
  member_id      INT REFERENCES members(id),
  transaction_id BIGINT REFERENCES transactions(id),
  type           VARCHAR(10) NOT NULL,
  points         INT NOT NULL,
  created_at     TIMESTAMP DEFAULT NOW()
);

-- ── Seed data ─────────────────────────────────────────────
INSERT INTO branches (name, city) VALUES ('Koperasi Pusat', 'Yogyakarta')
  ON CONFLICT DO NOTHING;

INSERT INTO cashier_registers (branch_id, register_code)
  VALUES (1, 'REG-001') ON CONFLICT DO NOTHING;

-- Password: 1234 (bcrypt hash)
INSERT INTO employees (branch_id, nik, full_name, role, pin_hash)
VALUES (1, 'EMP001', 'Kasir Satu', 'cashier',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lh...')
ON CONFLICT DO NOTHING;

INSERT INTO categories (name) VALUES ('Makanan'), ('Minuman'), ('Kebersihan')
  ON CONFLICT DO NOTHING;

INSERT INTO suppliers (name) VALUES ('PT Unilever Indonesia'), ('PT Indofood')
  ON CONFLICT DO NOTHING;

INSERT INTO products (category_id, supplier_id, barcode, name, unit, buy_price, sell_price, min_stock)
VALUES
  (1, 2, '8999999001', 'Indomie Goreng', 'pcs', 2500, 3500, 10),
  (2, 1, '8999999002', 'Aqua 600ml', 'pcs', 2000, 3000, 20),
  (3, 1, '8999999003', 'Sunlight 200ml', 'pcs', 5000, 7500, 5)
ON CONFLICT DO NOTHING;

INSERT INTO stocks (product_id, branch_id, quantity) VALUES
  (1, 1, 100), (2, 1, 200), (3, 1, 50)
ON CONFLICT (product_id, branch_id) DO NOTHING;

INSERT INTO members (member_code, full_name, phone)
VALUES ('MBR001', 'Budi Santoso', '081234567890')
ON CONFLICT DO NOTHING;
