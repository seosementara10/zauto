package panel

import (
	"strings"
	"testing"
)

func TestPanelHTMLStructure(t *testing.T) {
	html := panelHTML()
	if !strings.Contains(html, `id="pages"`) || strings.Contains(html, "page-dashboard") {
		t.Fatal("layout should expose empty #pages container, not inline page views")
	}
	if !strings.Contains(html, "preflightBanner") {
		t.Fatal("preflight banner missing from panel layout")
	}
	if !strings.Contains(html, "/assets/styles.css") || !strings.Contains(html, "/assets/tailwind.css") || !strings.Contains(html, "/assets/app.js") {
		t.Fatal("asset links missing from panel HTML")
	}
	if !strings.Contains(html, "lucide.min.js") {
		t.Fatal("Lucide CDN script missing from panel HTML")
	}
}

func TestPanelPageFragments(t *testing.T) {
	checks := map[string]string{
		"dashboard":  "dash-automation-btn",
		"devices":    "dev-stats-grid",
		"accounts":   "accountsPanel",
		"skrip":      "taskList",
		"text":       "textList",
		"settings":   "btnSaveSettings",
		"kontrol":    "page-kontrol",
		"log":        "logPanel",
	}
	for name, needle := range checks {
		html := panelPageHTML(name)
		if html == "" || !strings.Contains(html, needle) {
			t.Fatalf("page %q missing %q", name, needle)
		}
	}
}

func TestPanelCSSSubstitutesWindowSize(t *testing.T) {
	css := panelCSS()
	if strings.Contains(css, "{{PW}}") || strings.Contains(css, "{{WH}}") {
		t.Fatal("unsubstituted template placeholders in panel CSS")
	}
	if strings.Contains(css, "%!d") {
		t.Fatal("fmt corruption detected in panel CSS")
	}
	if !strings.Contains(css, "@media (min-width: 960px)") {
		t.Fatal("responsive wide layout CSS missing")
	}
}

func TestPanelJSFeatures(t *testing.T) {
	js := panelJS()
	if !strings.Contains(js, "function showPage(") {
		t.Fatal("showPage missing — tab navigation may be broken")
	}
	if !strings.Contains(js, "loadPages") {
		t.Fatal("loadPages missing — split page fragments may not load")
	}
	if strings.Contains(js, "fitPanelWindow") {
		t.Fatal("fitPanelWindow should be removed — Wails owns window size")
	}
	if !strings.Contains(js, "Mirror aktif") {
		t.Fatal("device row template missing active mirror badge")
	}
	if !strings.Contains(js, "devices/disable-all") {
		t.Fatal("disable-all API missing from panel UI")
	}
	if !strings.Contains(js, "devices/mirror-retry") {
		t.Fatal("mirror-retry API missing from panel UI")
	}
	if !strings.Contains(js, "run/ack") {
		t.Fatal("run ack missing from panel UI")
	}
	if !strings.Contains(js, "accounts/assign") {
		t.Fatal("accounts assign missing from panel UI")
	}
	if !strings.Contains(js, "confirmDisableAll") {
		t.Fatal("disable-all confirmation missing from panel UI")
	}
	if !strings.Contains(js, "mirror-relayout") {
		t.Fatal("mirror relayout missing from panel UI")
	}
	if !strings.Contains(js, "setupKeyboardShortcuts") {
		t.Fatal("keyboard shortcuts missing from panel UI")
	}
	if !strings.Contains(js, "import-preview") {
		t.Fatal("import preview missing from panel UI")
	}
	if !strings.Contains(js, "fetchServerLog") {
		t.Fatal("server log tab missing from panel UI")
	}
	if !strings.Contains(js, "registerDevice") || !strings.Contains(js, "createAccount") {
		t.Fatal("manual device/account forms missing from panel UI")
	}
	if !strings.Contains(js, "loadPostTexts") {
		t.Fatal("text page loadPostTexts missing from panel UI")
	}
}

func TestPanelSettingsPage(t *testing.T) {
	html := panelPageHTML("settings")
	js := panelJS()
	if !strings.Contains(html, "page-settings") || !strings.Contains(js, "saveSettings") {
		t.Fatal("settings page missing from panel UI")
	}
}
