package login

import (
	"testing"

	"zauto/internal/state"
	"zauto/internal/ui"
)

func loginScreenVisible(resolver *ui.Resolver, d state.Detection, snap ui.Snapshot) bool {
	return d.State == state.UILogin && d.Confidence >= state.VerifyMinConfidence && FormVisibleStrict(resolver, snap)
}

func TestFormVisibleStrictRejectsMasukSubstringAlone(t *testing.T) {
	resolver := ui.NewDefaultResolver()
	snap := ui.Snapshot{
		XML: `<hierarchy><node text="Masukkan kode promo" /><node text="Apa yang Anda pikirkan?" /></hierarchy>`,
	}
	if FormVisibleStrict(resolver, snap) {
		t.Fatal("expected false when login form fields are absent")
	}
}

func TestFormVisibleStrictRejectsFeedXMLPasswordSubstring(t *testing.T) {
	resolver := ui.NewDefaultResolver()
	snap := ui.Snapshot{
		XML: `<hierarchy>
			<node text="Apa yang Anda pikirkan?" />
			<node text="Reset your password today" />
			<node text="contact us by email" />
		</hierarchy>`,
	}
	if FormVisibleStrict(resolver, snap) {
		t.Fatal("expected false when feed XML contains password/email substrings only")
	}
}

func TestLoginScreenVisibleAcceptsLoginForm(t *testing.T) {
	resolver := ui.NewDefaultResolver()
	snap := ui.ParseHierarchy(`<hierarchy>
		<node text="Nomor ponsel atau email" bounds="[0,400][720,500]" class="android.widget.EditText"/>
		<node text="Kata sandi" bounds="[0,520][720,620]" class="android.widget.EditText"/>
		<node text="Masuk" clickable="true" bounds="[0,700][720,780]" class="android.widget.Button"/>
	</hierarchy>`)
	d := state.Detection{State: state.UILogin, Confidence: 1}
	if !loginScreenVisible(resolver, d, snap) {
		t.Fatal("expected true for full login form")
	}
}
