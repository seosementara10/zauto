package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"

	"zauto/internal/config"
	"zauto/internal/pool"
	"zauto/internal/store"
	"zauto/internal/worker"
)

// Scheduler distributes jobs across the device pool with bounded concurrency.
type Scheduler struct {
	Workflow    *config.Workflow
	Pool        *pool.Pool
	Workers     int
	ProjectRoot string
	LogDir      string
	RunID       string
	Store       *store.Store
}

func (s *Scheduler) Run(ctx context.Context) []worker.Result {
	n := s.Pool.Count()
	workers := s.Workers
	if workers <= 0 {
		workers = n
	}
	if workers > n {
		workers = n
	}

	log.Printf("Scheduler: %d device(s), %d worker(s), driver=%s",
		n, workers, s.Workflow.Automation.Driver)

	results := make([]worker.Result, n)
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	for _, dev := range s.Pool.All() {
		wg.Add(1)
		go func(d *pool.Device) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			select {
			case <-ctx.Done():
				results[d.Index] = worker.Result{Serial: d.Serial, Errors: []string{ctx.Err().Error()}}
				return
			default:
			}

			d.MarkBusy()
			r := worker.RunDevice(ctx, s.Workflow, d, s.ProjectRoot, s.LogDir, s.RunID, s.Store)
			if len(r.Errors) > 0 {
				s.Pool.Release(d, fmt.Errorf("task errors"))
			} else {
				s.Pool.MarkIdle(d)
			}
			results[d.Index] = r
		}(dev)
	}
	wg.Wait()
	return results
}
