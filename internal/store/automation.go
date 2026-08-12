package store

import (
	"context"
	"encoding/json"
	"fmt"
)

func scanAutomationParams(raw []byte) map[string]interface{} {
	if len(raw) == 0 {
		return map[string]interface{}{}
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil || m == nil {
		return map[string]interface{}{}
	}
	return m
}

func (s *Store) SetAccountAutomation(ctx context.Context, accountID int64, flow string, params map[string]interface{}, enabled bool) error {
	if accountID <= 0 {
		return fmt.Errorf("invalid account id")
	}
	if flow == "" {
		flow = "facebook_login_logout"
	}
	raw, err := json.Marshal(params)
	if err != nil {
		return err
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE facebook_accounts
		SET automation_flow = $2, automation_params = $3, automation_enabled = $4
		WHERE id = $1 AND status = 'active'`, accountID, flow, raw, enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("account %d not found", accountID)
	}
	return nil
}
