package config

import "testing"

func TestExpandPipelineWithPmClear(t *testing.T) {
	steps := []string{StepPmClear, StepLogin, StepAutoPost, StepLogout}
	actions, err := ExpandPipelineSteps(steps, DefaultPipelineParams(steps))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"facebook_login", "facebook_auto_post", "facebook_logout", "force_stop_app"}
	if len(actions) != len(want) {
		t.Fatalf("got %d actions want %d: %v", len(actions), len(want), actionTypes(actions))
	}
	for i, w := range want {
		if actions[i].Type != w {
			t.Fatalf("action[%d]=%q want %q", i, actions[i].Type, w)
		}
	}
}

func TestPipelineIncludesPmClear(t *testing.T) {
	params := DefaultPipelineParams([]string{StepPmClear, StepLogin})
	if !PipelineIncludesPmClear("facebook_pipeline", params) {
		t.Fatal("expected pm clear in pipeline")
	}
	params = DefaultPipelineParams([]string{StepLogin, StepLogout})
	if PipelineIncludesPmClear("facebook_pipeline", params) {
		t.Fatal("pm clear should be off")
	}
}

func TestNormalizePipelinePmClearOnlyRejected(t *testing.T) {
	_, err := NormalizePipelineSteps([]string{StepPmClear})
	if err == nil {
		t.Fatal("expected error for pm_clear-only pipeline")
	}
}

func TestExpandPipelineNurhayati(t *testing.T) {
	steps := []string{StepLogin, StepAutoPost, StepFanpagePost, StepLogout}
	params := DefaultPipelineParams(steps)
	actions, err := ExpandPipelineSteps(steps, params)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"facebook_login", "facebook_auto_post", "facebook_fanpage_post", "facebook_logout", "force_stop_app"}
	if len(actions) != len(want) {
		t.Fatalf("got %d actions want %d: %v", len(actions), len(want), actionTypes(actions))
	}
	for i, w := range want {
		if actions[i].Type != w {
			t.Fatalf("action[%d]=%q want %q (all %v)", i, actions[i].Type, w, actionTypes(actions))
		}
	}
}

func TestExpandPipelineSherlly(t *testing.T) {
	steps := []string{StepLogin, StepFanpagePost, StepLogout}
	actions, err := ExpandPipelineSteps(steps, DefaultPipelineParams(steps))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"facebook_login", "facebook_fanpage_post", "facebook_logout", "force_stop_app"}
	for i, w := range want {
		if actions[i].Type != w {
			t.Fatalf("action[%d]=%q want %q", i, actions[i].Type, w)
		}
	}
}

func TestPresetFlowToStepsLegacy(t *testing.T) {
	steps := PresetFlowToSteps("facebook_login_fanpage_post_logout")
	want := []string{StepLogin, StepFanpagePost, StepLogout}
	for i, w := range want {
		if steps[i] != w {
			t.Fatalf("steps[%d]=%q want %q", i, steps[i], w)
		}
	}
}

func TestActionsForAccountUsesParamsSteps(t *testing.T) {
	params := DefaultPipelineParams([]string{StepLogin, StepAutoPost, StepLogout})
	actions, err := ActionsForAccount("facebook_pipeline", params)
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != 4 {
		t.Fatalf("expected 4 actions, got %d", len(actions))
	}
}

func actionTypes(actions []Action) []string {
	out := make([]string, len(actions))
	for i, a := range actions {
		out[i] = a.Type
	}
	return out
}
