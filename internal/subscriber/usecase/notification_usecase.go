package usecase

import (
	"asr_leasing_notification/internal/publisher/handler"
	"asr_leasing_notification/internal/subscriber/notifiers"
	"asr_leasing_notification/internal/subscriber/repository"
	"asr_leasing_notification/pkg/models"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/viper"
)

const maxRetries = 3

type NotificationService struct {
	repo      *repository.Queries
	notifiers map[repository.NotificationChannel]notifiers.Notifier
	publisher *handler.AMQPPublisher
}

func NewNotificationService(
	repo *repository.Queries,
	notifiers map[repository.NotificationChannel]notifiers.Notifier,
	publisher *handler.AMQPPublisher,
) *NotificationService {
	return &NotificationService{
		repo:      repo,
		notifiers: notifiers,
		publisher: publisher,
	}
}

func (s *NotificationService) ProcessNotification(body []byte) {
	var req models.NotificationRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("ERROR: Failed to unmarshal message body: %s", err)
		return
	}

	params := repository.CreateNotificationParams{
		UserID: sql.NullInt64{
			Int64: req.Metadata.UserID,
			Valid: req.Metadata.UserID != 0,
		},
		Channel:   repository.NotificationChannel(req.Type),
		Recipient: req.Recipient,
		Subject: sql.NullString{
			String: req.Subject,
			Valid:  req.Subject != "",
		},
		Body: sql.NullString{
			String: req.Body,
			Valid:  req.Body != "",
		},
		Source: req.Metadata.Source,
	}

	dbNotification, err := s.repo.CreateNotification(context.Background(), params)
	if err != nil {
		log.Printf("ERROR: Failed to create notification record in DB: %v", err)
		return
	}
	log.Printf("Created PENDING notification record ID %d from source '%s'", dbNotification.ID, dbNotification.Source)

	notifier, ok := s.notifiers[dbNotification.Channel]
	if !ok {
		s.handleSendFailure(dbNotification, fmt.Errorf("no notifier for channel '%s'", dbNotification.Channel))
		return
	}

	err = notifier.Send(context.Background(), dbNotification)
	if err != nil {
		s.handleSendFailure(dbNotification, err)
		return
	}

	s.repo.UpdateNotificationSuccess(context.Background(), dbNotification.ID)
	log.Printf("Successfully SENT notification ID %d", dbNotification.ID)
}

func (s *NotificationService) handleSendFailure(notification repository.Notification, sendErr error) {
	log.Printf("ERROR: Failed to send notification ID %d (attempt %d): %v", notification.ID, notification.RetryCount.Int32+1, sendErr)
	s.repo.UpdateNotificationFailure(context.Background(), repository.UpdateNotificationFailureParams{
		ID: notification.ID,
		ErrorMessage: sql.NullString{
			String: sendErr.Error(),
			Valid:  true,
		},
	})

	if notification.RetryCount.Int32+1 >= maxRetries {
		log.Printf("PERMANENTLY FAILED: Notification ID %d has reached max retries.", notification.ID)
		return
	}

	retryQueue := viper.GetString("rabbitmq.retry_config.queue")
	delay := viper.GetString("rabbitmq.retry_config.delay_ms")

	originalMessage := models.NotificationRequest{
		Type:      string(notification.Channel),
		Recipient: notification.Recipient,
		Subject:   notification.Subject.String,
		Body:      notification.Body.String,
		Metadata: models.Metadata{
			UserID: notification.UserID.Int64,
			Source: notification.Source,
		},
	}
	body, _ := json.Marshal(originalMessage)

	err := s.publisher.PublishWithTTL(body, retryQueue, delay)
	if err != nil {
		log.Printf("CRITICAL: Failed to publish to retry queue for ID %d: %v", notification.ID, err)
		return
	}
	log.Printf("Queued notification ID %d for retry in %s ms.", notification.ID, delay)
}
