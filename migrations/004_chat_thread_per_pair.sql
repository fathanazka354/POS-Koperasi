-- =========================================================
-- Satu thread chat per pasangan member + penjual (banyak produk dalam satu aliran)
-- + lampiran produk sebagai baris pesan (product_id pada chat_messages)
-- =========================================================

BEGIN;

-- Hapus constraint unik per produk
ALTER TABLE chat_conversations
  DROP CONSTRAINT IF EXISTS chat_conversations_product_id_member_id_seller_employee_id_key;

-- Gabungkan percakapan duplikat ke id terkecil per (member, penjual); pindahkan pesan
UPDATE chat_messages m
SET conversation_id = x.keep_id
FROM (
  SELECT c.id AS old_id,
         MIN(c.id) OVER (PARTITION BY c.member_id, c.seller_employee_id) AS keep_id
  FROM chat_conversations c
) x
WHERE m.conversation_id = x.old_id AND x.old_id <> x.keep_id;

DELETE FROM chat_conversations c
WHERE c.id IN (
  SELECT id FROM (
    SELECT id,
           ROW_NUMBER() OVER (
             PARTITION BY member_id, seller_employee_id ORDER BY id
           ) AS rn
    FROM chat_conversations
  ) t WHERE rn > 1
);

-- Kolom produk pada thread boleh kosong setelah merge / konteks lama
ALTER TABLE chat_conversations ALTER COLUMN product_id DROP NOT NULL;

-- Unik per pasangan (idempotent: bisa sudah ada dari percobaan migrasi sebelumnya)
ALTER TABLE chat_conversations DROP CONSTRAINT IF EXISTS chat_conversations_member_id_seller_employee_id_key;
DROP INDEX IF EXISTS chat_conversations_member_id_seller_employee_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS chat_conversations_member_seller_uidx
  ON chat_conversations (member_id, seller_employee_id);

ALTER TABLE chat_messages
  ADD COLUMN IF NOT EXISTS product_id INT REFERENCES products(id) ON DELETE SET NULL;

COMMIT;
