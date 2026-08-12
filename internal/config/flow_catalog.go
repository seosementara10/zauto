package config

// FlowInfo describes what a named flow does (shown in panel setup).
type FlowInfo struct {
	Steps       []string
	Description string
}

// FlowInfoFor returns human-readable pipeline for a flow name.
func FlowInfoFor(flow string) FlowInfo {
	switch flow {
	case "facebook_pipeline":
		return FlowInfo{
			Steps:       []string{"Pipeline kustom"},
			Description: "Gabungkan Login, Post Beranda, Post Fanpage, Logout — diatur per akun",
		}
	case "facebook_login":
		return FlowInfo{
			Steps:       []string{"Login"},
			Description: "Login saja — akun tetap masuk setelah selesai",
		}
	case "facebook_logout":
		return FlowInfo{
			Steps:       []string{"Logout", "Force stop"},
			Description: "Logout dari menu (harus sudah login)",
		}
	case "facebook_login_logout":
		return FlowInfo{
			Steps:       []string{"Login", "Logout", "Force stop"},
			Description: "Uji penuh: login → logout → tutup app",
		}
	case "facebook_auto_post":
		return FlowInfo{
			Steps:       []string{"Post Beranda personal"},
			Description: "Post ke feed personal (harus sudah login)",
		}
	case "facebook_login_auto_post":
		return FlowInfo{
			Steps:       []string{"Login", "Post Beranda personal"},
			Description: "Login → posting teks/gambar ke Beranda personal (tanpa logout)",
		}
	case "facebook_login_auto_post_logout":
		return FlowInfo{
			Steps:       []string{"Login", "Post Beranda personal", "Logout", "Force stop"},
			Description: "Login → post personal → logout → tutup app",
		}
	case "facebook_fanpage_post":
		return FlowInfo{
			Steps:       []string{"Post Fanpage"},
			Description: "Post ke fanpage (harus sudah login + fanpage di DB)",
		}
	case "facebook_login_fanpage_post":
		return FlowInfo{
			Steps:       []string{"Login", "Post Fanpage"},
			Description: "Login → post ke fanpage dari database (batch jika fanpage_mode=all)",
		}
	case "facebook_login_fanpage_post_logout":
		return FlowInfo{
			Steps:       []string{"Login", "Post Fanpage", "Logout", "Force stop"},
			Description: "Login → post fanpage → logout → tutup app",
		}
	default:
		return FlowInfo{
			Steps:       []string{flow},
			Description: "Custom flow",
		}
	}
}

// ActionLabels returns short labels for expanded actions (fallback when flow unknown).
func ActionLabels(actions []Action) []string {
	out := make([]string, 0, len(actions))
	for _, a := range actions {
		switch a.Type {
		case "facebook_login":
			out = append(out, "Login")
		case "facebook_logout":
			out = append(out, "Logout")
		case "facebook_auto_post":
			out = append(out, "Post Beranda")
		case "facebook_fanpage_post":
			out = append(out, "Post Fanpage")
		case "force_stop_app":
			out = append(out, "Force stop")
		default:
			out = append(out, a.Type)
		}
	}
	return out
}

// FlowOption is a selectable per-account automation preset in the panel.
type FlowOption struct {
	Flow   string
	Label  string
	Params map[string]interface{}
}

// AvailableFlows lists flows selectable per account in the panel.
func AvailableFlows() []FlowOption {
	return []FlowOption{
		{Flow: "facebook_login_logout", Label: "Login + Logout", Params: map[string]interface{}{}},
		{Flow: "facebook_login_auto_post", Label: "Login + Post Beranda", Params: map[string]interface{}{"post_index": float64(0)}},
		{Flow: "facebook_login_auto_post_logout", Label: "Login + Post Beranda + Logout", Params: map[string]interface{}{"post_index": float64(0)}},
		{Flow: "facebook_login_fanpage_post", Label: "Login + Post Fanpage (semua)", Params: map[string]interface{}{"post_index": float64(0), "fanpage_mode": "all"}},
		{Flow: "facebook_login_fanpage_post_logout", Label: "Login + Post Fanpage + Logout", Params: map[string]interface{}{"post_index": float64(0), "fanpage_mode": "all"}},
	}
}
