package config

import "testing"

func TestFacebookLogoutFlow(t *testing.T) {
	actions, err := ExpandFlow("facebook_logout", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}
	want := []string{"facebook_logout", "force_stop_app"}
	for i, w := range want {
		if actions[i].Type != w {
			t.Fatalf("action[%d]=%q want %q", i, actions[i].Type, w)
		}
	}
}

func TestFacebookLoginLogoutFlow(t *testing.T) {
	actions, err := ExpandFlow("facebook_login_logout", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) < 3 {
		t.Fatalf("expected at least 3 actions, got %d", len(actions))
	}
	types := make([]string, len(actions))
	for i, a := range actions {
		types[i] = a.Type
	}
	want := []string{"facebook_login", "facebook_logout", "force_stop_app"}
	for i, w := range want {
		if types[i] != w {
			t.Fatalf("action[%d]=%q want %q (all: %v)", i, types[i], w, types)
		}
	}
}

func TestLoadConfigExpandsLoginLogoutFlow(t *testing.T) {
	wf, err := Load("../../config/config.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(wf.Tasks) == 0 {
		t.Fatal("no tasks")
	}
	actions := wf.Tasks[0].Actions
	if len(actions) < 3 {
		t.Fatalf("expected expanded actions, got %d", len(actions))
	}
	last := actions[len(actions)-1]
	if last.Type != "force_stop_app" {
		t.Fatalf("last action=%q want force_stop_app", last.Type)
	}
	if !wf.ForceStopAfterTask {
		t.Fatal("ForceStopAfterTask should default true")
	}
}

func TestFacebookLoginAutoPostFlow(t *testing.T) {
	actions, err := ExpandFlow("facebook_login_auto_post", map[string]interface{}{"post_index": float64(0)})
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}
	want := []string{"facebook_login", "facebook_auto_post"}
	for i, w := range want {
		if actions[i].Type != w {
			t.Fatalf("action[%d]=%q want %q", i, actions[i].Type, w)
		}
	}
}

func TestFacebookLoginFanpagePostFlow(t *testing.T) {
	actions, err := ExpandFlow("facebook_login_fanpage_post", map[string]interface{}{
		"post_index": float64(0), "fanpage_mode": "all",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}
	if actions[0].Type != "facebook_login" || actions[1].Type != "facebook_fanpage_post" {
		t.Fatalf("types=%q %q", actions[0].Type, actions[1].Type)
	}
	if actions[1].ParamString("fanpage_mode", "") != "all" {
		t.Fatalf("fanpage_mode=%q", actions[1].ParamString("fanpage_mode", ""))
	}
}
