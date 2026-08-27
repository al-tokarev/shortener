package audit_listeners

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/al-tokarev/shortener/internal/observer/events"
)

type AuditFileListener struct {
	file   *os.File
	writer *bufio.Writer
	mtx    *sync.Mutex
}

func NewAuditFileListener(f *os.File) *AuditFileListener {
	return &AuditFileListener{
		file:   f,
		writer: bufio.NewWriter(f),
		mtx:    &sync.Mutex{},
	}
}

func (l *AuditFileListener) Update(event interface{}) error {
	auditEvent, ok := event.(events.AuditEvent)
	if !ok {
		return fmt.Errorf("wrong event type")
	}

	eData, err := json.Marshal(auditEvent)
	if err != nil {
		return err
	}

	l.mtx.Lock()
	defer l.mtx.Unlock()

	if _, err := l.writer.Write(eData); err != nil {
		return err
	}
	if err := l.writer.WriteByte('\n'); err != nil {
		return err
	}

	return l.writer.Flush()
}

func (l *AuditFileListener) Close() error {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	if l.writer != nil {
		if err := l.writer.Flush(); err != nil {
			return err
		}
	}

	if l.file != nil {
		return l.file.Close()
	}

	return nil
}
