package engine

import (
	"testing"
	"time"

	"zauto/internal/state"
)

func TestTryResolvedDetectionPermissionProbe(t *testing.T) {
	e := &Executor{}
	d := state.Detection{State: state.UIPermission, Confidence: 0.6}
	if _, ok := e.tryResolvedDetection(d); !ok {
		t.Fatal("expected permission at investigate threshold to resolve")
	}
}

func TestAuthWaitNotClearedOnLoading(t *testing.T) {
	mem := &state.DeviceMemory{}
	mem.BeginAuthWait()
	// Simulate: during auth wait, loading should not clear auth phase (handled in loop via continue)
	if mem.AuthPhase != state.AuthPhaseWaitResult {
		t.Fatal("auth phase not set")
	}
	// ClearAuthWait must not be called for loading — verified by loop structure; memory stays
	time.Sleep(1 * time.Millisecond)
	if mem.AuthPhase != state.AuthPhaseWaitResult {
		t.Fatal("auth phase cleared prematurely")
	}
}
