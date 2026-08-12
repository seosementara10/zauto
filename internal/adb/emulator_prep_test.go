package adb

import "testing"

func TestLDPlayerADBPorts(t *testing.T) {
	ports := LDPlayerADBPorts(10)
	if len(ports) != 10 || ports[0] != 5555 || ports[9] != 5573 {
		t.Fatalf("LDPlayerADBPorts(10) = %v", ports)
	}
}

func TestIsEmulatorSerial(t *testing.T) {
	cases := map[string]bool{
		"127.0.0.1:5555":  true,
		"localhost:5555":  true,
		"emulator-5554":   true,
		"007519090c2dc80d": false,
		"":                false,
	}
	for serial, want := range cases {
		if got := IsEmulatorSerial(serial); got != want {
			t.Fatalf("IsEmulatorSerial(%q) = %v, want %v", serial, got, want)
		}
	}
}
