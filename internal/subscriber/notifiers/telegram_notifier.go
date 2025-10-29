package notifiers

import (
	"asr_leasing_notification/internal/subscriber/repository"
	"context"
	"log"
)

type TelegramNotifier struct {
	// Fields for FCM or APNs clients.
}

func NewTelegramNotifier() *TelegramNotifier {
	return &TelegramNotifier{}
}

func (n *TelegramNotifier) Send(ctx context.Context, notification repository.Notification) error {

	log.Printf("-----> Sending Telegram to [%s] with subject [%s]", notification.Recipient, notification.Subject.String)

	return nil
}
