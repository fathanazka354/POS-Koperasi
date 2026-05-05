// Package notification menyediakan in-memory notification repository sebagai fallback
// ketika MongoDB tidak tersedia (untuk lingkungan development/demo).
package notification

import (
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	notifcontract "github.com/fathanazka354/pos-koperasi/internal/usecase/notification"
	"github.com/fathanazka354/pos-koperasi/internal/entity/notification"
)

type MemoryRepository struct {
	mu    sync.RWMutex
	items []notification.Notification
}

func NewMemory() *MemoryRepository { return &MemoryRepository{} }

var _ notifcontract.Repository = (*MemoryRepository)(nil)

func (r *MemoryRepository) Create(n notification.Notification) (*notification.Notification, error) {
	n.ID = primitive.NewObjectID()
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now()
	}
	r.mu.Lock()
	r.items = append(r.items, n)
	r.mu.Unlock()
	return &n, nil
}

func (r *MemoryRepository) ListByMember(memberID int, limit int) ([]notification.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []notification.Notification
	for i := len(r.items) - 1; i >= 0 && len(out) < limit; i-- {
		if r.items[i].MemberID == memberID {
			out = append(out, r.items[i])
		}
	}
	return out, nil
}

func (r *MemoryRepository) MarkRead(id string, memberID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID.Hex() == id && r.items[i].MemberID == memberID {
			r.items[i].IsRead = true
			return nil
		}
	}
	return fmt.Errorf("notification not found")
}

func (r *MemoryRepository) MarkAllRead(memberID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].MemberID == memberID {
			r.items[i].IsRead = true
		}
	}
	return nil
}

func (r *MemoryRepository) CountUnread(memberID int) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, n := range r.items {
		if n.MemberID == memberID && !n.IsRead {
			count++
		}
	}
	return count, nil
}

