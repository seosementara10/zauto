package config

import "testing"

func TestFlowInfoForLoginAutoPost(t *testing.T) {
	info := FlowInfoFor("facebook_login_auto_post")
	if len(info.Steps) != 2 {
		t.Fatalf("steps=%v", info.Steps)
	}
	if info.Steps[0] != "Login" || info.Steps[1] != "Post Beranda personal" {
		t.Fatalf("steps=%v", info.Steps)
	}
}

func TestFacebookLoginAutoPostLogoutFlow(t *testing.T) {
	actions, err := ExpandFlow("facebook_login_auto_post_logout", nil)
	if err != nil {
		t.Fatal(err)
	}
	labels := ActionLabels(actions)
	want := []string{"Login", "Post Beranda", "Logout", "Force stop"}
	if len(labels) != len(want) {
		t.Fatalf("labels=%v want %v", labels, want)
	}
	for i := range want {
		if labels[i] != want[i] {
			t.Fatalf("labels=%v", labels)
		}
	}
}
