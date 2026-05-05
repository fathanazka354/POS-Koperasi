package queue

import (
	"context"
	"log"
)

// DispatchMemberNotify mengantri ke Redis jika tersedia; gagal enqueue atau tanpa Redis → fallback sinkron.
// sync adalah jalur sinkron (biasanya NotificationService.Notify); boleh nil jika hanya antrian.
func DispatchMemberNotify(ctx context.Context, q *RedisMemberNotifyQueue, sync func(memberID int, notifType, title, body string, data map[string]interface{}) error, memberID int, notifType, title, body string, data map[string]interface{}) {
	if data == nil {
		data = map[string]interface{}{}
	}
	if q != nil {
		job := MemberNotifyJob{
			MemberID: memberID,
			Type:     notifType,
			Title:    title,
			Body:     body,
			Data:     data,
		}
		if err := q.Enqueue(ctx, job); err != nil {
			log.Printf("queue: member notify enqueue gagal: %v — fallback sinkron", err)
			if sync != nil {
				_ = sync(memberID, notifType, title, body, data)
			}
			return
		}
		return
	}
	if sync != nil {
		_ = sync(memberID, notifType, title, body, data)
	}
}

