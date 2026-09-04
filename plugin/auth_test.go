package plugin

import (
	"encoding/json"
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

func TestParseAuth_ExplicitJSON(t *testing.T) {
	raw := []byte(`{
		"type": "commandcode",
		"session_token": "test-session-token-xyz",
		"email": "user@example.com",
		"label": "My Command Code Auth"
	}`)

	resp, err := ParseAuth(AuthParseRequest{
		FileName: "custom.json",
		RawJSON:  raw,
	})
	if err != nil {
		t.Fatalf("ParseAuth error: %v", err)
	}
	if !resp.Handled {
		t.Fatal("expected Handled=true for explicit commandcode type")
	}

	auth := resp.Auth
	if auth.Provider != PluginID {
		t.Errorf("Provider = %q, want %q", auth.Provider, PluginID)
	}
	if auth.ID != "custom" {
		t.Errorf("ID = %q, want %q", auth.ID, "custom")
	}
	if auth.Label != "My Command Code Auth" {
		t.Errorf("Label = %q, want %q", auth.Label, "My Command Code Auth")
	}

	var storage map[string]any
	if err := json.Unmarshal(auth.StorageJSON, &storage); err != nil {
		t.Fatalf("failed to unmarshal StorageJSON: %v", err)
	}
	if storage["session_token"] != "test-session-token-xyz" {
		t.Errorf("StorageJSON session_token = %v, want test-session-token-xyz", storage["session_token"])
	}

	if auth.Metadata["session_token"] != "test-session-token-xyz" {
		t.Errorf("Metadata session_token = %v, want test-session-token-xyz", auth.Metadata["session_token"])
	}
}

func TestParseAuth_FileNameMatch(t *testing.T) {
	raw := []byte(`{
		"token": "tok_987654"
	}`)

	resp, err := ParseAuth(AuthParseRequest{
		FileName: "commandcode-work.json",
		RawJSON:  raw,
	})
	if err != nil {
		t.Fatalf("ParseAuth error: %v", err)
	}
	if !resp.Handled {
		t.Fatal("expected Handled=true for commandcode-*.json filename")
	}
	if resp.Auth.ID != "commandcode-work" {
		t.Errorf("ID = %q, want commandcode-work", resp.Auth.ID)
	}
	if resp.Auth.Metadata["session_token"] != "tok_987654" {
		t.Errorf("session_token = %v, want tok_987654", resp.Auth.Metadata["session_token"])
	}
}

func TestParseAuth_CookieFormat(t *testing.T) {
	raw := []byte(`{
		"cookie": "__Secure-commandcode_prod_.session_token=cookie_tok_456; Path=/"
	}`)

	resp, err := ParseAuth(AuthParseRequest{
		FileName: "any.json",
		RawJSON:  raw,
	})
	if err != nil {
		t.Fatalf("ParseAuth error: %v", err)
	}
	if !resp.Handled {
		t.Fatal("expected Handled=true for cookie with __Secure-commandcode_prod_.session_token")
	}
	if resp.Auth.Metadata["session_token"] != "cookie_tok_456" {
		t.Errorf("session_token = %v, want cookie_tok_456", resp.Auth.Metadata["session_token"])
	}
}

func TestParseAuth_UnrelatedFile(t *testing.T) {
	raw := []byte(`{
		"type": "openai",
		"api_key": "sk-123456"
	}`)

	resp, err := ParseAuth(AuthParseRequest{
		FileName: "openai-test.json",
		RawJSON:  raw,
	})
	if err != nil {
		t.Fatalf("ParseAuth error: %v", err)
	}
	if resp.Handled {
		t.Fatal("expected Handled=false for unrelated credential file")
	}
}
