package notifiers

import (
	"asr_leasing_notification/internal/subscriber/repository"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
)

type OsonSMSNotifier struct {
	repo  *repository.Queries
	login string
	from  string
	hash  string
	url   string
}

func NewOsonSMSNotifier(
	repo *repository.Queries,
	login string,
	from string,
	hash string,
	url string,
) *OsonSMSNotifier {
	return &OsonSMSNotifier{
		repo:  repo,
		login: login,
		from:  from,
		hash:  hash,
		url:   url,
	}
}

func (n *OsonSMSNotifier) generateSha256Hash(content string) string {
	hashInBytes := sha256.Sum256([]byte(content))
	hashInHex := hex.EncodeToString(hashInBytes[:])

	return hashInHex
}

func (n *OsonSMSNotifier) Send(ctx context.Context, notification repository.Notification) error {
	log.Printf("-----> Sending SMS to [%s] with subject [%s]", notification.Recipient, notification.Subject.String)

	smsCount, err := n.repo.CountSmsSent(ctx)
	if err != nil {
		return fmt.Errorf("failed to count sms sent: %w", err)
	}

	txn_id := 0 + smsCount
	content := fmt.Sprintf("%d;%s;%s;%s;%s", txn_id, n.login, n.from, notification.Recipient, n.hash)
	str_hash := n.generateSha256Hash(content)
	url := fmt.Sprintf(
		"%s?from=%s&msg=%s&login=%s&str_hash=%s&phone_number=%s&txn_id=%d",
		n.url, n.from, notification.Body.String, n.login, str_hash, notification.Recipient, txn_id,
	)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to form SMS request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	return nil
}
