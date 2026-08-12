package post

import (
	"time"

	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/store"
)

func logScreenContext(e runtime.Exec, observe state.ObserveFn, where string, page *store.Fanpage) {
	_ = page
	e.LogScreen(observe, where, runtime.ScreenNote{})
}

func logScreenContextIfStale(e runtime.Exec, observe state.ObserveFn, where string, page *store.Fanpage, minInterval time.Duration) {
	_ = page
	e.LogScreenIfStale(observe, where, runtime.ScreenNote{}, minInterval)
}

func capturePostFailure(e runtime.Exec, observe state.ObserveFn, label string, page *store.Fanpage, detail string) {
	_ = page
	e.CaptureFailure(label, label, detail, observe, runtime.ScreenNote{})
}
