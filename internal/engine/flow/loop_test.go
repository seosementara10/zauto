package flow

import (
	"context"
	"testing"
	"time"

	"zauto/internal/state"
	"zauto/internal/ui"
)

// Regression: login screen must reach FillCreds, not DispatchOverlay (no overlay handler for UILogin).
func TestRunLoopLoginStateFillsCredentials(t *testing.T) {
	det := state.NewDetector()
	mem := &state.DeviceMemory{}
	loginSnap := ui.ParseHierarchy(`<hierarchy>
		<node text="Nomor ponsel atau email" bounds="[0,400][720,500]" class="android.widget.EditText"/>
		<node text="Kata sandi" bounds="[0,520][720,620]" class="android.widget.EditText"/>
		<node text="Masuk" clickable="true" bounds="[0,700][720,780]" class="android.widget.Button"/>
	</hierarchy>`)
	loggedInSnap := ui.ParseHierarchy(`<hierarchy>
		<node text="Apa yang Anda pikirkan?" bounds="[0,400][720,460]" class="android.widget.TextView"/>
		<node text="Menu" clickable="true" bounds="[620,100][700,180]" class="android.widget.Button"/>
	</hierarchy>`)
	invalidate := func() {}

	var filled bool
	observe := func() (ui.Snapshot, string, string) {
		if filled {
			return loggedInSnap, "com.facebook.katana", ""
		}
		return loginSnap, "com.facebook.katana", ""
	}
	spec := LoginSpec(func() error {
		filled = true
		return nil
	})

	err := RunLoop(context.Background(), Hooks{
		Event:    func(string, ...interface{}) {},
		Resolver: func() *ui.Resolver { return ui.NewResolver(70) },
		RunRecovery: func(*state.Detector, state.ObserveFn, *state.DeviceMemory, func()) (state.Detection, bool, error) {
			return state.Detection{}, false, nil
		},
		TryResolved: func(d state.Detection) (state.Detection, bool) {
			return d, d.Confidence >= state.VerifyMinConfidence
		},
		DispatchOverlay: func(*state.Detector, state.ObserveFn, state.Detection) error {
			t.Fatal("DispatchOverlay must not run for UILogin")
			return nil
		},
		InvalidateObserve:  func(func()) {},
		CaptureFlowTimeout: func(string, *state.DeviceMemory, *state.Detector, state.ObserveFn) {},
		ErrUnhandledState: func(d state.Detection) error {
			return &unhandledTestError{state: d.State}
		},
	}, det, observe, invalidate, mem, 3*time.Second, state.DefaultPollInterval, spec)

	if err != nil {
		t.Fatalf("RunLoop: %v", err)
	}
	if !filled {
		t.Fatal("expected FillCreds to run on login screen")
	}
}

type unhandledTestError struct{ state state.UIState }

func (e *unhandledTestError) Error() string { return "unhandled:" + string(e.state) }
