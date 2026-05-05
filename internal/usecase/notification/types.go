package notification

import "github.com/fathanazka354/pos-koperasi/internal/entity/notification"

type Repository interface {
	Create(n notification.Notification) (*notification.Notification, error)
	ListByMember(memberID int, limit int) ([]notification.Notification, error)
	MarkRead(id string, memberID int) error
	MarkAllRead(memberID int) error
	CountUnread(memberID int) (int, error)
}

type Usecase interface {
	// Dipanggil setelah payment settle
	Notify(memberID int, notifType, title, body string, data map[string]interface{}) error
	ListByMember(memberID int) ([]notification.Notification, error)
	MarkRead(id string, memberID int) error
	MarkAllRead(memberID int) error
	CountUnread(memberID int) (int, error)
}

