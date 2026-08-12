package ui

import "testing"

func TestComposerEditFieldFindsEditText(t *testing.T) {
	snap := ParseHierarchy(`<hierarchy>
		<node text="Buat postingan" bounds="[0,0][720,120]" class="android.widget.TextView"/>
		<node text="Apa yang Anda pikirkan?" bounds="[24,200][696,400]" class="android.widget.EditText" enabled="true"/>
	</hierarchy>`)
	edit := ComposerEditField(snap)
	if edit == nil {
		t.Fatal("expected EditText")
	}
	if edit.Text == "" {
		t.Fatalf("edit=%+v", edit)
	}
}

func TestComposerEditFieldFindsAutoCompleteTextView(t *testing.T) {
	snap := ParseHierarchy(`<hierarchy>
		<node text="Postingan baru" bounds="[96,80][340,168]" class="android.widget.TextView"/>
		<node text="Apa yang Anda pikirkan?" bounds="[0,400][720,468]" class="android.widget.AutoCompleteTextView" enabled="true"/>
	</hierarchy>`)
	edit := ComposerEditField(snap)
	if edit == nil {
		t.Fatal("expected AutoCompleteTextView")
	}
}
