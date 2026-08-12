package config

import "testing"

func TestActionTextsFromGoSlice(t *testing.T) {
	a := Action{
		Extra: map[string]interface{}{
			"texts": []string{"Login", "Masuk"},
		},
	}
	got := ActionTexts(a)
	if len(got) != 2 || got[0] != "Login" {
		t.Fatalf("ActionTexts() = %v", got)
	}
}

func TestFacebookLoginFlowActions(t *testing.T) {
	actions, err := ExpandFlow("facebook_login_logout", nil)
	if err != nil {
		t.Fatal(err)
	}
	if actions[0].Type != "facebook_login" {
		t.Fatalf("action[0]=%q want facebook_login", actions[0].Type)
	}
}
