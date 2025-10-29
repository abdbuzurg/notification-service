package notifiers

import (
	"asr_leasing_notification/internal/subscriber/repository"
	"context"
	"log"
)

type SmsNotifer struct {
	// SMS client config
}

func NewSmsNotifier() *SmsNotifer {
	return &SmsNotifer{}
}

func (n *SmsNotifer) Send(ctx context.Context, notification repository.Notification) error {

	log.Printf("-----> Sending SMS to [%s] with subject [%s]", notification.Recipient, notification.Subject.String)

	return nil
}
