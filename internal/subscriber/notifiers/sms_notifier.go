package notifiers

import (
	"asr_leasing_notification/internal/subscriber/repository"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/viper"
)

type SmsNotifer struct {
	login    string
	password string
	baseUrl  string
}

func NewSmsNotifier() *SmsNotifer {
	return &SmsNotifer{
		login:    viper.GetString("sms.login"),
		password: viper.GetString("sms.password"),
		baseUrl:  viper.GetString("sms.baseUrl"),
	}
}

type SmsBody struct {
	SourceAddress      string `json:"SourceAddress"`
	DestinationAddress string `json:"DestinationAddress"`
	MessageText        string `json:"MessageText"`
	CompanyLogin       string `json:"CompanyLogin"`
	CompanyPassword    string `json:"CompanyPassword"`
	MessageId          string `json:"MessageId"`
}

type SmsResponse struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	MessageId   string `json:"messageId"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

func (n *SmsNotifer) Send(ctx context.Context, notification repository.Notification) error {
	log.Printf("-----> Sending SMS to [%s] with subject [%s]", notification.Recipient, notification.Subject.String)

	body := SmsBody{
		SourceAddress:      "ASR Wallet",
		DestinationAddress: notification.Recipient,
		MessageText:        notification.Body.String,
		CompanyLogin:       n.login,
		CompanyPassword:    n.password,
		MessageId:          "",
	}
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshall request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.baseUrl+"/SMS/Send", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request:: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var smsResp SmsResponse
	if err := json.NewDecoder(resp.Body).Decode(&smsResp); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return nil
}
