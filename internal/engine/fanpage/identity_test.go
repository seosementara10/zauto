package fanpage

import (
	"testing"

	"zauto/internal/store"
	"zauto/internal/ui"
)

func TestVerifyFanpageContextDBMatchDisplayName(t *testing.T) {
	catalog := []store.Fanpage{
		{FBPageID: "615931763399", Name: "Ibu Nurhayati"},
		{FBPageID: "61592753657118", Name: "Fans Nurhayati"},
	}
	target := catalog[0]
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Ibu Nurhayati" enabled="true" bounds="[24,120][400,160]" class="android.view.ViewGroup"/>
  <node text="Kelola Halaman" enabled="true" bounds="[24,170][300,210]" class="android.widget.Button"/>
</hierarchy>`)
	check := verifyFanpageContextWithCatalog(nil, snap, target, catalog, 1600)
	if !check.OnFanpage || !check.TargetMatch {
		t.Fatalf("expected target fanpage match from DB name in UI, got %+v", check)
	}
}

func TestVerifyFanpageContextUINameMapsViaDBCatalog(t *testing.T) {
	catalog := []store.Fanpage{
		{FBPageID: "615931763399", Name: "Ibu Nurhayati"},
		{FBPageID: "61592753657118", Name: "Fans Nurhayati"},
	}
	target := catalog[0]
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Nurhayati Fans" enabled="true" bounds="[24,120][400,160]" class="android.view.ViewGroup"/>
  <node text="Kelola Halaman" enabled="true" bounds="[24,170][300,210]" class="android.widget.Button"/>
</hierarchy>`)
	check := verifyFanpageContextWithCatalog(nil, snap, target, catalog, 1600)
	if !check.OnFanpage {
		t.Fatalf("expected fanpage detected via DB catalog, got %+v", check)
	}
	if check.TargetMatch {
		t.Fatal("Nurhayati Fans UI should map to Fans Nurhayati in DB, not Ibu Nurhayati target")
	}
	if check.Matched == nil || check.Matched.FBPageID != "61592753657118" {
		t.Fatalf("expected matched Fans Nurhayati page, got %+v", check.Matched)
	}
}

func TestVerifyFanpageContextRejectsInsightOnly(t *testing.T) {
	catalog := []store.Fanpage{{FBPageID: "615931763399", Name: "Ibu Nurhayati"}}
	target := catalog[0]
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Apa yang Anda pikirkan?" enabled="true" bounds="[120,288][600,356]" class="android.widget.AutoCompleteTextView"/>
  <node text="Insight baru" enabled="true" bounds="[200,400][400,600]" class="android.view.ViewGroup"/>
</hierarchy>`)
	check := verifyFanpageContextWithCatalog(nil, snap, target, catalog, 1600)
	if check.OnFanpage || check.TargetMatch {
		t.Fatalf("insight-only personal feed must not pass DB fanpage check, got %+v", check)
	}
}

func TestVerifyFanpageContextWrongFanpageInDB(t *testing.T) {
	catalog := []store.Fanpage{
		{FBPageID: "615931763399", Name: "Ibu Nurhayati"},
		{FBPageID: "61592753657118", Name: "Pengalaman Nurhayati"},
	}
	target := catalog[0]
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Pengalaman Nurhayati" enabled="true" bounds="[24,120][400,160]" class="android.view.ViewGroup"/>
  <node text="Kelola Halaman" enabled="true" bounds="[24,170][300,210]" class="android.widget.Button"/>
</hierarchy>`)
	check := verifyFanpageContextWithCatalog(nil, snap, target, catalog, 1600)
	if !check.OnFanpage {
		t.Fatal("expected on some fanpage")
	}
	if check.TargetMatch {
		t.Fatal("must not match target when UI shows different DB fanpage")
	}
}

func TestVerifyFanpageContextIDInHeader(t *testing.T) {
	catalog := []store.Fanpage{{FBPageID: "615931763399", Name: "Ibu Nurhayati"}}
	target := catalog[0]
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="615931763399" enabled="true" bounds="[24,100][400,140]" class="android.view.ViewGroup"/>
  <node text="Kelola Halaman" enabled="true" bounds="[24,150][300,190]" class="android.widget.Button"/>
</hierarchy>`)
	check := verifyFanpageContextWithCatalog(nil, snap, target, catalog, 1600)
	if !check.TargetMatch {
		t.Fatalf("expected id match in header, got %+v", check)
	}
}

func TestNamesFuzzyMatchReordersWords(t *testing.T) {
	if !namesFuzzyMatch("Nurhayati Fans", "Fans Nurhayati") {
		t.Fatal("expected token reorder match")
	}
	if !namesFuzzyMatch("Nurhayati Fans", "Ibu Nurhayati") {
		t.Fatal("expected shared token match")
	}
}

func TestFanpageNameNoiseRejectsBukaProfil(t *testing.T) {
	if !fanpageNameNoise("Buka profil") {
		t.Fatal("Buka profil must be treated as noise")
	}
}

func TestFindFanpageListEntryByDisplayName(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Nurhayati Guguk" clickable="true" enabled="true" bounds="[24,200][400,260]" class="android.widget.Button"/>
  <node text="Nurhayati Fans" clickable="true" enabled="true" bounds="[24,280][400,340]" class="android.widget.Button"/>
</hierarchy>`)
	page := store.Fanpage{FBPageID: "61592753657118", Name: "Fans Nurhayati"}
	catalog := []store.Fanpage{
		{FBPageID: "615931763399", Name: "Ibu Nurhayati"},
		page,
	}
	hit := findFanpageListEntry(ui.NewDefaultResolver(), snap, catalog, page)
	if hit == nil || hit.Label != "Nurhayati Fans" {
		t.Fatalf("expected Nurhayati Fans row for DB Fans Nurhayati, got %v", hit)
	}
}

func TestComposerAuthorMatchesFanpageName(t *testing.T) {
	snap := ui.ParseHierarchy(`<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy>
  <node text="Nurhayati Fans" content-desc="Nurhayati Fans" enabled="true" bounds="[154,289][394,322]" class="android.widget.Button"/>
</hierarchy>`)
	if !composerAuthorMatches(snap, []string{"Nurhayati Fans", "Ibu Nurhayati"}, 1600) {
		t.Fatal("expected composer author match")
	}
}
