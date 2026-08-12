package engine

import (
	"time"

	"zauto/internal/state"
	"zauto/internal/ui"
)

const observeCacheExtra = 50 * time.Millisecond

type observeCache struct {
	snap ui.Snapshot
	pkg  string
	act  string
	at   time.Time
}

// cachedObserve deduplicates DumpUI + ForegroundPackage within one poll tick.
// Call invalidate after any tap/input so the next detect sees fresh UI.
func (e *Executor) cachedObserve() (state.ObserveFn, func()) {
	var cache observeCache
	ttl := e.Session.PollInterval() + observeCacheExtra
	invalidate := func() { cache.at = time.Time{} }
	observe := func() (ui.Snapshot, string, string) {
		if !cache.at.IsZero() && time.Since(cache.at) < ttl {
			return cache.snap, cache.pkg, cache.act
		}
		snap := e.Session.ReadUI(true)
		pkg := e.Session.Client.ForegroundPackage()
		cache = observeCache{snap: snap, pkg: pkg, act: "", at: time.Now()}
		return cache.snap, cache.pkg, cache.act
	}
	return observe, invalidate
}

func (e *Executor) invalidateObserve(invalidate func()) {
	if invalidate != nil {
		invalidate()
	}
}

func (e *Executor) deviceIndex() int {
	if v, ok := e.Session.Runtime["device_index"].(int); ok {
		return v
	}
	return 0
}
