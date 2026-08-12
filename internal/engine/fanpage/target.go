package fanpage

import (
	"context"
	"fmt"
	"strings"

	"zauto/internal/config"
	"zauto/internal/engine/runtime"
	"zauto/internal/store"
)

// ResolveFanpageTargets picks fanpage(s) from action params and the bound account in DB.
func ResolveFanpageTargets(e runtime.Exec, action config.Action) ([]store.Fanpage, error) {
	fps, err := AccountFanpages(e.Sess(), e.Ctx())
	if err != nil {
		return nil, err
	}
	return SelectFanpages(fps, action)
}

// SelectFanpages filters fanpages by action params (testable without full Exec).
func SelectFanpages(fps []store.Fanpage, action config.Action) ([]store.Fanpage, error) {
	if len(fps) == 0 {
		return nil, fmt.Errorf("account has no fanpages in database — import account with page IDs first")
	}

	mode := strings.ToLower(strings.TrimSpace(action.ParamString("fanpage_mode", "single")))
	if mode == "all" {
		return fps, nil
	}

	if id := strings.TrimSpace(action.ParamString("fanpage_id", "")); id != "" {
		for _, fp := range fps {
			if fp.FBPageID == id || fp.Name == id {
				return []store.Fanpage{fp}, nil
			}
		}
		return nil, fmt.Errorf("fanpage_id %q not found for this account (have %d pages)", id, len(fps))
	}

	idx := action.ParamInt("fanpage_index", 0)
	if idx < 0 || idx >= len(fps) {
		return nil, fmt.Errorf("fanpage_index %d out of range (account has %d fanpages)", idx, len(fps))
	}
	return []store.Fanpage{fps[idx]}, nil
}

// AccountFanpages returns fanpages bound to the session or loaded from store.
func AccountFanpages(sess *runtime.Session, ctx context.Context) ([]store.Fanpage, error) {
	if raw, ok := sess.Runtime["fanpages"]; ok {
		switch v := raw.(type) {
		case []store.Fanpage:
			return v, nil
		}
	}
	accountID, ok := sess.Runtime["account_id"].(int64)
	if !ok || accountID <= 0 || sess.Store == nil {
		return nil, fmt.Errorf("fanpages not bound — account_id missing from session")
	}
	acc, err := sess.Store.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	sess.Runtime["fanpages"] = acc.Fanpages
	return acc.Fanpages, nil
}

func setActiveFanpage(e runtime.Exec, fp store.Fanpage) {
	delete(e.Sess().Runtime, "fanpage_display_name")
	e.Sess().Runtime["fanpage_id"] = fp.ID
	e.Sess().Runtime["fanpage_fb_id"] = fp.FBPageID
	e.Sess().Runtime["fanpage_name"] = fp.Name
}
