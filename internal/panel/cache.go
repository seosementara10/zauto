package panel

import (
	"os"
	"sync"
	"time"

	"zauto/internal/config"
)

type resourceCache struct {
	mu sync.Mutex

	wfPath string
	wfMod  time.Time
	wf     *config.Workflow
	wfErr  error
}

func (s *Server) loadWorkflow() (*config.Workflow, error) {
	s.resCache.mu.Lock()
	defer s.resCache.mu.Unlock()

	info, err := os.Stat(s.ConfigPath)
	if err != nil {
		return nil, err
	}
	if s.resCache.wf != nil && s.resCache.wfPath == s.ConfigPath && s.resCache.wfMod.Equal(info.ModTime()) {
		return s.resCache.wf, s.resCache.wfErr
	}
	wf, err := config.Load(s.ConfigPath)
	s.resCache.wfPath = s.ConfigPath
	s.resCache.wfMod = info.ModTime()
	s.resCache.wf = wf
	s.resCache.wfErr = err
	return wf, err
}

func (s *Server) invalidateWorkflowCache() {
	s.resCache.mu.Lock()
	defer s.resCache.mu.Unlock()
	s.resCache.wf = nil
	s.resCache.wfMod = time.Time{}
}
