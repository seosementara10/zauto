package worker

import (
	"testing"

	"zauto/internal/config"
)

func TestShouldPmClearFromPipelineStep(t *testing.T) {
	wf := &config.Workflow{ClearAppBeforeOpen: false}
	task := config.Task{
		Flow:   "facebook_pipeline",
		Params: config.DefaultPipelineParams([]string{config.StepPmClear, config.StepLogin}),
	}
	if !shouldPmClear(wf, task) {
		t.Fatal("expected pm clear when pipeline step checked")
	}
}

func TestShouldPmClearOffWhenPipelineStepUnchecked(t *testing.T) {
	wf := &config.Workflow{ClearAppBeforeOpen: true}
	task := config.Task{
		Flow:   "facebook_pipeline",
		Params: config.DefaultPipelineParams([]string{config.StepLogin, config.StepLogout}),
	}
	if shouldPmClear(wf, task) {
		t.Fatal("global clear must not apply when account has custom pipeline steps")
	}
}

func TestShouldPmClearGlobalForLegacyTask(t *testing.T) {
	wf := &config.Workflow{ClearAppBeforeOpen: true}
	task := config.Task{Flow: "facebook_login_logout", Params: map[string]interface{}{}}
	if !shouldPmClear(wf, task) {
		t.Fatal("expected global clear for legacy task without pipeline params")
	}
}
