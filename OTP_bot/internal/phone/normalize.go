package phone

import "strings"

// Normalize возвращает канонический вид телефона в формате +<digits>.
func Normalize(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	var digits strings.Builder
	digits.Grow(len(trimmed))
	for _, r := range trimmed {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	if digits.Len() == 0 {
		return ""
	}
	return "+" + digits.String()
}
