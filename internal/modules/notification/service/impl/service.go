package impl

import (
	"github.com/yourname/pos-koperasi/internal/modules/notification/contract"
	"github.com/yourname/pos-koperasi/internal/modules/notification/domain"
	notifyws "github.com/yourname/pos-koperasi/internal/modules/notification/ws"
)

type Service struct {
	repo contract.NotificationRepository
	hub  *notifyws.NotifyHub
}

func New(repo contract.NotificationRepository, hub *notifyws.NotifyHub) *Service {
	return &Service{repo: repo, hub: hub}
}

var _ contract.NotificationService = (*Service)(nil)

func (s *Service) Notify(memberID int, notifType, title, body string, data map[string]interface{}) error {
	n := domain.Notification{
		MemberID: memberID,
		Type:     notifType,
		Title:    title,
		Body:     body,
		Data:     data,
		IsRead:   false,
	}
	created, err := s.repo.Create(n)
	if err != nil {
		return err
	}
	// Push ke WebSocket jika member sedang terkoneksi
	s.hub.Push(memberID, map[string]interface{}{
		"event":        "notification",
		"notification": created,
	})
	return nil
}

func (s *Service) ListByMember(memberID int) ([]domain.Notification, error) {
	return s.repo.ListByMember(memberID, 50)
}

func (s *Service) MarkRead(id string, memberID int) error {
	return s.repo.MarkRead(id, memberID)
}

func (s *Service) MarkAllRead(memberID int) error {
	return s.repo.MarkAllRead(memberID)
}

func (s *Service) CountUnread(memberID int) (int, error) {
	return s.repo.CountUnread(memberID)
}
