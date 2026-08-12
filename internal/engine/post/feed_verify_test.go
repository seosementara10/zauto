package post

import (
	"testing"

	"zauto/internal/state"
	"zauto/internal/ui"
)

func TestFeedPostPublishingDetectsProgressBar(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node class="android.widget.ProgressBar" enabled="true" bounds="[0,900][720,908]"/>
</hierarchy>`)
	if !feedPostPublishing(snap, 1600) {
		t.Fatal("expected progress bar in feed area")
	}
}

func TestFeedFreshPostVisibleBaruSaja(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Baru saja · Dibagikan ke: Teman Anda" enabled="true" bounds="[120,886][400,914]" class="android.view.ViewGroup"/>
</hierarchy>`)
	if !feedFreshPostVisible(snap, 1600) {
		t.Fatal("expected Baru saja timestamp")
	}
}

func TestComposerNeedsNextBeforePublish(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Postingan baru" enabled="true" bounds="[245,105][475,145]" class="android.view.ViewGroup"/>
  <node text="Berikutnya" content-desc="Berikutnya" clickable="true" enabled="true" bounds="[580,95][680,145]" class="android.widget.Button"/>
</hierarchy>`)
	acts := scanComposerActions(snap, 720, 1600)
	if !composerNeedsNextBeforePublish(acts, 1600) {
		t.Fatal("expected Berikutnya before publish on fanpage composer")
	}
}

func TestComposerScreenOpenFalseOnFeedPill(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Apa yang Anda pikirkan?" enabled="true" bounds="[120,288][600,356]" class="android.widget.AutoCompleteTextView"/>
  <node text="Suka" enabled="true" bounds="[117,1261][175,1291]" class="android.view.ViewGroup"/>
  <node text="Menu" enabled="true" bounds="[600,100][700,150]" class="android.view.ViewGroup"/>
</hierarchy>`)
	resolver := ui.NewDefaultResolver()
	if composerScreenOpenSnap(resolver, snap) {
		t.Fatal("feed composer pill must not count as create-post screen")
	}
	if !resolver.TextExists(snap, state.LoggedInFeedHints) {
		t.Fatal("expected feed hints")
	}
}

func TestComposerScreenOpenTrueWithTitle(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Buat postingan" enabled="true" bounds="[96,80][340,168]" class="android.widget.TextView"/>
  <node text="Apa yang Anda pikirkan?" enabled="true" bounds="[0,400][720,468]" class="android.widget.AutoCompleteTextView"/>
</hierarchy>`)
	resolver := ui.NewDefaultResolver()
	if !composerScreenOpenSnap(resolver, snap) {
		t.Fatal("expected create-post screen when title is visible")
	}
}

func TestComposerScreenOpenFalseWhenTitleOnlyInFeedBody(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Apa yang Anda pikirkan?" enabled="true" bounds="[120,288][600,356]" class="android.widget.AutoCompleteTextView"/>
  <node text="Postingan baru" enabled="true" bounds="[120,900][400,940]" class="android.view.ViewGroup"/>
  <node text="Baru saja" enabled="true" bounds="[120,950][200,980]" class="android.view.ViewGroup"/>
</hierarchy>`)
	resolver := ui.NewDefaultResolver()
	if composerScreenOpenSnap(resolver, snap) {
		t.Fatal("Postingan baru in feed body must not count as composer screen")
	}
}
