package audit_listeners

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/al-tokarev/shortener/internal/observer/events"
)

type AuditURLListener struct {
	URL string
}

func (l AuditURLListener) Update(event interface{}) error {
	auditEvent, ok := event.(events.AuditEvent)
	if !ok {
		return fmt.Errorf("wrong event type")
	}

	eData, err := json.Marshal(auditEvent)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", l.URL, bytes.NewBuffer(eData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
