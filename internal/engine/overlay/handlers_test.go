package overlay

import (
	"testing"

	"zauto/internal/state"
)

func TestEveryBlockingOverlayHasHandler(t *testing.T) {
	for _, s := range state.DefaultRegistry.BlockingOverlays() {
		if _, ok := handlers[s]; !ok {
			t.Fatalf("missing handler for blocking overlay %s", s)
		}
	}
}
