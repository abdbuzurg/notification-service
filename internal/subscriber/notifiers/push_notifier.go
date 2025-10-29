package notifiers

import (
	"asr_leasing_notification/internal/subscriber/repository"
	"context"
	"log"
)

type PushNotifier struct {
	// Fields for FCM or APNs clients.
}

func NewPushNotifier() *PushNotifier {
	return &PushNotifier{}
}

func (n *PushNotifier) Send(ctx context.Context, notification repository.Notification) error {

	log.Printf("-----> Sending PUSH to [%s] with subject [%s]", notification.Recipient, notification.Subject.String)

	return nil
}
