package auditlisteners

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/al-tokarev/shortener/internal/observer/events"
	"github.com/hashicorp/go-retryablehttp"
)

type AuditURLListener struct {
	url    string
	client *retryablehttp.Client
}

func NewAuditURLListener(url string) *AuditURLListener {
	al := AuditURLListener{
		url:    url,
		client: retryablehttp.NewClient(),
	}
	al.client.RetryMax = 3
	al.client.RetryWaitMin = 1 * time.Second
	al.client.RetryWaitMax = 5 * time.Second
	al.client.HTTPClient.Timeout = 10 * time.Second

	return &al
}

func (l *AuditURLListener) Update(event interface{}) error {
	auditEvent, ok := event.(events.AuditEvent)
	if !ok {
		return fmt.Errorf("wrong event type")
	}

	eData, err := json.Marshal(auditEvent)
	if err != nil {
		return err
	}
	req, err := retryablehttp.NewRequest("POST", l.url, bytes.NewBuffer(eData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := l.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return nil
}

func (l *AuditURLListener) Close() error {
	return nil
}
