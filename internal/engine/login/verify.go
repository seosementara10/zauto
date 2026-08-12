package login

import (
	"fmt"
	"time"

	"zauto/internal/engine/runtime"
	"zauto/internal/state"
	"zauto/internal/ui"
)

func VerifyFieldsFilled(e runtime.Exec, emailHints, emailRIDs, passHints, passRIDs []string, email string) error {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		snap := e.Sess().ReadUI(true)
		if err := verifyFieldsFilledOnce(e, snap, emailHints, emailRIDs, passHints, passRIDs, email); err == nil {
			return nil
		}
		time.Sleep(e.Sess().PollInterval())
	}
	snap := e.Sess().ReadUI(true)
	return verifyFieldsFilledOnce(e, snap, emailHints, emailRIDs, passHints, passRIDs, email)
}

func verifyFieldsFilledOnce(e runtime.Exec, snap ui.Snapshot, emailHints, emailRIDs, passHints, passRIDs []string, email string) error {
	resolver := e.Sess().Resolver

	emailField := ui.FindLoginEmailField(resolver, snap, emailHints, emailRIDs)
	if emailField == nil {
		return fmt.Errorf("email field not found after fill")
	}
	if !ui.FieldHasExpectedValue(emailField.Element, email) {
		got := emailField.Element.Text
		if got == "" {
			got = emailField.Element.ContentDesc
		}
		return fmt.Errorf("email field value mismatch (got %q)", got)
	}

	passField := ui.FindLoginPasswordField(resolver, snap, passHints, passRIDs)
	if passField == nil {
		return fmt.Errorf("password field not found after fill")
	}
	typed, _ := e.Sess().Runtime[runtimePasswordTypedKey].(bool)
	if !typed {
		return fmt.Errorf("password not typed")
	}
	pkg := e.Sess().Client.ForegroundPackage()
	if state.IMEBlocksInput(snap, pkg) {
		return fmt.Errorf("keyboard settings overlay blocking login")
	}
	e.Event("VERIFY login fields filled email=%q", emailField.Element.Text)
	return nil
}
