package login

import (
	"testing"

	"zauto/internal/state"
	"zauto/internal/ui"
)

func TestFormReadyTwoEdits(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="61592118521974" enabled="true" bounds="[48,400][672,480]" class="android.widget.EditText"/>
  <node text="" enabled="true" password="true" bounds="[48,520][672,600]" class="android.widget.EditText"/>
</hierarchy>`)
	resolver := ui.NewDefaultResolver()
	if !FormVisibleStrict(resolver, snap) && len(ui.LoginFormEdits(snap)) < 2 {
		t.Fatal("expected login form signals")
	}
}

func TestLoginAccountFinderTextsCatalogued(t *testing.T) {
	if len(state.LoginAccountFinderTexts) == 0 {
		t.Fatal("expected account finder texts in state catalog")
	}
}
