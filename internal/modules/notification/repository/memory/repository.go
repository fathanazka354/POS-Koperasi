// Package memory menyediakan in-memory notification repository sebagai fallback
// ketika MongoDB tidak tersedia (untuk lingkungan development/demo).
package memory

import (
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/yourname/pos-koperasi/internal/modules/notification/contract"
	"github.com/yourname/pos-koperasi/internal/modules/notification/domain"
)

type Repository struct {
	mu    sync.RWMutex
	items []domain.Notification
}

func New() *Repository { return &Repository{} }

var _ contract.NotificationRepository = (*Repository)(nil)

func (r *Repository) Create(n domain.Notification) (*domain.Notification, error) {
	n.ID = primitive.NewObjectID()
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now()
	}
	r.mu.Lock()
	r.items = append(r.items, n)
	r.mu.Unlock()
	return &n, nil
}

func (r *Repository) ListByMember(memberID int, limit int) ([]domain.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Notification
	for i := len(r.items) - 1; i >= 0 && len(out) < limit; i-- {
		if r.items[i].MemberID == memberID {
			out = append(out, r.items[i])
		}
	}
	return out, nil
}

func (r *Repository) MarkRead(id string, memberID int) error {
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

func (r *Repository) MarkAllRead(memberID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].MemberID == memberID {
			r.items[i].IsRead = true
		}
	}
	return nil
}

func (r *Repository) CountUnread(memberID int) (int, error) {
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
