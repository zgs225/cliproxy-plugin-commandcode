package plugin

import (
	"regexp"
	"strings"
)

var cookieRegex = regexp.MustCompile(`(?:^|;\s*)__Secure-commandcode_prod_\.session_token=([^;]+)`)

// ExtractSessionToken extracts the clean session token from a raw string or cookie string.
func ExtractSessionToken(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if matches := cookieRegex.FindStringSubmatch(raw); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	if strings.HasPrefix(raw, "__Secure-commandcode_prod_.session_token=") {
		trimmed := strings.TrimPrefix(raw, "__Secure-commandcode_prod_.session_token=")
		if idx := strings.Index(trimmed, ";"); idx != -1 {
			trimmed = trimmed[:idx]
		}
		return strings.TrimSpace(trimmed)
	}
	return raw
}

// FormatSessionCookie ensures the token is formatted as the upstream Cookie header value.
func FormatSessionCookie(token string) string {
	clean := ExtractSessionToken(token)
	if clean == "" {
		return ""
	}
	return "__Secure-commandcode_prod_.session_token=" + clean
}

