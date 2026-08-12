package textutil

import "strings"

var adbTextReplacer = strings.NewReplacer(
	"\u2014", "-", // em dash
	"\u2013", "-", // en dash
	"\u2026", "...", // ellipsis
	"\u2018", "'", // left single quote
	"\u2019", "'", // right single quote
	"\u201c", "\"", // left double quote
	"\u201d", "\"", // right double quote
	"\u00a0", " ", // nbsp
)

// SanitizeADBText replaces common smart punctuation and drops non-ASCII runes
// so adb "input text" can type the full string reliably.
func SanitizeADBText(s string) string {
	s = strings.TrimSpace(adbTextReplacer.Replace(s))
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '\n', '\r', '\t':
			b.WriteByte(' ')
		default:
			if r >= 32 && r <= 126 {
				b.WriteRune(r)
			}
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
