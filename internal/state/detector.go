package state

import (
	"strconv"
	"time"

	"zauto/internal/ui"
)

type signal struct {
	label  string
	texts  []string
	points float64
}

type rule struct {
	state    UIState
	signals  []signal
	minScore float64
}

// Facebook-oriented rules — multi-evidence scoring, not single keyword.
var facebookRules = []rule{
	{
		state: UIPermission,
		signals: []signal{
			{label: "allow/deny", texts: append(PermissionAllowTexts, PermissionDenyTexts...), points: 40},
			{label: "notification", texts: NotificationPermissionPromptTexts, points: 35},
			{label: "contacts", texts: ContactPermissionPromptTexts, points: 35},
		},
		minScore: 70,
	},
	{
		state: UIOnboarding,
		signals: []signal{
			{label: "join", texts: []string{"Gabung Facebook", "Join Facebook"}, points: 40},
			{label: "start", texts: []string{"Mulai", "Get started", "Get Started"}, points: 30},
			{label: "have profile", texts: OnboardingHaveAccountTexts, points: 30},
		},
		minScore: 70,
	},
	{
		state: UILogin,
		signals: []signal{
			{label: "email field", texts: LoginEmailFieldTexts, points: 40},
			{label: "password field", texts: LoginPasswordFieldTexts, points: 30},
			{label: "login btn", texts: LoginButtonTexts, points: 20},
			{label: "forgot", texts: []string{"Lupa kata sandi", "Forgot password"}, points: 10},
		},
		minScore: 70,
	},
	{
		state: UI2FACheckpoint,
		signals: []signal{
			{label: "2fa title", texts: TwoFactorTitleTexts, points: 50},
			{label: "2fa action", texts: TwoFactorActionTexts, points: 30},
			{label: "code field", texts: []string{"login code", "kode login", "Enter code", "Masukkan kode"}, points: 20},
		},
		minScore: 70,
	},
	{
		state: UILoading,
		signals: []signal{
			{label: "loading text", texts: []string{"Loading", "Memuat", "Please wait"}, points: 60},
			{label: "progress", texts: []string{"ProgressBar", "progress_bar"}, points: 40},
		},
		minScore: 70,
	},
	{
		state: UIPasswordManagerSheet,
		signals: []signal{
			{label: "google pm", texts: PasswordManagerTitleTexts, points: 35},
			{label: "save pwd title", texts: PasswordManagerSaveTitleTexts, points: 35},
			{label: "never btn", texts: PasswordManagerNeverTexts, points: 25},
			{label: "continue btn", texts: PasswordManagerContinueTexts, points: 20},
		},
		minScore: 70,
	},
	{
		state: UIKeyboardSettings,
		signals: []signal{
			{label: "settings title", texts: KeyboardSettingsTitleTexts, points: 50},
			{label: "settings menu", texts: KeyboardSettingsMenuTexts, points: 30},
			{label: "settings nav", texts: KeyboardSettingsNavTexts, points: 20},
		},
		minScore: 70,
	},
	{
		state: UISavedProfileScreen,
		signals: []signal{
			{label: "continue profile", texts: SavedProfileContinueTexts, points: 35},
			{label: "other profile", texts: SavedProfileOtherTexts, points: 35},
			{label: "create account", texts: SavedProfileCreateTexts, points: 30},
		},
		minScore: 70,
	},
	{
		state: UIContactFollowPrompt,
		signals: []signal{
			{label: "contact title", texts: ContactFollowTitleTexts, points: 40},
			{label: "skip confirm", texts: ContactSkipConfirmTitleTexts, points: 40},
			{label: "skip", texts: ContactFollowSkipTexts, points: 20},
			{label: "continue/upload", texts: append(ContactFollowContinueTexts, ContactUploadTexts...), points: 20},
		},
		minScore: 70,
	},
	{
		state: UILocationServicesPrompt,
		signals: []signal{
			{label: "location intro", texts: LocationServicesIntroTexts, points: 45},
			{label: "skip", texts: PostLoginSkipTexts, points: 35},
			{label: "continue", texts: []string{"Lanjutkan", "Continue"}, points: 20},
		},
		minScore: 70,
	},
	{
		state: UILoginAccountFinderPrompt,
		signals: []signal{
			{label: "find account", texts: LoginAccountFinderTexts, points: 45},
			{label: "phone step", texts: LoginAccountFinderPromptTexts, points: 35},
			{label: "dismiss ok", texts: LoginAccountFinderDismissTexts, points: 20},
		},
		minScore: 70,
	},
	{
		state: UISaveLoginPrompt,
		signals: []signal{
			{label: "save heading", texts: SaveLoginTitleTexts, points: 50},
			{label: "save btn", texts: SaveLoginSaveTexts, points: 25},
			{label: "not-now btn", texts: SaveLoginLaterTexts, points: 25},
		},
		minScore: 70,
	},
	{
		state: UILogoutConfirmPrompt,
		signals: []signal{
			{label: "logout title", texts: LogoutConfirmTitleTexts, points: 50},
			{label: "logout btn", texts: LogoutConfirmButtonTexts, points: 30},
			{label: "cancel btn", texts: LogoutCancelTexts, points: 20},
		},
		minScore: 70,
	},
	{
		state: UILocaleSetupError,
		signals: []signal{
			{label: "indonesia setup msg", texts: LocaleSetupMessageTexts, points: 50},
			{label: "continue english", texts: LocaleSetupContinueEnglishTexts, points: 30},
			{label: "try again btn", texts: LocaleSetupTryAgainTexts, points: 20},
		},
		minScore: 70,
	},
	{
		state: UIFanpageHomeIntro,
		signals: []signal{
			{label: "intro title", texts: FanpageHomeIntroTitleTexts, points: 50},
			{label: "intro body", texts: FanpageHomeIntroBodyTexts, points: 35},
			{label: "skip", texts: FanpageHomeIntroSkipTexts, points: 15},
		},
		minScore: 70,
	},
	{
		state: UIPostPromotePrompt,
		signals: []signal{
			{label: "promote title", texts: PostPromoteTitleTexts, points: 50},
			{label: "promote btn", texts: PostPromoteButtonTexts, points: 25},
			{label: "later btn", texts: PostPromoteLaterTexts, points: 25},
		},
		minScore: 70,
	},
	{
		state: UILoggedIn,
		signals: []signal{
			{label: "composer", texts: []string{"What's on your mind", "Apa yang Anda pikirkan"}, points: 40},
			{label: "menu", texts: []string{"Menu"}, points: 30},
			{label: "feed", texts: []string{"News Feed", "Beranda", "Home"}, points: 20},
			{label: "stories", texts: []string{"Stories", "Cerita"}, points: 10},
		},
		minScore: 70,
	},
	{
		state: UIError,
		signals: []signal{
			{label: "wrong password", texts: []string{"Wrong password", "Kata sandi salah", "incorrect password"}, points: 50},
			{label: "try again", texts: []string{"Try again", "Coba lagi", "Something went wrong"}, points: 50},
		},
		minScore: 70,
	},
}

type Detector struct {
	resolver *ui.Resolver
	rules    []rule
}

func NewDetector() *Detector {
	return &Detector{resolver: ui.NewResolver(70), rules: facebookRules}
}

func (d *Detector) Detect(snap ui.Snapshot, pkg, activity string) Detection {
	var candidates []Detection
	now := time.Now()

	for _, r := range d.rules {
		score, evidence := d.scoreRuleFor(snap, pkg, r)
		if score < r.minScore {
			continue
		}
		conf := score / 100.0
		if conf > 1 {
			conf = 1
		}
		candidates = append(candidates, Detection{
			State:      r.state,
			Score:      score,
			Confidence: conf,
			Evidence:   evidence,
			Package:    pkg,
			Activity:   activity,
			At:         now,
		})
	}
	if len(candidates) == 0 {
		return Detection{State: UIUnknown, Confidence: 0, Package: pkg, Activity: activity, At: now}
	}
	return pickBestCandidate(candidates)
}

func pickBestCandidate(candidates []Detection) Detection {
	if len(candidates) == 0 {
		return Detection{State: UIUnknown}
	}
	best := candidates[0]
	for _, c := range candidates[1:] {
		bp, cp := Priority(best.State), Priority(c.State)
		if cp > bp || (cp == bp && c.Score > best.Score) {
			best = c
		}
	}
	return best
}

func (d *Detector) scoreRuleFor(snap ui.Snapshot, pkg string, r rule) (float64, []string) {
	switch r.state {
	case UIPermission:
		return d.scorePermission(snap, pkg, r)
	case UIKeyboardSettings:
		return d.scoreKeyboardSettings(snap, pkg, r)
	default:
		return d.scoreRule(snap, r)
	}
}

func (d *Detector) scoreRule(snap ui.Snapshot, r rule) (float64, []string) {
	var total float64
	var evidence []string
	for _, sig := range r.signals {
		if d.resolver.TextExists(snap, sig.texts) {
			total += sig.points
			evidence = append(evidence, sig.label+" (+"+strconv.Itoa(int(sig.points))+")")
		}
	}
	return total, evidence
}

// scorePermission requires dialog structure (allow+deny) and/or system permission window context.
func (d *Detector) scorePermission(snap ui.Snapshot, pkg string, r rule) (float64, []string) {
	total, evidence := d.scoreRule(snap, r)
	hasAllow := d.resolver.TextExists(snap, PermissionAllowTexts)
	hasDeny := d.resolver.TextExists(snap, PermissionDenyTexts)
	sysWindow := IsSystemPermissionPackage(pkg) || HasSystemPermissionElement(snap)

	if hasAllow && hasDeny {
		total += 15
		evidence = append(evidence, "allow_deny_pair (+15)")
	} else {
		total *= 0.5
		evidence = append(evidence, "incomplete_allow_deny (×0.5)")
	}
	if sysWindow {
		total += 15
		evidence = append(evidence, "system_permission_window (+15)")
	} else if IsFacebookPackage(pkg) && !(hasAllow && hasDeny) {
		return 0, nil
	}
	return total, evidence
}

// scoreKeyboardSettings requires Gboard/IME package context — not just any "Setelan" text elsewhere.
func (d *Detector) scoreKeyboardSettings(snap ui.Snapshot, pkg string, r rule) (float64, []string) {
	total, evidence := d.scoreRule(snap, r)
	if !IMEPackageContext(snap, pkg) {
		return 0, nil
	}
	total += 20
	evidence = append(evidence, "ime_package (+20)")
	return total, evidence
}
