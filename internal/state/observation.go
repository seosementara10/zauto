package state

import "time"

// ObservationSnapshot captures evidence when UI is unknown — for post-mortem review.
type ObservationSnapshot struct {
	Timestamp     time.Time          `json:"timestamp"`
	Serial        string             `json:"serial"`
	Package       string             `json:"package"`
	UIState       UIState            `json:"ui_state"`
	UnknownKind   UnknownKind        `json:"unknown_kind,omitempty"`
	Confidence    float64            `json:"confidence"`
	PreviousState UIState            `json:"previous_state,omitempty"`
	LastAction    string             `json:"last_action,omitempty"`
	Probes        map[UIState]float64 `json:"probes,omitempty"`
	Hints         []string           `json:"hints,omitempty"`
	Screenshot    string             `json:"screenshot,omitempty"`
	Hierarchy     string             `json:"hierarchy,omitempty"`
}

func BuildObservationSnapshot(
	serial string,
	mem *DeviceMemory,
	inv Investigation,
	screenshotPath, hierarchyPath string,
) ObservationSnapshot {
	prev := UIState("")
	lastAction := ""
	conf := 0.0
	uiState := UIUnknown
	kind := UnknownKindNone
	if mem != nil {
		prev = mem.PreviousUI
		lastAction = mem.LastAction
		conf = inv.Detection.Confidence
		uiState = inv.Detection.State
		kind = inv.Detection.UnknownKind
	}
	return ObservationSnapshot{
		Timestamp:     time.Now(),
		Serial:        serial,
		Package:       inv.Detection.Package,
		UIState:       uiState,
		UnknownKind:   kind,
		Confidence:    conf,
		PreviousState: prev,
		LastAction:    lastAction,
		Probes:        inv.Probes,
		Hints:         inv.Hints,
		Screenshot:    screenshotPath,
		Hierarchy:     hierarchyPath,
	}
}
