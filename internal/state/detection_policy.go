package state

// DetectionPolicy is the single source for confidence thresholds across detector and engine.
type DetectionPolicy struct {
	VerifyThreshold      float64
	ExecuteThreshold     float64
	InvestigateThreshold float64
}

const (
	verifyMinConfidence      = 0.70
	executeMinConfidence     = 0.90
	investigateMinConfidence = 0.55
)

// DefaultDetectionPolicy is used by detector scoring and engine loops.
var DefaultDetectionPolicy = DetectionPolicy{
	VerifyThreshold:      verifyMinConfidence,
	ExecuteThreshold:     executeMinConfidence,
	InvestigateThreshold: investigateMinConfidence,
}

func (p DetectionPolicy) MinScorePoints() float64 { return p.VerifyThreshold * 100 }

const (
	VerifyMinConfidence      = verifyMinConfidence
	ExecuteMinConfidence     = executeMinConfidence
	InvestigateMinConfidence = investigateMinConfidence
)
