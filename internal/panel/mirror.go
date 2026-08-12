package panel

import (
	"log"
	"sync"
	"time"

	"zauto/internal/monitor"
)

const (
	mirrorSyncDelay  = 200 * time.Millisecond
	mirrorLaunchWait = 5 * time.Second
)

func (s *Server) mirrorOpts() monitor.Options {
	opts := monitor.FarmOptions(s.ProjectRoot)
	opts.StartX = s.panel.mirrorStartX()
	return opts
}

func (s *Server) mirrorLayoutStartX(deviceCount, tileW int) int {
	fallback := s.panel.mirrorStartX()
	if s.panel.hasLive() {
		return fallback
	}
	return mirrorStartXHeadless(deviceCount, tileW, fallback)
}

func (s *Server) mirrorScreenSizeLocked() (w, h int) {
	for _, d := range s.devices {
		if s.enabled[d.Serial] && d.Resolution != "" {
			return monitor.ParseResolution(d.Resolution)
		}
	}
	return 720, 1600
}

// enabledSerialsOrderedLocked returns enabled serials in device-list order. Caller must hold s.mu.
func (s *Server) enabledSerialsOrderedLocked() []string {
	var out []string
	seen := map[string]bool{}
	for _, d := range s.devices {
		if s.enabled[d.Serial] {
			out = append(out, d.Serial)
			seen[d.Serial] = true
		}
	}
	for serial, on := range s.enabled {
		if on && !seen[serial] {
			out = append(out, serial)
		}
	}
	return out
}

func (s *Server) setMirrorErrorLocked(serial, msg string) {
	if msg == "" {
		delete(s.mirrorErrors, serial)
		return
	}
	s.mirrorErrors[serial] = msg
}

func (s *Server) setMirrorError(serial, msg string) {
	s.mirrorMu.Lock()
	defer s.mirrorMu.Unlock()
	s.setMirrorErrorLocked(serial, msg)
}

func (s *Server) mirrorErrorForSerial(serial string) string {
	s.mirrorMu.Lock()
	defer s.mirrorMu.Unlock()
	return s.mirrorErrors[serial]
}

func (s *Server) mirrorOpenForSerial(serial string) bool {
	s.mirrorMu.Lock()
	defer s.mirrorMu.Unlock()
	return s.mirrorOpenForSerialLocked(serial)
}

func (s *Server) mirrorOpenForSerialLocked(serial string) bool {
	if cmd, ok := s.mirrors[serial]; ok && cmd != nil && monitor.ProcessAlive(cmd) {
		return true
	}
	return monitor.ScrcpyRunningForSerial(serial)
}

// mirrorRunningLocked reports whether scrcpy is actually running (not merely launching).
func (s *Server) mirrorRunningLocked(serial string) bool {
	s.mirrorMu.Lock()
	defer s.mirrorMu.Unlock()
	return s.mirrorOpenForSerialLocked(serial)
}

func (s *Server) mirrorAliveLocked(serial string) bool {
	s.mirrorMu.Lock()
	defer s.mirrorMu.Unlock()
	if s.mirrorLaunching[serial] {
		return true
	}
	return s.mirrorOpenForSerialLocked(serial)
}

func (s *Server) trackMirrorAliveLocked(serial string, slot int) {
	s.mirrorSlot[serial] = slot
	monitor.DedupeScrcpyForSerial(serial)
}

type mirrorJob struct {
	serial string
	slot   int
	tile   monitor.WindowTile
}

// applyMirrorLayoutLocked syncs scrcpy windows with enabled devices.
func (s *Server) applyMirrorLayoutLocked() {
	s.mirrorMu.Lock()

	s.mu.RLock()
	shutting := s.shuttingDown
	serials := s.enabledSerialsOrderedLocked()
	sw, sh := s.mirrorScreenSizeLocked()
	s.mu.RUnlock()
	if shutting {
		s.mirrorMu.Unlock()
		return
	}

	if len(serials) == 0 {
		for serial := range s.mirrors {
			s.stopMirrorLocked(serial)
		}
		for serial := range s.mirrorLaunching {
			delete(s.mirrorLaunching, serial)
		}
		s.syncMirrorOpenFlagsLocked()
		s.mirrorMu.Unlock()
		s.broadcastState()
		return
	}

	opts := s.mirrorOpts()
	tileW, _ := monitor.ContentSize(sw, sh, opts.MaxSize)
	startX := s.mirrorLayoutStartX(len(serials), tileW)
	opts.StartX = startX
	log.Printf("Mirror sync: %d HP switch ON @ x=%d", len(serials), startX)
	tiles := monitor.ComputeTiles(sw, sh, opts.MaxSize, len(serials), startX, opts.StartY)

	desired := make(map[string]int, len(serials))
	for i, serial := range serials {
		desired[serial] = i
	}

	for serial := range s.mirrors {
		if _, ok := desired[serial]; !ok {
			s.stopMirrorLocked(serial)
		}
	}

	var jobs []mirrorJob
	for i, serial := range serials {
		slot := i
		if s.mirrorOpenForSerialLocked(serial) || s.mirrorLaunching[serial] {
			if s.mirrorOpenForSerialLocked(serial) {
				s.trackMirrorAliveLocked(serial, slot)
			}
			continue
		}
		jobs = append(jobs, mirrorJob{serial: serial, slot: slot, tile: tiles[slot]})
	}

	if len(jobs) == 0 {
		if len(tiles) > 0 {
			log.Printf("Mirror layout: %d HP aktif (tanpa restart)", len(serials))
		}
		s.syncMirrorOpenFlagsLocked()
		s.mirrorMu.Unlock()
		s.broadcastState()
		return
	}

	scrcpy, err := monitor.FindScrcpy(opts.ProjectRoot)
	if err != nil {
		log.Printf("Mirror: scrcpy tidak ditemukan: %v", err)
		for _, job := range jobs {
			s.setMirrorErrorLocked(job.serial, "scrcpy tidak ditemukan")
		}
		s.syncMirrorOpenFlagsLocked()
		s.mirrorMu.Unlock()
		s.broadcastState()
		return
	}
	scrcpyDir := monitor.ScrcpyDir(scrcpy)

	for _, job := range jobs {
		s.mirrorLaunching[job.serial] = true
	}
	s.syncMirrorOpenFlagsLocked()
	s.mirrorMu.Unlock()
	s.broadcastState()

	var wg sync.WaitGroup
	for _, job := range jobs {
		wg.Add(1)
		go func(job mirrorJob) {
			defer wg.Done()
			s.launchOneMirror(scrcpy, scrcpyDir, job, opts)
		}(job)
	}
	wg.Wait()

	s.mirrorMu.Lock()
	if len(tiles) > 0 && len(jobs) > 0 {
		log.Printf("Mirror layout: %d HP switch ON, %d mirror diluncurkan @ x=%d.. (%dx%d)", len(serials), len(jobs), startX, tiles[0].W, tiles[0].H)
	}
	s.syncMirrorOpenFlagsLocked()
	s.mirrorMu.Unlock()
	s.broadcastState()
}

func (s *Server) launchOneMirror(scrcpy, scrcpyDir string, job mirrorJob, opts monitor.Options) {
	if job.slot > 0 {
		time.Sleep(time.Duration(job.slot) * 600 * time.Millisecond)
	}

	s.mu.RLock()
	shutting := s.shuttingDown
	enabled := s.enabled[job.serial]
	s.mu.RUnlock()
	if shutting || !enabled {
		s.mirrorMu.Lock()
		delete(s.mirrorLaunching, job.serial)
		s.syncMirrorOpenFlagsLocked()
		s.mirrorMu.Unlock()
		s.broadcastState()
		return
	}

	s.mirrorMu.Lock()
	if s.mirrorOpenForSerialLocked(job.serial) {
		delete(s.mirrorLaunching, job.serial)
		s.syncMirrorOpenFlagsLocked()
		s.mirrorMu.Unlock()
		s.broadcastState()
		return
	}
	s.mirrorMu.Unlock()

	cmd, err := monitor.StartOneAtWith(scrcpy, scrcpyDir, job.serial, job.slot+1, job.tile, opts)
	if err != nil {
		log.Printf("Mirror gagal %s: %v", job.serial, err)
		s.mirrorMu.Lock()
		s.setMirrorErrorLocked(job.serial, "gagal buka scrcpy")
		delete(s.mirrorLaunching, job.serial)
		s.syncMirrorOpenFlagsLocked()
		s.mirrorMu.Unlock()
		s.broadcastState()
		return
	}

	deadline := time.Now().Add(mirrorLaunchWait)
	for time.Now().Before(deadline) {
		if monitor.ProcessAlive(cmd) || monitor.ScrcpyRunningForSerial(job.serial) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	s.mirrorMu.Lock()
	defer s.mirrorMu.Unlock()

	if !monitor.ProcessAlive(cmd) && !monitor.ScrcpyRunningForSerial(job.serial) {
		log.Printf("Mirror %s: scrcpy tidak muncul — cek ADB/USB", job.serial[len(job.serial)-8:])
		s.setMirrorErrorLocked(job.serial, "scrcpy exit — cek ADB/USB")
		delete(s.mirrorLaunching, job.serial)
		s.syncMirrorOpenFlagsLocked()
		go s.broadcastState()
		return
	}

	monitor.BringScrcpyWindowToFront(job.serial, job.slot+1)
	s.setMirrorErrorLocked(job.serial, "")
	s.mirrors[job.serial] = cmd
	s.mirrorSlot[job.serial] = job.slot
	delete(s.mirrorLaunching, job.serial)
	s.syncMirrorOpenFlagsLocked()
	go s.broadcastState()
}

// syncMirrorOpenFlagsLocked updates device mirror flags. Caller must hold mirrorMu.
func (s *Server) syncMirrorOpenFlagsLocked() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.devices {
		serial := s.devices[i].Serial
		if !s.enabled[serial] {
			s.devices[i].MirrorOpen = false
			s.devices[i].MirrorError = ""
			continue
		}
		if s.mirrorLaunching[serial] {
			s.devices[i].MirrorOpen = false
			s.devices[i].MirrorError = ""
			continue
		}
		if s.mirrorOpenForSerialLocked(serial) {
			s.devices[i].MirrorOpen = true
			s.devices[i].MirrorError = ""
			continue
		}
		s.devices[i].MirrorOpen = false
		errMsg := s.mirrorErrors[serial]
		if errMsg == "" {
			errMsg = "mirror belum terbuka"
		}
		s.devices[i].MirrorError = errMsg
	}
	s.bumpStateRevLocked()
}

func (s *Server) stopMirrorLocked(serial string) {
	if cmd := s.mirrors[serial]; cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	monitor.StopScrcpyForSerial(serial)
	delete(s.mirrors, serial)
	delete(s.mirrorSlot, serial)
	delete(s.mirrorLaunching, serial)
	s.setMirrorErrorLocked(serial, "")
}

func (s *Server) syncMirrors() {
	s.applyMirrorLayoutLocked()
}

func (s *Server) runMirrorSyncOnce() {
	s.syncMirrors()
}

func (s *Server) requestSyncMirrors() {
	s.mirrorSyncMu.Lock()
	defer s.mirrorSyncMu.Unlock()

	if s.mirrorSyncPending {
		s.mirrorSyncAgain = true
		return
	}
	if s.mirrorDebounce != nil {
		s.mirrorDebounce.Stop()
	}
	s.mirrorDebounce = time.AfterFunc(mirrorSyncDelay, func() {
		s.mirrorSyncMu.Lock()
		s.mirrorDebounce = nil
		s.mirrorSyncPending = true
		s.mirrorSyncMu.Unlock()

		s.runMirrorSyncOnce()

		s.mirrorSyncMu.Lock()
		again := s.mirrorSyncAgain
		s.mirrorSyncAgain = false
		s.mirrorSyncPending = false
		s.mirrorSyncMu.Unlock()
		if again {
			s.requestSyncMirrors()
		}
	})
}

func (s *Server) stopAllMirrors() {
	s.mirrorMu.Lock()
	defer s.mirrorMu.Unlock()
	for serial := range s.mirrors {
		s.stopMirrorLocked(serial)
	}
}

func (s *Server) relayoutMirrorsLocked() (moved, alive, restarted int) {
	// Caller must hold s.mirrorMu.
	s.mu.RLock()
	serials := s.enabledSerialsOrderedLocked()
	sw, sh := s.mirrorScreenSizeLocked()
	s.mu.RUnlock()

	if len(serials) == 0 {
		return 0, 0, 0
	}

	opts := s.mirrorOpts()
	tileW, _ := monitor.ContentSize(sw, sh, opts.MaxSize)
	startX := s.mirrorLayoutStartX(len(serials), tileW)
	tiles := monitor.ComputeTiles(sw, sh, opts.MaxSize, len(serials), startX, opts.StartY)
	hpNums := make(map[string]int, len(serials))
	for i, serial := range serials {
		hpNums[serial] = i + 1
	}

	movedFlags := monitor.RelayoutScrcpyWindows(serials, tiles, hpNums)
	for i, serial := range serials {
		if !s.mirrorOpenForSerialLocked(serial) {
			continue
		}
		alive++
		if i < len(movedFlags) && movedFlags[i] {
			moved++
			continue
		}
		s.stopMirrorLocked(serial)
		restarted++
	}
	return moved, alive, restarted
}
