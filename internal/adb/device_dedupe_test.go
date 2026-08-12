package adb

import (
	"reflect"
	"testing"
)

func TestEmulatorPeerSerial(t *testing.T) {
	peer, ok := emulatorPeerSerial("emulator-5554")
	if !ok || peer != "127.0.0.1:5555" {
		t.Fatalf("emulator-5554 peer = %q ok=%v", peer, ok)
	}
	peer, ok = emulatorPeerSerial("127.0.0.1:5555")
	if !ok || peer != "emulator-5554" {
		t.Fatalf("127.0.0.1:5555 peer = %q ok=%v", peer, ok)
	}
	if _, ok := emulatorPeerSerial("007519090c2dc80d"); ok {
		t.Fatal("physical serial should not have peer")
	}
}

func TestDedupeDevices(t *testing.T) {
	in := []string{
		"007519090c2dc80d",
		"007519840c4a4c97",
		"00751984103209a4",
		"127.0.0.1:5555",
		"emulator-5554",
	}
	want := []string{
		"007519090c2dc80d",
		"007519840c4a4c97",
		"00751984103209a4",
		"127.0.0.1:5555",
	}
	got := DedupeDevices(in)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DedupeDevices() = %#v, want %#v", got, want)
	}
}

func TestDedupeDevicesNoChange(t *testing.T) {
	in := []string{"007519090c2dc80d", "127.0.0.1:5555"}
	got := DedupeDevices(in)
	if !reflect.DeepEqual(got, in) {
		t.Fatalf("unexpected change: %#v", got)
	}
}
