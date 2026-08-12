package runtime

import (
	"time"

	"zauto/internal/state"
)

// ScreenNote carries optional caller context for SCREEN / CAPTURE logs.
type ScreenNote struct {
	Profile string // personal | fanpage | unknown
	Detail  string
}

// ScreenLogger is implemented by engine.Executor for unified diagnostics.
type ScreenLogger interface {
	LogScreen(observe state.ObserveFn, where string, note ScreenNote)
	LogScreenIfStale(observe state.ObserveFn, where string, note ScreenNote, minInterval time.Duration)
	CaptureFailure(label, where, detail string, observe state.ObserveFn, note ScreenNote) (screenshotPath, dumpPath string)
}
