package runcontrol

import (
	"context"
	"sync"
	"time"

	"zauto/internal/state"
)

// Control coordinates pause, stop, and per-device flags during a run.
type Control struct {
	mu            sync.RWMutex
	globalPaused  bool
	pausedDevices map[string]bool
	cancel        context.CancelFunc
}

// Default is the shared control instance used by panel and workers.
var Default = &Control{pausedDevices: map[string]bool{}}

func (c *Control) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.globalPaused = false
	c.pausedDevices = map[string]bool{}
	c.cancel = nil
}

func (c *Control) PauseAll() {
	c.mu.Lock()
	c.globalPaused = true
	c.mu.Unlock()
}

func (c *Control) ResumeAll() {
	c.mu.Lock()
	c.globalPaused = false
	c.pausedDevices = map[string]bool{}
	c.mu.Unlock()
}

func (c *Control) PauseDevice(serial string) {
	c.mu.Lock()
	c.pausedDevices[serial] = true
	c.mu.Unlock()
}

func (c *Control) ResumeDevice(serial string) {
	c.mu.Lock()
	delete(c.pausedDevices, serial)
	c.mu.Unlock()
}

func (c *Control) IsPaused(serial string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.globalPaused {
		return true
	}
	return c.pausedDevices[serial]
}

func (c *Control) SetCancel(cancel context.CancelFunc) {
	c.mu.Lock()
	c.cancel = cancel
	c.mu.Unlock()
}

func (c *Control) Stop() {
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
	}
	c.globalPaused = false
	c.mu.Unlock()
}

// WaitIfPaused blocks until unpaused or ctx cancelled.
func (c *Control) WaitIfPaused(ctx context.Context, serial string) error {
	for {
		if !c.IsPaused(serial) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(state.DefaultPollInterval):
		}
	}
}
