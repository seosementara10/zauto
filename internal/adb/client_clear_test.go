package adb

import "testing"

func TestPmClearSucceeded(t *testing.T) {
	if !pmClearSucceeded("Success\n") {
		t.Fatal("expected success")
	}
	if pmClearSucceeded("Error: package not found") {
		t.Fatal("expected failure")
	}
}
