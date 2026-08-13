package domain

// Notification представляет уведомление
type Notification struct {
	UserID    string                 `json:"user_id"`
	Type      string                 `json:"type"`
	Amount    float64                `json:"amount"`
	Currency  string                 `json:"currency"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}
