package chat

import (
	"database/sql"
	"fmt"

	"gorm.io/gorm"

	"github.com/fathanazka354/pos-koperasi/internal/model"
	authdomain "github.com/fathanazka354/pos-koperasi/internal/entity/auth"
	chatcontract "github.com/fathanazka354/pos-koperasi/internal/usecase/chat"
	productdomain "github.com/fathanazka354/pos-koperasi/internal/entity/product"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository { return &Repository{db: db} }

var _ chatcontract.Repository = (*Repository)(nil)

func (r *Repository) CreateConversation(c model.ChatConversation) (int64, error) {
	var id int64
	var pid interface{}
	if c.ProductID != nil {
		pid = *c.ProductID
	}
	var bid interface{}
	if c.BranchID != nil {
		bid = *c.BranchID
	}
	err := r.db.Raw(`
		INSERT INTO chat_conversations (product_id, member_id, seller_employee_id, branch_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (member_id, seller_employee_id)
		DO UPDATE SET
			product_id = COALESCE(EXCLUDED.product_id, chat_conversations.product_id),
			branch_id = COALESCE(EXCLUDED.branch_id, chat_conversations.branch_id)
		RETURNING id`,
		pid, c.MemberID, c.SellerEmployeeID, bid,
	).Scan(&id).Error
	return id, err
}

func (r *Repository) GetConversation(id int64) (*model.ChatConversation, error) {
	var row model.ChatConversation
	err := r.db.Raw(`
		SELECT c.id, c.product_id, c.member_id, c.seller_employee_id, c.branch_id, c.last_message_at, c.created_at,
		       COALESCE(p.name, '') AS product_name, m.full_name AS member_name, e.full_name AS seller_name
		FROM chat_conversations c
		LEFT JOIN products p ON p.id = c.product_id
		JOIN members m ON m.id = c.member_id
		JOIN employees e ON e.id = c.seller_employee_id
		WHERE c.id = $1`, id).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, sql.ErrNoRows
	}
	return &row, nil
}

// Hanya pakai m2.body agar query jalan sebelum/ tanpa migrasi 004; stub produk = body kosong.
const lastMessageSubquery = `(SELECT CASE
			WHEN TRIM(COALESCE(m2.body, '')) = '' THEN '📎 Produk'
			ELSE LEFT(TRIM(COALESCE(m2.body, '')), 500)
		END
		FROM chat_messages m2
		WHERE m2.conversation_id = c.id
		ORDER BY m2.created_at DESC LIMIT 1) AS last_message`

func (r *Repository) ListForMember(memberID int) ([]model.ChatConversation, error) {
	var rows []model.ChatConversation
	err := r.db.Raw(`
		SELECT c.id, c.product_id, c.member_id, c.seller_employee_id, c.branch_id, c.last_message_at, c.created_at,
		       COALESCE(p.name, '') AS product_name, m.full_name AS member_name, e.full_name AS seller_name,
		       `+lastMessageSubquery+`
		FROM chat_conversations c
		LEFT JOIN products p ON p.id = c.product_id
		JOIN members m ON m.id = c.member_id
		JOIN employees e ON e.id = c.seller_employee_id
		WHERE c.member_id = $1
		ORDER BY c.last_message_at DESC NULLS LAST, c.created_at DESC`, memberID).Scan(&rows).Error
	return rows, err
}

func (r *Repository) ListForEmployee(employeeID int) ([]model.ChatConversation, error) {
	var rows []model.ChatConversation
	err := r.db.Raw(`
		SELECT c.id, c.product_id, c.member_id, c.seller_employee_id, c.branch_id, c.last_message_at, c.created_at,
		       COALESCE(p.name, '') AS product_name, m.full_name AS member_name, e.full_name AS seller_name,
		       `+lastMessageSubquery+`
		FROM chat_conversations c
		LEFT JOIN products p ON p.id = c.product_id
		JOIN members m ON m.id = c.member_id
		JOIN employees e ON e.id = c.seller_employee_id
		WHERE c.seller_employee_id = $1
		ORDER BY c.last_message_at DESC NULLS LAST, c.created_at DESC`, employeeID).Scan(&rows).Error
	return rows, err
}

func (r *Repository) ShouldInsertProductContext(convID int64, productID int) (bool, error) {
	var last sql.NullInt64
	row := r.db.Raw(`
		SELECT product_id FROM chat_messages
		WHERE conversation_id = $1
		ORDER BY created_at DESC
		LIMIT 1`, convID).Row()
	err := row.Scan(&last)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if last.Valid && int(last.Int64) == productID {
		return false, nil
	}
	return true, nil
}

func (r *Repository) InsertMessage(convID int64, senderRole, body string, productID *int) (*model.ChatMessage, error) {
	var msg model.ChatMessage
	if err := r.db.Raw(`
		INSERT INTO chat_messages (conversation_id, sender_role, body, product_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, conversation_id, sender_role, body, product_id, created_at`,
		convID, senderRole, body, productID,
	).Scan(&msg).Error; err != nil {
		return nil, err
	}
	if err := r.db.Exec(`UPDATE chat_conversations SET last_message_at = NOW() WHERE id = $1`, convID).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *Repository) ListMessages(convID int64, limit int) ([]model.ChatMessage, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var rows []model.ChatMessage
	err := r.db.Raw(`
		SELECT id, conversation_id, sender_role, body, product_id, created_at
		FROM chat_messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
		LIMIT $2`, convID, limit).Scan(&rows).Error
	return rows, err
}

func (r *Repository) ParticipantRole(convID int64, memberID *int, employeeID *int) (string, error) {
	var mid, eid int
	row := r.db.Raw(`SELECT member_id, seller_employee_id FROM chat_conversations WHERE id = $1`, convID).Row()
	if err := row.Scan(&mid, &eid); err != nil {
		return "", err
	}
	if memberID != nil && *memberID == mid {
		return "buyer", nil
	}
	if employeeID != nil && *employeeID == eid {
		return "seller", nil
	}
	return "", fmt.Errorf("bukan peserta percakapan")
}

func (r *Repository) GetActiveEmployeeByID(id int) (*authdomain.Employee, error) {
	var emp authdomain.Employee
	if err := r.db.Raw(`SELECT * FROM employees WHERE id = $1 AND is_active = true`, id).Scan(&emp).Error; err != nil {
		return nil, err
	}
	if emp.ID == 0 {
		return nil, sql.ErrNoRows
	}
	return &emp, nil
}

func (r *Repository) GetActiveProductByID(id int) (*productdomain.Product, error) {
	var p productdomain.Product
	if err := r.db.Raw(`SELECT * FROM products WHERE id = $1 AND is_active = true`, id).Scan(&p).Error; err != nil {
		return nil, err
	}
	if p.ID == 0 {
		return nil, sql.ErrNoRows
	}
	return &p, nil
}

