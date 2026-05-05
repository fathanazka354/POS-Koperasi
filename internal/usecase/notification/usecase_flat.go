package notification

import (
	"github.com/fathanazka354/pos-koperasi/internal/entity/notification"
	notifyws "github.com/fathanazka354/pos-koperasi/internal/gateway/notifyws"
)

type usecase struct {
	repo Repository
	hub  *notifyws.NotifyHub
}

func New(repo Repository, hub *notifyws.NotifyHub) Usecase {
	return &usecase{repo: repo, hub: hub}
}

func (u *usecase) Notify(memberID int, notifType, title, body string, data map[string]interface{}) error {
	n := notification.Notification{
		MemberID: memberID,
		Type:     notifType,
		Title:    title,
		Body:     body,
		Data:     data,
		IsRead:   false,
	}
	created, err := u.repo.Create(n)
	if err != nil {
		return err
	}
	// Push ke WebSocket jika member sedang terkoneksi
	u.hub.Push(memberID, map[string]interface{}{
		"event":        "notification",
		"notification": created,
	})
	return nil
}

func (u *usecase) ListByMember(memberID int) ([]notification.Notification, error) {
	return u.repo.ListByMember(memberID, 50)
}

func (u *usecase) MarkRead(id string, memberID int) error {
	return u.repo.MarkRead(id, memberID)
}

func (u *usecase) MarkAllRead(memberID int) error {
	return u.repo.MarkAllRead(memberID)
}

func (u *usecase) CountUnread(memberID int) (int, error) {
	return u.repo.CountUnread(memberID)
}

