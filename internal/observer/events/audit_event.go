package events

const (
	Follow  = "follow"
	Shorten = "shorten"
)

type AuditEvent struct {
	Ts     int64  `json:"ts"`
	Action string `json:"action"`
	UserID string `json:"user_id,omitempty"`
	URL    string `json:"url"`
}

func (e AuditEvent) GetName() string {
	return "audit"
}
