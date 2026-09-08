package plugin

import (
	"testing"
)

func TestExtractSessionToken(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "raw token",
			input:    "abc123token",
			expected: "abc123token",
		},
		{
			name:     "single cookie",
			input:    "__Secure-commandcode_prod_.session_token=secret_tok_123",
			expected: "secret_tok_123",
		},
		{
			name:     "cookie with semicolons and trailing params",
			input:    "__Secure-commandcode_prod_.session_token=secret_tok_123; Path=/; Secure; HttpOnly",
			expected: "secret_tok_123",
		},
		{
			name:     "multi-cookie string",
			input:    "some_other_cookie=xyz; __Secure-commandcode_prod_.session_token=secret_tok_123; foo=bar",
			expected: "secret_tok_123",
		},
		{
			name:     "empty string",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractSessionToken(tt.input)
			if got != tt.expected {
				t.Errorf("ExtractSessionToken(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatSessionCookie(t *testing.T) {
	got := FormatSessionCookie("my-token")
	want := "__Secure-commandcode_prod_.session_token=my-token"
	if got != want {
		t.Errorf("FormatSessionCookie() = %q, want %q", got, want)
	}

	gotCookie := FormatSessionCookie("__Secure-commandcode_prod_.session_token=my-token; Path=/")
	if gotCookie != want {
		t.Errorf("FormatSessionCookie() from cookie = %q, want %q", gotCookie, want)
	}
}
