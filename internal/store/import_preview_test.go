package store

import "testing"

func TestPreviewImportContent(t *testing.T) {
	content := "# comment\n\nuser@mail.com|secret|Name One\nbad line only one field\n"
	rows, err := PreviewImportContent(content)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows=%d", len(rows))
	}
	if !rows[0].Valid || rows[0].LoginID != "user@mail.com" {
		t.Fatalf("row0=%+v", rows[0])
	}
	if rows[1].Valid || rows[1].Error == "" {
		t.Fatalf("row1 should be invalid: %+v", rows[1])
	}
}
