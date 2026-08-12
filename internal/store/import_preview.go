package store

import (
	"context"
	"fmt"
	"strings"
)

// ImportPreviewRow is a sanitized preview line for the import UI.
type ImportPreviewRow struct {
	LineNo       int    `json:"line_no"`
	Name         string `json:"name"`
	LoginID      string `json:"login_id"`
	FanpageCount int    `json:"fanpage_count"`
	Valid        bool   `json:"valid"`
	Error        string `json:"error,omitempty"`
}

func parseImportContent(content string) ([]importRow, []ImportPreviewRow, error) {
	var rows []importRow
	var preview []ImportPreviewRow
	for i, line := range strings.Split(content, "\n") {
		lineNo := i + 1
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		row, ok, err := parseImportLine(line, "preview", lineNo)
		p := ImportPreviewRow{LineNo: lineNo, Valid: ok && err == nil}
		if err != nil {
			p.Error = err.Error()
			preview = append(preview, p)
			continue
		}
		if !ok {
			continue
		}
		acc := Account{Email: row.Email, ProfileID: row.ProfileID}
		p.Name = row.Name
		p.LoginID = acc.LoginID()
		p.FanpageCount = len(row.Fanpages)
		preview = append(preview, p)
		rows = append(rows, row)
	}
	if len(rows) == 0 && len(preview) == 0 {
		return nil, nil, fmt.Errorf("no accounts found in preview")
	}
	return rows, preview, nil
}

// PreviewImportContent parses account file text without writing to the database.
func PreviewImportContent(content string) ([]ImportPreviewRow, error) {
	_, preview, err := parseImportContent(content)
	return preview, err
}

// ImportAccountsContent imports accounts from raw file text.
func (s *Store) ImportAccountsContent(ctx context.Context, content string) (int, error) {
	rows, _, err := parseImportContent(content)
	if err != nil {
		return 0, err
	}
	imported := 0
	for _, row := range rows {
		if err := s.upsertImportRow(ctx, row); err != nil {
			return imported, err
		}
		imported++
	}
	if imported == 0 {
		return 0, fmt.Errorf("no valid accounts imported")
	}
	return imported, nil
}
