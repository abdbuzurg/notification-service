package notifiers

import (
	"asr_leasing_notification/internal/subscriber/repository"
	"context"
	"log"
)

type EmailNotifier struct {
	// SMTP Config
}

func NewEmailNotifier() *EmailNotifier {
	return &EmailNotifier{}
}

func (n *EmailNotifier) Send(ctx context.Context, notification repository.Notification) error {

	log.Printf("-----> Sending EMAIL to [%s] with subject [%s]", notification.Recipient, notification.Subject.String)

	return nil
}
