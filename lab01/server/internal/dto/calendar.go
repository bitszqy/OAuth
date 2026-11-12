package dto

type PlainCalendarEvent struct {
	ID          string   `json:"id"`
	ResourceID  string   `json:"resource_id"`
	Start       string   `json:"start"`
	End         string   `json:"end"`
	Title       string   `json:"title"`
	Location    string   `json:"location"`
	Attendees   []string `json:"attendees"`
	Description string   `json:"description"`
}

type EncryptedField struct {
	Algorithm  string `json:"algorithm"`
	KeyID      string `json:"key_id"`
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
}

type EncryptedCalendarEvent struct {
	ID        string                    `json:"id"`
	ResourceID string                   `json:"resource_id"`
	Fields    map[string]EncryptedField `json:"fields"`
}

type CalendarResponse struct {
	Events    []PlainCalendarEvent     `json:"events,omitempty"`
	Encrypted []EncryptedCalendarEvent `json:"encrypted,omitempty"`
}

type TaskKeyRequest struct {
	TaskID     string `json:"task_id" binding:"required"`
	ResourceID string `json:"resource_id" binding:"required"`
}

type FieldKey struct {
	Algorithm string `json:"algorithm"`
	KeyID     string `json:"key_id"`
	Key       string `json:"key"`
	AAD       string `json:"aad"`
}

type TaskKeyResponse struct {
	TaskID            string              `json:"task_id"`
	AllowedAttributes []string            `json:"allowed_attributes"`
	Keys              map[string]FieldKey `json:"keys"`
}

type TaskGrant struct {
	UserID   string
	ClientID string
	TaskID   string
}
