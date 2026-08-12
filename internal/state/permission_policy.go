package state

import (
	"zauto/internal/ui"
)

// PermissionKind identifies which Android permission dialog is shown.
type PermissionKind string

const (
	PermKindNotification PermissionKind = "notification"
	PermKindContacts     PermissionKind = "contacts"
	PermKindLocation     PermissionKind = "location"
	PermKindCamera       PermissionKind = "camera"
	PermKindUnknown      PermissionKind = "unknown"
)

// PermissionAction is allow, deny, or review (fail-closed stop).
type PermissionAction string

const (
	PermActionDeny   PermissionAction = "deny"
	PermActionAllow  PermissionAction = "allow"
	PermActionReview PermissionAction = "review"
)

// PermissionPolicy maps permission kinds to actions. Unconfigured kinds fail closed (review).
type PermissionPolicy struct {
	rules map[PermissionKind]PermissionAction
}

// DefaultPermissionPolicy denies notification and contacts; unknown kinds require review.
func DefaultPermissionPolicy() PermissionPolicy {
	return PermissionPolicy{rules: map[PermissionKind]PermissionAction{
		PermKindNotification: PermActionDeny,
		PermKindContacts:     PermActionDeny,
		PermKindLocation:     PermActionDeny,
		PermKindCamera:       PermActionDeny,
	}}
}

// UniformPermissionPolicy applies one action to all known permission kinds (config override).
func UniformPermissionPolicy(action PermissionAction) PermissionPolicy {
	return PermissionPolicy{rules: map[PermissionKind]PermissionAction{
		PermKindNotification: action,
		PermKindContacts:     action,
		PermKindLocation:     action,
		PermKindCamera:       action,
	}}
}

// ActionFor returns the configured action. Unknown kinds or missing rules → review (fail closed).
func (p PermissionPolicy) ActionFor(kind PermissionKind) PermissionAction {
	if kind == PermKindUnknown {
		return PermActionReview
	}
	if act, ok := p.rules[kind]; ok {
		return act
	}
	return PermActionReview
}

// IdentifyPermissionKind inspects UI to determine permission type (no taps).
func IdentifyPermissionKind(resolver *ui.Resolver, snap ui.Snapshot, pkg string) PermissionKind {
	if resolver.TextExists(snap, NotificationPermissionPromptTexts) {
		return PermKindNotification
	}
	if resolver.TextExists(snap, ContactPermissionPromptTexts) {
		return PermKindContacts
	}
	if resolver.TextExists(snap, LocationPermissionPromptTexts) {
		return PermKindLocation
	}
	if resolver.TextExists(snap, CameraPermissionPromptTexts) {
		return PermKindCamera
	}
	if IsSystemPermissionPackage(pkg) || HasSystemPermissionElement(snap) {
		return PermKindUnknown
	}
	return PermKindUnknown
}
