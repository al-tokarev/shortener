package audit_listeners

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/al-tokarev/shortener/internal/observer/events"
)

type AuditFileListener struct {
	Path string
}

func (l AuditFileListener) Update(event interface{}) error {
	auditEvent, ok := event.(events.AuditEvent)
	if !ok {
		return fmt.Errorf("wrong event type")
	}

	file, err := os.OpenFile(l.Path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	eData, err := json.Marshal(auditEvent)
	if err != nil {
		return err
	}
	if _, err := w.Write(eData); err != nil {
		return err
	}
	if err := w.WriteByte('\n'); err != nil {
		return err
	}

	return w.Flush()
}
