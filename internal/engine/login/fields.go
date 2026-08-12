package login

import (
	"fmt"

	"zauto/internal/config"
	"zauto/internal/engine/overlay"
	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

func FillFields(e runtime.Exec, _ config.Action) error {
	st := e.Sess().Store
	if st == nil {
		return fmt.Errorf("database not configured — set database.url in config")
	}
	accountID, ok := e.Sess().Runtime["account_id"].(int64)
	if !ok || accountID <= 0 {
		return fmt.Errorf("no account bound to device %s — import accounts and assign slots (zauto --db-import / --db-auto-assign)", e.Sess().Serial)
	}
	account, err := st.GetAccountByID(e.Ctx(), accountID)
	if err != nil {
		return err
	}
	loginID := account.LoginID()
	if loginID == "" {
		return fmt.Errorf("account %d has no email or profile_id", accountID)
	}

	e.Sess().Runtime["login_id"] = loginID
	e.Event("ACT account=%s id=%d", loginID, accountID)

	ResetFillState(e)
	if err := EnsureFormReady(e); err != nil {
		return err
	}

	emailFields := append([]string{
		"Mobile number or email", "Email or phone number", "Email", "Email atau nomor telepon",
	}, state.LoginEmailFieldTexts...)
	emailRIDs := []string{"login_username", "login_email", "m_login_email"}
	passFields := append([]string(nil), state.LoginPasswordFieldTexts...)
	passRIDs := []string{"login_password", "password", "m_login_password"}
	loginBtns := append([]string(nil), state.LoginButtonTexts...)

	if err := e.FillLoginField(emailFields, emailRIDs, loginID, 0); err != nil {
		return err
	}
	overlay.DismissSoftKeyboard(e)
	overlay.DismissKeyboardSettingsIfPresent(e)
	if err := e.FillLoginField(passFields, passRIDs, account.Password, 1); err != nil {
		return err
	}
	overlay.DismissKeyboardSettingsIfPresent(e)
	if err := VerifyFieldsFilled(e, emailFields, emailRIDs, passFields, passRIDs, loginID); err != nil {
		return err
	}
	overlay.DismissSoftKeyboard(e)
	overlay.DismissKeyboardSettingsIfPresent(e)
	loginQueries := []ui.FindQuery{
		{Texts: loginBtns, ContentDescs: loginBtns, PreferClickable: true, MinCenterY: 1},
	}
	if err := e.TapFirstQuery(loginQueries, 12); err != nil {
		return fmt.Errorf("Log in button not found")
	}
	return nil
}
