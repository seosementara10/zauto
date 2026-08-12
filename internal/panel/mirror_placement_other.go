//go:build !windows

package panel

func mirrorStartXHeadless(_, _, fallback int) int { return fallback }
