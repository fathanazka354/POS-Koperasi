-- =========================================================
-- Chat marketplace (buyer ↔ seller per produk) + riwayat pesan
-- =========================================================

-- Satu thread per kombinasi: pembeli (member) + penjual (karyawan) + produk
-- (mirip “chat ke penjual” di marketplace untuk satu listing)
CREATE TABLE IF NOT EXISTS chat_conversations (
  id                  BIGSERIAL PRIMARY KEY,
  product_id          INT NOT NULL REFERENCES products(id),
  member_id           INT NOT NULL REFERENCES members(id),
  seller_employee_id  INT NOT NULL REFERENCES employees(id),
  branch_id           INT REFERENCES branches(id),
  last_message_at     TIMESTAMP,
  created_at          TIMESTAMP DEFAULT NOW(),
  UNIQUE (product_id, member_id, seller_employee_id)
);

CREATE INDEX IF NOT EXISTS idx_chat_conv_member ON chat_conversations(member_id, last_message_at DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_chat_conv_seller ON chat_conversations(seller_employee_id, last_message_at DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_chat_conv_product ON chat_conversations(product_id);

CREATE TABLE IF NOT EXISTS chat_messages (
  id               BIGSERIAL PRIMARY KEY,
  conversation_id  BIGINT NOT NULL REFERENCES chat_conversations(id) ON DELETE CASCADE,
  sender_role      VARCHAR(20) NOT NULL,
  body             TEXT NOT NULL,
  created_at       TIMESTAMP DEFAULT NOW(),
  CONSTRAINT chk_chat_sender_role CHECK (sender_role IN ('buyer', 'seller'))
);

CREATE INDEX IF NOT EXISTS idx_chat_msg_conv ON chat_messages(conversation_id, created_at ASC);
