package ui

import "testing"

func TestLoginFormEditsOrdersByY(t *testing.T) {
	snap := Snapshot{
		Elements: []Element{
			{ClassName: "android.widget.EditText", Enabled: true, Package: "com.facebook.katana", Bounds: [4]int{50, 400, 350, 460}},
			{ClassName: "android.widget.EditText", Enabled: true, Package: "com.facebook.katana", Bounds: [4]int{50, 250, 350, 310}},
			{ClassName: "android.widget.Button", Enabled: true, Bounds: [4]int{50, 500, 350, 560}},
			{ClassName: "android.widget.EditText", Enabled: true, Bounds: [4]int{50, 50, 80, 70}}, // too small
		},
	}
	edits := LoginFormEdits(snap)
	if len(edits) != 2 {
		t.Fatalf("got %d edits, want 2", len(edits))
	}
	if edits[0].Bounds[1] != 250 || edits[1].Bounds[1] != 400 {
		t.Fatalf("wrong order: %+v", edits)
	}
}

func TestLoginFormEditsIncludesPasswordBelow900(t *testing.T) {
	snap := Snapshot{
		Elements: []Element{
			{ClassName: "android.widget.EditText", Enabled: true, Package: "com.facebook.katana", Bounds: [4]int{64, 837, 584, 879}, ContentDesc: "Nomor ponsel atau email,"},
			{ClassName: "android.widget.EditText", Enabled: true, Password: true, Package: "com.facebook.katana", Bounds: [4]int{64, 981, 584, 1023}, ContentDesc: "Kata sandi,"},
		},
	}
	edits := LoginFormEdits(snap)
	if len(edits) != 2 {
		t.Fatalf("got %d edits, want 2 (password below y=900 must be included)", len(edits))
	}
	if !edits[1].Password {
		t.Fatal("expected password field as second edit")
	}
}

func TestFindLoginPasswordFieldByContentDesc(t *testing.T) {
	snap := Snapshot{
		Elements: []Element{
			{ClassName: "android.widget.EditText", Enabled: true, Package: "com.facebook.katana", Bounds: [4]int{64, 837, 584, 879}, ContentDesc: "Nomor ponsel atau email,"},
			{ClassName: "android.widget.EditText", Enabled: true, Password: true, Package: "com.facebook.katana", Bounds: [4]int{64, 981, 584, 1023}, ContentDesc: "Kata sandi,"},
		},
	}
	resolver := NewDefaultResolver()
	got := FindLoginPasswordField(resolver, snap, []string{"Kata sandi", "Password"}, nil)
	if got == nil || !got.Element.Password {
		t.Fatal("expected password field by content-desc")
	}
}
