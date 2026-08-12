package login

import (
	"zauto/internal/state"
	"zauto/internal/ui"
)

func FormVisibleStrict(resolver *ui.Resolver, snap ui.Snapshot) bool {
	return state.LoginFormReady(resolver, snap)
}
