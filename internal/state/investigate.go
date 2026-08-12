package state

import (
	"strings"
	"time"

	"zauto/internal/ui"
)

// Investigation is the result of a recovery pass over an uncertain screen.
type Investigation struct {
	Detection Detection
	Method    string
	Hints     []string
	Probes    map[UIState]float64
}

// Investigate re-observes UI hierarchy and tries harder to classify an uncertain screen.
// It never performs taps — investigation only. Single pass over rules (no Detect + reprobe).
func (d *Detector) Investigate(snap ui.Snapshot, pkg, activity string) Investigation {
	inv := Investigation{Probes: map[UIState]float64{}}
	now := time.Now()

	var candidates []Detection
	probeThreshold := func(minScore float64) float64 {
		return minScore * InvestigateMinConfidence / VerifyMinConfidence
	}

	for _, r := range d.rules {
		score, evidence := d.scoreRuleFor(snap, pkg, r)
		conf := score / 100.0
		if conf > 1 {
			conf = 1
		}
		inv.Probes[r.state] = conf
		if score >= probeThreshold(r.minScore) && len(evidence) > 0 {
			inv.Hints = append(inv.Hints, string(r.state)+": "+strings.Join(evidence, ", "))
		}
		if score >= r.minScore {
			candidates = append(candidates, Detection{
				State: r.state, Score: score, Confidence: conf,
				Evidence: evidence, Package: pkg, Activity: activity, At: now,
			})
		}
	}

	primary := pickBestCandidate(candidates)
	if primary.State != UIUnknown && primary.Confidence >= VerifyMinConfidence {
		inv.Detection = primary
		inv.Method = "primary"
		return inv
	}

	best := bestProbeDetection(inv.Probes)
	if best.State != UIUnknown && best.Confidence >= InvestigateMinConfidence {
		best.Package = pkg
		best.Activity = activity
		best.At = now
		inv.Detection = best
		inv.Method = "probe"
		return inv
	}

	if hint := SystemPermissionHint(snap, pkg); hint != "" {
		inv.Hints = append(inv.Hints, hint)
		if perm := inv.Probes[UIPermission]; perm >= InvestigateMinConfidence {
			inv.Detection = Detection{
				State: UIPermission, Confidence: perm, Package: pkg, Activity: activity, At: now,
			}
			inv.Method = "system_permission_package"
			return inv
		}
	}

	if shell := DialogShellClass(snap); shell != "" {
		inv.Hints = append(inv.Hints, "dialog_shell:"+shell)
	}

	var detection Detection
	if primary.State != UIUnknown {
		detection = primary
	} else if best.State != UIUnknown && best.State != "" {
		detection = best
	} else {
		detection = Detection{State: UIUnknown, Confidence: 0, Package: pkg, Activity: activity, At: now}
	}
	inv.Detection = EnrichUnknownDetection(detection, snap, pkg, inv)
	inv.Method = "unresolved"
	return inv
}

func bestProbeDetection(probes map[UIState]float64) Detection {
	var best Detection
	for s, conf := range probes {
		if conf <= 0 {
			continue
		}
		bp, cp := Priority(best.State), Priority(s)
		if best.State == "" || cp > bp || (cp == bp && conf > best.Confidence) {
			best = Detection{State: s, Confidence: conf}
		}
	}
	return best
}
