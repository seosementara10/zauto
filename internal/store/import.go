package store

import (
	"context"
	"fmt"
	"os"
	"strings"
)

type importRow struct {
	Name         string
	Email        string
	Password     string
	ProfileID    string
	Fanpages     []string
	FanpagesOnly bool
	LoginKey     string
}

// ImportAccountsFile loads accounts + fanpages from a legacy pipe/TSV file into PostgreSQL.
func (s *Store) ImportAccountsFile(ctx context.Context, path string) (int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var rows []importRow
	for i, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		row, ok, err := parseImportLine(line, path, i+1)
		if err != nil {
			return 0, err
		}
		if ok {
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		return 0, fmt.Errorf("no accounts in %s", path)
	}
	imported := 0
	for _, row := range rows {
		if err := s.upsertImportRow(ctx, row); err != nil {
			return imported, err
		}
		imported++
	}
	return imported, nil
}

func (s *Store) upsertImportRow(ctx context.Context, row importRow) error {
	if row.FanpagesOnly {
		return s.SyncFanpagesForLoginID(ctx, row.LoginKey, row.Fanpages)
	}
	accountID, err := s.findAccountID(ctx, row.Email, row.ProfileID)
	if err != nil {
		return err
	}
	if accountID == 0 {
		err = s.pool.QueryRow(ctx, `
			INSERT INTO facebook_accounts (name, email, password, profile_id)
			VALUES ($1, $2, $3, $4) RETURNING id`,
			row.Name, row.Email, row.Password, row.ProfileID).Scan(&accountID)
		if err != nil {
			return err
		}
	} else {
		_, err = s.pool.Exec(ctx, `
			UPDATE facebook_accounts SET name = $1, email = $2, password = $3, profile_id = $4
			WHERE id = $5`, row.Name, row.Email, row.Password, row.ProfileID, accountID)
		if err != nil {
			return err
		}
	}
	for _, fp := range row.Fanpages {
		if err := s.upsertFanpageToken(ctx, accountID, fp); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) findAccountID(ctx context.Context, email, profileID string) (int64, error) {
	var id int64
	if profileID != "" {
		err := s.pool.QueryRow(ctx, `
			SELECT id FROM facebook_accounts WHERE profile_id = $1 LIMIT 1`, profileID).Scan(&id)
		if err == nil {
			return id, nil
		}
	}
	if email != "" {
		err := s.pool.QueryRow(ctx, `
			SELECT id FROM facebook_accounts WHERE email = $1 LIMIT 1`, email).Scan(&id)
		if err == nil {
			return id, nil
		}
	}
	return 0, nil
}

func parseImportLine(line, path string, lineNo int) (importRow, bool, error) {
	if strings.HasPrefix(line, "@fanpages|") {
		parts := strings.Split(line, "|")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		if len(parts) < 3 {
			return importRow{}, false, fmt.Errorf("%s:%d @fanpages needs login id and at least one fanpage", path, lineNo)
		}
		var fps []string
		for _, fp := range parts[2:] {
			if fp = strings.TrimSpace(fp); fp != "" {
				fps = append(fps, fp)
			}
		}
		if len(fps) == 0 {
			return importRow{}, false, fmt.Errorf("%s:%d @fanpages needs at least one fanpage", path, lineNo)
		}
		return importRow{FanpagesOnly: true, LoginKey: parts[1], Fanpages: fps}, true, nil
	}
	if strings.Contains(line, "\t") {
		return parseImportTSV(line)
	}
	parts := strings.Split(line, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	switch len(parts) {
	case 2:
		id, pass := parts[0], parts[1]
		if id == "" || pass == "" {
			return importRow{}, false, fmt.Errorf("%s:%d invalid account line", path, lineNo)
		}
		row := importRow{Password: pass}
		if strings.Contains(id, "@") {
			row.Email = id
		} else {
			row.ProfileID = id
		}
		return row, true, nil
	case 3:
		if strings.Contains(parts[0], "@") {
			return importRow{Email: parts[0], Password: parts[1], Name: parts[2]}, true, nil
		}
		return importRow{Name: parts[0], Password: parts[1], ProfileID: parts[2]}, true, nil
	default:
		if len(parts) < 4 {
			return importRow{}, false, fmt.Errorf("%s:%d invalid account line", path, lineNo)
		}
		row := importRow{
			Name: parts[0], Password: parts[1], Email: parts[2], ProfileID: parts[3],
		}
		for _, fp := range parts[4:] {
			if fp = strings.TrimSpace(fp); fp != "" {
				row.Fanpages = append(row.Fanpages, fp)
			}
		}
		if row.Password == "" || (row.ProfileID == "" && row.Email == "") {
			return importRow{}, false, fmt.Errorf("%s:%d missing login id or password", path, lineNo)
		}
		return row, true, nil
	}
}

func parseImportTSV(line string) (importRow, bool, error) {
	parts := strings.Split(line, "\t")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	if len(parts) == 0 {
		return importRow{}, false, nil
	}
	first := strings.ToLower(parts[0])
	if first == "no." || first == "no" || strings.EqualFold(parts[0], "Tgl") ||
		strings.Contains(strings.ToLower(line), "profile utama") {
		return importRow{}, false, nil
	}
	var row importRow
	switch {
	case len(parts) >= 6 && isNumericCol(parts[0]):
		row = importRow{Name: parts[2], Password: parts[3], Email: parts[4], ProfileID: parts[5]}
		for _, fp := range parts[6:] {
			if fp = strings.TrimSpace(fp); fp != "" {
				row.Fanpages = append(row.Fanpages, fp)
			}
		}
	case len(parts) >= 5:
		row = importRow{Name: parts[1], Password: parts[2], Email: parts[3], ProfileID: parts[4]}
		for _, fp := range parts[5:] {
			if fp = strings.TrimSpace(fp); fp != "" {
				row.Fanpages = append(row.Fanpages, fp)
			}
		}
	default:
		return importRow{}, false, fmt.Errorf("invalid TSV account line")
	}
	if row.Password == "" || (row.ProfileID == "" && row.Email == "") {
		return importRow{}, false, fmt.Errorf("account missing password or login id")
	}
	return row, true, nil
}

func isNumericCol(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
