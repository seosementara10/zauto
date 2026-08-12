package store

import (
	"context"
	"fmt"

	"zauto/internal/config"
)

// AccountSummary is a panel-friendly account row with optional device assignment.
type AccountSummary struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	LoginID            string `json:"login_id"`
	AssignedSerial     string `json:"assigned_serial,omitempty"`
	SlotNo             int    `json:"slot_no,omitempty"`
	AutomationFlow     string                 `json:"automation_flow"`
	AutomationParams   map[string]interface{} `json:"automation_params,omitempty"`
	AutomationSteps    []string               `json:"automation_steps,omitempty"`
	AutomationEnabled  bool                   `json:"automation_enabled"`
	FanpageCount       int                    `json:"fanpage_count"`
	Fanpages           []Fanpage              `json:"fanpages,omitempty"`
}

func (s *Store) ListAccountSummaries(ctx context.Context) ([]AccountSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT fa.id, fa.name, fa.email, fa.profile_id,
			COALESCE(d.serial, ''), COALESCE(das.slot_no, 0),
			fa.automation_flow, fa.automation_params, fa.automation_enabled,
			(SELECT COUNT(*) FROM fanpages fp WHERE fp.account_id = fa.id AND fp.status = 'active')
		FROM facebook_accounts fa
		LEFT JOIN device_account_slots das
			ON das.account_id = fa.id AND das.active = true
		LEFT JOIN devices d
			ON d.id = das.device_id AND d.status = 'active'
		WHERE fa.status = 'active'
		ORDER BY fa.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AccountSummary
	for rows.Next() {
		var a AccountSummary
		var email, profileID, serial string
		var slotNo, fanpageCount int
		var flow string
		var paramsRaw []byte
		var enabled bool
		if err := rows.Scan(&a.ID, &a.Name, &email, &profileID, &serial, &slotNo,
			&flow, &paramsRaw, &enabled, &fanpageCount); err != nil {
			return nil, err
		}
		acc := Account{Email: email, ProfileID: profileID}
		a.LoginID = acc.LoginID()
		a.AutomationFlow = flow
		a.AutomationParams = scanAutomationParams(paramsRaw)
		a.AutomationSteps = config.AccountPipelineSteps(flow, a.AutomationParams)
		a.AutomationEnabled = enabled
		a.FanpageCount = fanpageCount
		if serial != "" {
			a.AssignedSerial = serial
			a.SlotNo = slotNo
		}
		out = append(out, a)
	}
	if len(out) == 0 {
		return out, rows.Err()
	}
	ids := make([]int64, len(out))
	for i := range out {
		ids[i] = out[i].ID
	}
	fpMap, err := s.FanpagesByAccountIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range out {
		if fps, ok := fpMap[out[i].ID]; ok {
			out[i].Fanpages = fps
			out[i].FanpageCount = len(fps)
		}
	}
	return out, rows.Err()
}

func (s *Store) UnassignDevice(ctx context.Context, serial string, slotNo int) error {
	if slotNo <= 0 {
		slotNo = 1
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE device_account_slots das
		SET active = false
		FROM devices d
		WHERE das.device_id = d.id
			AND d.serial = $1
			AND das.slot_no = $2
			AND das.active = true`, serial, slotNo)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("tidak ada assignment untuk HP %s slot %d", serial, slotNo)
	}
	return nil
}

func (s *Store) UnassignAccount(ctx context.Context, accountID int64) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE device_account_slots SET active = false
		WHERE account_id = $1 AND active = true`, accountID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("akun %d belum di-assign", accountID)
	}
	return nil
}
