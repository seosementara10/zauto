package ui

import "testing"

func TestFieldHasExpectedValue(t *testing.T) {
	field := Element{Text: "61592118521974"}
	if !FieldHasExpectedValue(field, "61592118521974") {
		t.Fatal("exact match should pass")
	}
	dup := Element{Text: "6159211852197461592118521974"}
	if FieldHasExpectedValue(dup, "61592118521974") {
		t.Fatal("duplicated value should fail")
	}
	partial := Element{Text: "61592118521974extra"}
	if FieldHasExpectedValue(partial, "61592118521974") {
		t.Fatal("extra suffix should fail")
	}
}

func TestFieldNeedsClear(t *testing.T) {
	empty := Element{Text: ""}
	if FieldNeedsClear(empty, "abc") {
		t.Fatal("empty field should not need clear")
	}
	ok := Element{Text: "abc"}
	if FieldNeedsClear(ok, "abc") {
		t.Fatal("correct value should not need clear")
	}
	dup := Element{Text: "abcabc"}
	if !FieldNeedsClear(dup, "abc") {
		t.Fatal("duplicate should need clear")
	}
	// Password fields use content-desc as label; must not trigger clear/backspace.
	pw := Element{Password: true, Text: "", ContentDesc: "Kata sandi,"}
	if FieldNeedsClear(pw, "secret123") {
		t.Fatal("password field label must not trigger clear")
	}
}
