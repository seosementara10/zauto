package worker

import (
	"fmt"

	"zauto/internal/config"
	"zauto/internal/store"
)

// TasksForAccount builds runnable tasks from the account automation profile.
func TasksForAccount(acc store.Account) ([]config.Task, error) {
	if !acc.AutomationEnabled {
		return nil, fmt.Errorf("automation disabled for account %d (%s)", acc.ID, acc.Name)
	}
	flow := acc.AutomationFlow
	if flow == "" {
		flow = "facebook_login_logout"
	}
	params := acc.AutomationParams
	if params == nil {
		params = map[string]interface{}{}
	}
	actions, err := config.ActionsForAccount(flow, params)
	if err != nil {
		return nil, fmt.Errorf("account %d flow %q: %w", acc.ID, flow, err)
	}
	return []config.Task{{
		Name:    flow,
		App:     "facebook",
		Flow:    flow,
		Params:  params,
		Actions: actions,
	}}, nil
}
