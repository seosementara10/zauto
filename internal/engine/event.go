package engine

import (
	"fmt"
	"log"
)

func (e *Executor) event(msg string, args ...interface{}) {
	line := fmt.Sprintf(msg, args...)
	if e.Session.DevLog != nil {
		e.Session.DevLog.Info("EVENT %s", line)
		return
	}
	log.Printf("[%s] EVENT %s", e.Session.Serial, line)
}
