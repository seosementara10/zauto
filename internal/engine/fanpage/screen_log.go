package fanpage

import (
	"time"

	"zauto/internal/engine/post"
	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/store"
)

// screenNoteForPage adds fanpage-specific profile hints on top of engine detector SCREEN logs.
func screenNoteForPage(e runtime.Exec, observe state.ObserveFn, page *store.Fanpage) runtime.ScreenNote {
	note := runtime.ScreenNote{}
	if page == nil {
		return note
	}
	snap := e.ReadSnap(observe)
	_, h := post.ScreenDims(e)
	check := verifyFanpageContext(e, snap, *page, h)
	if check.OnFanpage && check.TargetMatch {
		note.Profile = "fanpage"
		note.Detail = check.Reason + " db=" + check.Matched.Name
		return note
	}
	if check.OnFanpage && !check.TargetMatch {
		note.Profile = "fanpage"
		note.Detail = "wrong_page:" + check.DBCompare
		return note
	}
	if post.OnFeedVisible(e, snap) {
		note.Profile = "personal"
	}
	return note
}

func logScreenContext(e runtime.Exec, observe state.ObserveFn, where string, page *store.Fanpage) {
	e.LogScreen(observe, where, screenNoteForPage(e, observe, page))
}

func logScreenContextIfStale(e runtime.Exec, observe state.ObserveFn, where string, page *store.Fanpage, minInterval time.Duration) {
	e.LogScreenIfStale(observe, where, screenNoteForPage(e, observe, page), minInterval)
}

func capturePostFailure(e runtime.Exec, observe state.ObserveFn, label string, page *store.Fanpage, detail string) {
	e.CaptureFailure(label, label, detail, observe, screenNoteForPage(e, observe, page))
}
