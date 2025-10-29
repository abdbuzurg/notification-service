package models

type Metadata struct {
	UserID int64  `json:"user_id"`
	Source string `json:"source"`
}

type NotificationRequest struct {
	Type      string   `json:"type"`
	Recipient string   `json:"recipient"`
	Subject   string   `json:"subject"`
	Body      string   `json:"body"`
	Metadata  Metadata `json:"metadata"`
}
