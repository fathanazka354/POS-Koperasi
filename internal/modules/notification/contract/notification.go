package contract

import "github.com/yourname/pos-koperasi/internal/modules/notification/domain"

type NotificationRepository interface {
	Create(n domain.Notification) (*domain.Notification, error)
	ListByMember(memberID int, limit int) ([]domain.Notification, error)
	MarkRead(id string, memberID int) error
	MarkAllRead(memberID int) error
	CountUnread(memberID int) (int, error)
}

type NotificationService interface {
	// Dipanggil setelah payment settle
	Notify(memberID int, notifType, title, body string, data map[string]interface{}) error
	ListByMember(memberID int) ([]domain.Notification, error)
	MarkRead(id string, memberID int) error
	MarkAllRead(memberID int) error
	CountUnread(memberID int) (int, error)
}
