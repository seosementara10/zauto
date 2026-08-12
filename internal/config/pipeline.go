package config

import "fmt"

// Pipeline step IDs (fixed execution order).
const (
	StepPmClear     = "pm_clear"
	StepLogin       = "login"
	StepAutoPost    = "auto_post"
	StepFanpagePost = "fanpage_post"
	StepLogout      = "logout"
)

// AllPipelineSteps is the canonical order shown in the panel.
var AllPipelineSteps = []string{StepPmClear, StepLogin, StepAutoPost, StepFanpagePost, StepLogout}

// PipelineStepLabel returns a short UI label for a step id.
func PipelineStepLabel(step string) string {
	switch step {
	case StepPmClear:
		return "PM Clear"
	case StepLogin:
		return "Login"
	case StepAutoPost:
		return "Post Beranda"
	case StepFanpagePost:
		return "Post Fanpage"
	case StepLogout:
		return "Logout"
	default:
		return step
	}
}

// PipelineStepLabels returns labels for step ids.
func PipelineStepLabels(steps []string) []string {
	out := make([]string, 0, len(steps))
	for _, s := range steps {
		out = append(out, PipelineStepLabel(s))
	}
	return out
}

// PresetFlowToSteps maps legacy single-flow presets to composable steps.
func PresetFlowToSteps(flow string) []string {
	switch flow {
	case "facebook_login":
		return []string{StepLogin}
	case "facebook_logout":
		return []string{StepLogout}
	case "facebook_login_logout":
		return []string{StepLogin, StepLogout}
	case "facebook_auto_post":
		return []string{StepAutoPost}
	case "facebook_login_auto_post":
		return []string{StepLogin, StepAutoPost}
	case "facebook_login_auto_post_logout":
		return []string{StepLogin, StepAutoPost, StepLogout}
	case "facebook_fanpage_post":
		return []string{StepFanpagePost}
	case "facebook_login_fanpage_post":
		return []string{StepLogin, StepFanpagePost}
	case "facebook_login_fanpage_post_logout":
		return []string{StepLogin, StepFanpagePost, StepLogout}
	default:
		return nil
	}
}

func parseStepsParam(raw interface{}) []string {
	switch v := raw.(type) {
	case []string:
		return filterKnownSteps(v)
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return filterKnownSteps(out)
	default:
		return nil
	}
}

func filterKnownSteps(in []string) []string {
	allowed := map[string]struct{}{
		StepPmClear: {}, StepLogin: {}, StepAutoPost: {}, StepFanpagePost: {}, StepLogout: {},
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, step := range AllPipelineSteps {
		for _, s := range in {
			if s != step {
				continue
			}
			if _, ok := allowed[s]; !ok {
				continue
			}
			if _, dup := seen[s]; dup {
				continue
			}
			seen[s] = struct{}{}
			out = append(out, s)
			break
		}
	}
	return out
}

// AccountPipelineSteps resolves enabled steps for an account (params override legacy flow).
func AccountPipelineSteps(flow string, params map[string]interface{}) []string {
	if params != nil {
		if steps := parseStepsParam(params["steps"]); len(steps) > 0 {
			return steps
		}
	}
	if flow == "facebook_pipeline" {
		return []string{StepLogin, StepLogout}
	}
	if steps := PresetFlowToSteps(flow); len(steps) > 0 {
		return steps
	}
	return []string{StepLogin, StepLogout}
}

// DefaultPipelineParams returns shared post/fanpage params for a pipeline.
func DefaultPipelineParams(steps []string) map[string]interface{} {
	params := map[string]interface{}{
		"steps":        steps,
		"post_index":   float64(0),
		"fanpage_mode": "all",
		"post_source":  "db",
	}
	return params
}

// PipelineIncludesPmClear reports whether the account pipeline enables pm clear before open.
func PipelineIncludesPmClear(flow string, params map[string]interface{}) bool {
	for _, step := range AccountPipelineSteps(flow, params) {
		if step == StepPmClear {
			return true
		}
	}
	return false
}

// ExpandPipelineSteps builds actions for the given steps in canonical order.
// pm_clear is handled at app prepare time and does not produce actions.
func ExpandPipelineSteps(steps []string, params map[string]interface{}) ([]Action, error) {
	runnable := runnablePipelineSteps(steps)
	if len(runnable) == 0 {
		return nil, fmt.Errorf("pipeline: at least one runnable step required (login, post, or logout)")
	}
	var actions []Action
	for _, step := range runnable {
		switch step {
		case StepPmClear:
			continue
		case StepLogin:
			actions = append(actions, FacebookLoginFlow(params)...)
		case StepAutoPost:
			actions = append(actions, FacebookAutoPostFlow(params)...)
		case StepFanpagePost:
			actions = append(actions, FacebookFanpagePostFlow(params)...)
		case StepLogout:
			actions = append(actions, FacebookLogoutFlow(params)...)
		default:
			return nil, fmt.Errorf("pipeline: unknown step %q", step)
		}
	}
	return actions, nil
}

// ActionsForAccount expands runnable actions from flow + params (pipeline or legacy preset).
func ActionsForAccount(flow string, params map[string]interface{}) ([]Action, error) {
	steps := AccountPipelineSteps(flow, params)
	if len(steps) > 0 && (flow == "facebook_pipeline" || ParamsHasSteps(params) || PresetFlowToSteps(flow) != nil) {
		return ExpandPipelineSteps(steps, params)
	}
	return ExpandFlow(flow, params)
}

// NormalizePipelineSteps validates and orders step ids from the panel.
func NormalizePipelineSteps(steps []string) ([]string, error) {
	out := filterKnownSteps(steps)
	if len(out) == 0 {
		return nil, fmt.Errorf("pipeline: at least one step required")
	}
	if len(runnablePipelineSteps(out)) == 0 {
		return nil, fmt.Errorf("pipeline: at least one runnable step required (login, post, or logout)")
	}
	return out, nil
}

// ParamsHasSteps reports whether automation params include a custom pipeline step list.
func ParamsHasSteps(params map[string]interface{}) bool {
	if params == nil {
		return false
	}
	return len(parseStepsParam(params["steps"])) > 0
}

func runnablePipelineSteps(steps []string) []string {
	out := make([]string, 0, len(steps))
	for _, step := range filterKnownSteps(steps) {
		if step == StepPmClear {
			continue
		}
		out = append(out, step)
	}
	return out
}
