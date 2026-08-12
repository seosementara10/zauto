package config

func FacebookLiteRegisterFlow(params map[string]interface{}) []Action {
	nameIdx, emailIdx := 0, 0
	if params != nil {
		if v, ok := params["name_index"].(float64); ok {
			nameIdx = int(v)
		}
		if v, ok := params["email_index"].(float64); ok {
			emailIdx = int(v)
		}
	}
	return []Action{
		{Type: "tap_create_account", Extra: map[string]interface{}{"type": "tap_create_account", "timeout_sec": 20.0}},
		{Type: "handle_permission", Optional: true, Extra: map[string]interface{}{"type": "handle_permission", "policy": "deny", "optional": true}},
		{Type: "tap_text", Optional: true, Texts: []string{"SKIP", "Skip", "Lewati"}, Extra: map[string]interface{}{"type": "tap_text", "optional": true, "texts": []string{"SKIP", "Skip", "Lewati"}, "timeout_sec": 0.5, "fast_poll": true}},
		{Type: "fill_name", Extra: map[string]interface{}{"type": "fill_name", "name_index": nameIdx, "skip_verify": true, "form_timeout_sec": 5.0}},
		{Type: "set_birthday", Extra: map[string]interface{}{"type": "set_birthday", "min_age": 21, "max_age": 49, "month_swipes": 2, "day_swipes": 2, "skip_verify": true}},
		{Type: "set_gender", Extra: map[string]interface{}{"type": "set_gender", "skip_verify": true}},
		{Type: "tap_text", Texts: []string{"Sign up with email", "Sign Up with Email", "Daftar dengan email", "Daftar dengan e-mail"}, Extra: map[string]interface{}{
			"type": "tap_text", "texts": []string{"Sign up with email", "Sign Up with Email", "Daftar dengan email", "Daftar dengan e-mail"},
			"timeout_sec": 8.0, "prefer": "bottom",
			"verify": map[string]interface{}{"texts": []string{"What's your email?", "Sign up with mobile number", "Enter the email where you can be contacted"}, "timeout_sec": 10.0},
		}},
		{Type: "fill_email", Extra: map[string]interface{}{"type": "fill_email", "email_index": emailIdx, "skip_verify": true, "form_timeout_sec": 8.0}},
	}
}

func FacebookLoginFlow(params map[string]interface{}) []Action {
	_ = params
	return []Action{
		{Type: "facebook_login", Extra: map[string]interface{}{
			"type": "facebook_login",
		}},
	}
}

// FacebookLogoutFlow logs out via menu when the device is already logged in.
func FacebookLogoutFlow(params map[string]interface{}) []Action {
	_ = params
	return []Action{
		{Type: "facebook_logout", Extra: map[string]interface{}{
			"type": "facebook_logout", "settle_sec": 2.0, "verify_timeout_sec": 20.0,
		}},
		{Type: "force_stop_app", Extra: map[string]interface{}{
			"type": "force_stop_app", "app": "facebook",
		}},
	}
}

// FacebookLoginLogoutFlow opens FB, logs in, logs out via menu, then force-stops the app.
func FacebookLoginLogoutFlow(params map[string]interface{}) []Action {
	actions := FacebookLoginFlow(params)
	actions = append(actions, FacebookLogoutFlow(params)...)
	return actions
}

func autoPostExtra(params map[string]interface{}) map[string]interface{} {
	extra := map[string]interface{}{
		"type":                 "facebook_auto_post",
		"composer_timeout_sec": 15.0,
		"verify_timeout_sec":   45.0,
	}
	if params == nil {
		return extra
	}
	for _, k := range []string{"posts_file", "images_dir", "post_text", "image_path", "post_index", "post_source", "post_category"} {
		if v, ok := params[k]; ok {
			extra[k] = v
		}
	}
	return extra
}

func fanpagePostExtra(params map[string]interface{}) map[string]interface{} {
	extra := map[string]interface{}{
		"type":                 "facebook_fanpage_post",
		"composer_timeout_sec": 15.0,
		"verify_timeout_sec":   45.0,
		"switch_timeout_sec":   20.0,
		"fanpage_mode":         "single",
	}
	if params == nil {
		return extra
	}
	for _, k := range []string{
		"posts_file", "images_dir", "post_text", "image_path", "post_index",
		"post_source", "post_category",
		"fanpage_id", "fanpage_index", "fanpage_mode",
	} {
		if v, ok := params[k]; ok {
			extra[k] = v
		}
	}
	return extra
}

// FacebookAutoPostFlow posts to personal feed — device must already be logged in.
func FacebookAutoPostFlow(params map[string]interface{}) []Action {
	return []Action{{Type: "facebook_auto_post", Extra: autoPostExtra(params)}}
}

// FacebookLoginAutoPostFlow logs in then posts text/image to Beranda (personal account).
func FacebookLoginAutoPostFlow(params map[string]interface{}) []Action {
	actions := FacebookLoginFlow(params)
	actions = append(actions, FacebookAutoPostFlow(params)...)
	return actions
}

// FacebookFanpagePostFlow posts to fanpage(s) — device must already be logged in.
func FacebookFanpagePostFlow(params map[string]interface{}) []Action {
	return []Action{{Type: "facebook_fanpage_post", Extra: fanpagePostExtra(params)}}
}

// FacebookLoginFanpagePostFlow logs in then posts to fanpage(s) from DB.
func FacebookLoginFanpagePostFlow(params map[string]interface{}) []Action {
	actions := FacebookLoginFlow(params)
	actions = append(actions, FacebookFanpagePostFlow(params)...)
	return actions
}

// FacebookLoginAutoPostLogoutFlow logs in, posts to personal feed, then logs out.
func FacebookLoginAutoPostLogoutFlow(params map[string]interface{}) []Action {
	actions := FacebookLoginAutoPostFlow(params)
	actions = append(actions, FacebookLogoutFlow(params)...)
	return actions
}

// FacebookLoginFanpagePostLogoutFlow logs in, posts to fanpage(s), then logs out.
func FacebookLoginFanpagePostLogoutFlow(params map[string]interface{}) []Action {
	actions := FacebookLoginFanpagePostFlow(params)
	actions = append(actions, FacebookLogoutFlow(params)...)
	return actions
}
