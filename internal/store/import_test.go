package store

import "testing"

func TestParseImportLinePipe(t *testing.T) {
	row, ok, err := parseImportLine(
		"Nurhayati Guguk|password||61592118521974|Ibu Nurhayati:615931763399|Pengalaman Nurhayati:61593132631889",
		"accounts.txt", 1)
	if err != nil || !ok {
		t.Fatalf("parse: ok=%v err=%v", ok, err)
	}
	if row.Name != "Nurhayati Guguk" || row.Password != "password" || row.ProfileID != "61592118521974" {
		t.Fatalf("unexpected row: %+v", row)
	}
	if len(row.Fanpages) != 2 {
		t.Fatalf("fanpages=%v", row.Fanpages)
	}
}

func TestParseFanpageToken(t *testing.T) {
	name, id := parseFanpageToken("Ibu Nurhayati:615931763399")
	if name != "Ibu Nurhayati" || id != "615931763399" {
		t.Fatalf("got name=%q id=%q", name, id)
	}
	name, id = parseFanpageToken("615931763399")
	if name != "615931763399" || id != "615931763399" {
		t.Fatalf("legacy id: name=%q id=%q", name, id)
	}
}

func TestParseImportLineFanpagesOnly(t *testing.T) {
	row, ok, err := parseImportLine(
		"@fanpages|61592118521974|Ibu Nurhayati:615931763399|Fans Nurhayati:61592753657118",
		"fanpages.txt", 1)
	if err != nil || !ok {
		t.Fatalf("parse: ok=%v err=%v", ok, err)
	}
	if !row.FanpagesOnly || row.LoginKey != "61592118521974" || len(row.Fanpages) != 2 {
		t.Fatalf("unexpected row: %+v", row)
	}
}

func TestAccountLoginID(t *testing.T) {
	if got := (Account{ProfileID: "61592118521974"}).LoginID(); got != "61592118521974" {
		t.Fatalf("profile login id: %q", got)
	}
	if got := (Account{Email: "a@b.com"}).LoginID(); got != "a@b.com" {
		t.Fatalf("email login id: %q", got)
	}
}
