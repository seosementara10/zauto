package state

import "testing"

func TestResetFlowClearsAuthAndUnknown(t *testing.T) {
	m := &DeviceMemory{
		UnknownCount: 3,
		AuthPhase:    AuthPhaseWaitResult,
	}
	m.ResetFlow()
	if m.UnknownCount != 0 || m.AuthPhase != AuthPhaseNone {
		t.Fatalf("ResetFlow: unknown=%d auth=%q", m.UnknownCount, m.AuthPhase)
	}
}
