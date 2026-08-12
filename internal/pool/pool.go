package pool

import (
	"context"
	"sync"
)

// State of a device in the pool.
type State int

const (
	StateIdle State = iota
	StateBusy
	StateError
)

// Device tracks one Android serial in the pool.
type Device struct {
	Serial string
	Index  int
	state  State
	mu     sync.Mutex
}

func (d *Device) State() State {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state
}

func (d *Device) MarkBusy() { d.setState(StateBusy) }

func (d *Device) setState(s State) {
	d.mu.Lock()
	d.state = s
	d.mu.Unlock()
}

// Pool manages available devices for workers (100+ device scale).
type Pool struct {
	devices []*Device
	mu      sync.Mutex
}

func New(serials []string) *Pool {
	devices := make([]*Device, len(serials))
	for i, s := range serials {
		devices[i] = &Device{Serial: s, Index: i, state: StateIdle}
	}
	return &Pool{devices: devices}
}

func (p *Pool) All() []*Device {
	return p.devices
}

func (p *Pool) Count() int { return len(p.devices) }

// Acquire blocks until a device is idle or ctx is cancelled.
func (p *Pool) Acquire(ctx context.Context) (*Device, error) {
	for {
		p.mu.Lock()
		for _, d := range p.devices {
			if d.State() == StateIdle {
				d.setState(StateBusy)
				p.mu.Unlock()
				return d, nil
			}
		}
		p.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		// brief spin — scheduler assigns one device per worker goroutine in practice
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}
}

func (p *Pool) Release(d *Device, err error) {
	if err != nil {
		d.setState(StateError)
	} else {
		d.setState(StateIdle)
	}
}

// MarkIdle resets device after successful run.
func (p *Pool) MarkIdle(d *Device) {
	d.setState(StateIdle)
}
