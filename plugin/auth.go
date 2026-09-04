package plugin

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var cookieRegex = regexp.MustCompile(`(?:^|;\s*)__Secure-commandcode_prod_\.session_token=([^;]+)`)

// RawAuthContent represents possible structures inside a commandcode credential JSON file.
type RawAuthContent struct {
	Type                   string `json:"type"`
	Provider               string `json:"provider"`
	ID                     string `json:"id"`
	Label                  string `json:"label"`
	Name                   string `json:"name"`
	Email                  string `json:"email"`
	SessionToken           string `json:"session_token"`
	CommandCodeSession     string `json:"commandcode_session_token"`
	Cookie                 string `json:"cookie"`
	Token                  string `json:"token"`
	UpstreamBase           string `json:"api_base"`
}

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

// ParseAuth handles auth.parse requests for Command Code credentials.
func ParseAuth(req AuthParseRequest) (AuthParseResponse, error) {
	lowerFileName := strings.ToLower(req.FileName)
	isCommandCodeFile := strings.HasPrefix(lowerFileName, "commandcode") && strings.HasSuffix(lowerFileName, ".json")
	isCommandCodeProvider := strings.EqualFold(req.Provider, PluginID)

	var content RawAuthContent
	var rawMap map[string]any
	if len(req.RawJSON) > 0 {
		if err := json.Unmarshal(req.RawJSON, &content); err == nil {
			_ = json.Unmarshal(req.RawJSON, &rawMap)
		}
	}

	isExplicitCommandCode := strings.EqualFold(content.Type, PluginID) ||
		strings.EqualFold(content.Provider, PluginID) ||
		content.SessionToken != "" ||
		content.CommandCodeSession != "" ||
		strings.Contains(content.Cookie, "__Secure-commandcode_prod_.session_token")

	if !isCommandCodeFile && !isCommandCodeProvider && !isExplicitCommandCode {
		return AuthParseResponse{Handled: false}, nil
	}

	// Extract session token
	token := content.SessionToken
	if token == "" {
		token = content.CommandCodeSession
	}
	if token == "" && content.Cookie != "" {
		token = ExtractSessionToken(content.Cookie)
	}
	if token == "" && (isCommandCodeFile || isCommandCodeProvider || isExplicitCommandCode) {
		token = content.Token
	}
	token = ExtractSessionToken(token)

	// Determine ID
	authID := content.ID
	if authID == "" && req.FileName != "" {
		base := filepath.Base(req.FileName)
		authID = strings.TrimSuffix(base, filepath.Ext(base))
	}
	if authID == "" {
		authID = "commandcode-default"
	}

	// Determine Label
	label := content.Label
	if label == "" {
		label = content.Name
	}
	if label == "" && content.Email != "" {
		label = fmt.Sprintf("Command Code (%s)", content.Email)
	}
	if label == "" {
		label = fmt.Sprintf("Command Code (%s)", authID)
	}

	// Build clean StorageJSON
	storageMap := map[string]any{
		"type":          PluginID,
		"provider":      PluginID,
		"session_token": token,
	}
	if content.Email != "" {
		storageMap["email"] = content.Email
	}
	if content.Label != "" {
		storageMap["label"] = content.Label
	}
	if content.UpstreamBase != "" {
		storageMap["api_base"] = content.UpstreamBase
	}
	for k, v := range rawMap {
		if _, exists := storageMap[k]; !exists {
			storageMap[k] = v
		}
	}
	storageJSON, _ := json.Marshal(storageMap)

	metadata := map[string]any{
		"type":          PluginID,
		"session_token": token,
	}
	if content.Email != "" {
		metadata["email"] = content.Email
	}

	attributes := map[string]string{
		"provider": PluginID,
	}

	authData := AuthData{
		Provider:         PluginID,
		ID:               authID,
		FileName:         req.FileName,
		Label:            label,
		Disabled:         false,
		StorageJSON:      storageJSON,
		Metadata:         metadata,
		Attributes:       attributes,
		NextRefreshAfter: time.Now().Add(24 * time.Hour).UTC(),
	}

	return AuthParseResponse{
		Handled: true,
		Auth:    authData,
	}, nil
}
