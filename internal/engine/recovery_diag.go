package engine

import (
	"strings"

	"zauto/internal/state"
	"zauto/internal/ui"
)

func (e *Executor) logRecoveryDiagnostic(snap ui.Snapshot, pkg string, inv state.Investigation) {
	var labels []string
	for _, el := range snap.Elements {
		if lbl := el.Label(); lbl != "" && el.Clickable {
			labels = append(labels, lbl)
			if len(labels) >= 12 {
				break
			}
		}
	}
	e.Event("RECOVERY diagnostic pkg=%s method=%s kind=%s probes=%v clickable=%v",
		pkg, inv.Method, inv.Detection.UnknownKind, inv.Probes, labels)
	if len(snap.XML) > 0 {
		snippet := snap.XML
		if len(snippet) > 400 {
			snippet = snippet[:400] + "..."
		}
		e.Event("RECOVERY hierarchy_snippet=%s", strings.ReplaceAll(snippet, "\n", " "))
	}
}

func (e *Executor) captureFlowTimeout(label string, mem *state.DeviceMemory, det *state.Detector, observe state.ObserveFn) {
	snap, pkg, act := observe()
	inv := det.Investigate(snap, pkg, act)
	e.LogScreen(observe, "timeout:"+label, ScreenNote{Detail: string(mem.UIState)})
	screenshot, dump := e.captureRecoveryArtifacts(label, snap)
	e.Event("TIMEOUT %s serial=%s state=%s kind=%s screenshot=%s hierarchy=%s",
		label, e.Session.Serial, mem.UIState, inv.Detection.UnknownKind, screenshot, dump)
	e.logRecoveryDiagnostic(snap, pkg, inv)
}
