package notifiers

import (
	"asr_leasing_notification/internal/subscriber/repository"
	"context"
)

type Notifier interface {
	Send(ctx context.Context, notification repository.Notification) error
}
