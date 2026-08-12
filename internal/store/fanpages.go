package store

import (
	"context"
	"fmt"
	"strings"
)

// FanpagesByAccountIDs returns active fanpages grouped by account id.
func (s *Store) FanpagesByAccountIDs(ctx context.Context, accountIDs []int64) (map[int64][]Fanpage, error) {
	out := map[int64][]Fanpage{}
	if len(accountIDs) == 0 {
		return out, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, account_id, fb_page_id, name
		FROM fanpages
		WHERE account_id = ANY($1) AND status = 'active'
		ORDER BY account_id, id`, accountIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var fp Fanpage
		var accountID int64
		if err := rows.Scan(&fp.ID, &accountID, &fp.FBPageID, &fp.Name); err != nil {
			return nil, err
		}
		out[accountID] = append(out[accountID], fp)
	}
	return out, rows.Err()
}

func parseFanpageToken(raw string) (name, pageID string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	if idx := strings.Index(raw, ":"); idx > 0 {
		left := strings.TrimSpace(raw[:idx])
		right := strings.TrimSpace(raw[idx+1:])
		if left != "" && isNumericCol(right) {
			return left, right
		}
	}
	return raw, raw
}

func (s *Store) upsertFanpageToken(ctx context.Context, accountID int64, token string) error {
	name, pageID := parseFanpageToken(token)
	if pageID == "" {
		return nil
	}
	if name == "" {
		name = pageID
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO fanpages (account_id, fb_page_id, name)
		VALUES ($1, $2, $3)
		ON CONFLICT (account_id, fb_page_id)
		DO UPDATE SET name = EXCLUDED.name, status = 'active'`,
		accountID, pageID, name)
	return err
}

// SyncFanpagesForLoginID upserts fanpages for an account matched by profile_id or email.
func (s *Store) SyncFanpagesForLoginID(ctx context.Context, loginID string, tokens []string) error {
	loginID = strings.TrimSpace(loginID)
	if loginID == "" {
		return fmt.Errorf("login id required for fanpage sync")
	}
	email, profileID := loginID, loginID
	if strings.Contains(loginID, "@") {
		profileID = ""
	} else {
		email = ""
	}
	accountID, err := s.findAccountID(ctx, email, profileID)
	if err != nil {
		return err
	}
	if accountID == 0 {
		return fmt.Errorf("account not found for login id %q", loginID)
	}
	for _, tok := range tokens {
		if err := s.upsertFanpageToken(ctx, accountID, tok); err != nil {
			return err
		}
	}
	return nil
}
