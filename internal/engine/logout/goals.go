package logout

import (
	"zauto/internal/engine/login"
	"zauto/internal/state"
	"zauto/internal/ui"
)

func savedProfileVisible(resolver *ui.Resolver, snap ui.Snapshot) bool {
	if resolver == nil {
		resolver = ui.NewDefaultResolver()
	}
	hasContinue := resolver.Find(snap, ui.FindQuery{Texts: state.SavedProfileContinueTexts, PreferClickable: true}) != nil
	hasOther := resolver.Find(snap, ui.FindQuery{Texts: state.SavedProfileOtherTexts}) != nil
	return hasContinue && hasOther
}

func (lf *logoutFlow) successGoal(resolver *ui.Resolver, d state.Detection, snap ui.Snapshot) bool {
	if !lf.keluarTapped {
		return false
	}
	if d.State == state.UILoggedIn && d.Confidence >= state.VerifyMinConfidence {
		return false
	}
	if d.State == state.UISavedProfileScreen && d.Confidence >= state.VerifyMinConfidence {
		return savedProfileVisible(resolver, snap)
	}
	if state.IsOverlayState(d.State) && d.Confidence >= state.VerifyMinConfidence {
		return false
	}
	if d.State != state.UILogin || d.Confidence < state.VerifyMinConfidence {
		return false
	}
	return login.FormVisibleStrict(resolver, snap)
}
