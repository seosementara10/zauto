package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var mu sync.Mutex

// DeviceLogger writes per-device log files under logs/.
type DeviceLogger struct {
	serial string
	file   *os.File
	logger *log.Logger
}

func ForDevice(logDir, serial, runID string) *DeviceLogger {
	_ = os.MkdirAll(logDir, 0755)
	safe := strings.ReplaceAll(serial, ":", "_")
	path := filepath.Join(logDir, fmt.Sprintf("device_%s_%s.log", safe, runID))
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return &DeviceLogger{serial: serial, logger: log.Default()}
	}
	w := io.MultiWriter(f, os.Stdout)
	return &DeviceLogger{
		serial: serial,
		file:   f,
		logger: log.New(w, fmt.Sprintf("[%s] ", serial), log.LstdFlags),
	}
}

func (l *DeviceLogger) Info(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	l.logger.Printf("INFO "+format, args...)
}

func (l *DeviceLogger) Warn(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	l.logger.Printf("WARN "+format, args...)
}

func (l *DeviceLogger) Error(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	l.logger.Printf("ERROR "+format, args...)
}

func (l *DeviceLogger) Close() {
	if l.file != nil {
		_ = l.file.Close()
	}
}
