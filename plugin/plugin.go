package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	PluginID       = "commandcode"
	PluginName     = "commandcode"
	PluginVersion  = "0.4.0"
	PluginAuthor   = "zgs225"
	PluginRepo     = "https://github.com/zgs225/cliproxy-plugin-commandcode"
	PluginLogo     = "https://raw.githubusercontent.com/zgs225/cliproxy-plugin-commandcode/main/assets/logo.svg"
	SchemaVersion1 = 1
)

// PluginConfig holds the runtime configuration parsed from YAML.
type PluginConfig struct {
	mu              sync.RWMutex
	SessionToken    string   `yaml:"session_token" json:"session_token"`
	APIBase         string   `yaml:"api_base" json:"api_base"`
	OpenCodeAPIKey  string   `yaml:"opencode_api_key" json:"opencode_api_key"`
	OpenCodeAPIKeys []string `yaml:"opencode_api_keys" json:"opencode_api_keys"`
	OpenCodeAPIBase string   `yaml:"opencode_api_base" json:"opencode_api_base"`
}

// UpdateFromYAML updates the configuration from raw YAML bytes.
func (c *PluginConfig) UpdateFromYAML(raw []byte) error {
	if len(raw) == 0 {
		return nil
	}
	var tmp struct {
		SessionToken    string   `yaml:"session_token"`
		APIBase         string   `yaml:"api_base"`
		OpenCodeAPIKey  string   `yaml:"opencode_api_key"`
		OpenCodeAPIKeys []string `yaml:"opencode_api_keys"`
		OpenCodeAPIBase string   `yaml:"opencode_api_base"`
	}
	if err := yaml.Unmarshal(raw, &tmp); err != nil {
		return fmt.Errorf("unmarshal config_yaml: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if tmp.SessionToken != "" {
		c.SessionToken = ExtractSessionToken(tmp.SessionToken)
	}
	if tmp.APIBase != "" {
		c.APIBase = strings.TrimRight(tmp.APIBase, "/")
	}
	if tmp.OpenCodeAPIKey != "" {
		// OpenCode API key is a plain Bearer token; do not run it through
		// ExtractSessionToken (that is Command Code cookie specific).
		c.OpenCodeAPIKey = strings.TrimSpace(tmp.OpenCodeAPIKey)
	}
	// Merge rule: opencode_api_keys (YAML list) wins when non-empty after
	// trimming/dedup; otherwise opencode_api_key (scalar) degrades to a
	// single-key list; both empty means no keys.
	c.OpenCodeAPIKeys = normalizeOpenCodeKeys(tmp.OpenCodeAPIKeys)
	if len(c.OpenCodeAPIKeys) == 0 {
		if single := strings.TrimSpace(tmp.OpenCodeAPIKey); single != "" {
			c.OpenCodeAPIKeys = []string{single}
		}
	}
	if tmp.OpenCodeAPIBase != "" {
		c.OpenCodeAPIBase = strings.TrimRight(tmp.OpenCodeAPIBase, "/")
	}
	if c.APIBase == "" {
		c.APIBase = DefaultAPIBase
	}
	return nil
}

// GetSessionToken safely returns the session token.
func (c *PluginConfig) GetSessionToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.SessionToken
}

// SetSessionToken safely sets the session token.
func (c *PluginConfig) SetSessionToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SessionToken = ExtractSessionToken(token)
}

// GetAPIBase safely returns the API base URL.
func (c *PluginConfig) GetAPIBase() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.APIBase == "" {
		return DefaultAPIBase
	}
	return c.APIBase
}

// GetOpenCodeAPIKey safely returns the single configured OpenCode Go API key
// (scalar opencode_api_key field; kept for backward compatibility).
func (c *PluginConfig) GetOpenCodeAPIKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.OpenCodeAPIKey
}

// GetOpenCodeAPIKeys safely returns the configured OpenCode Go API keys.
// The list field wins; when it is empty the scalar OpenCodeAPIKey degrades
// to a single-key list (same merge rule as UpdateFromYAML). The returned
// slice is a copy; callers may not mutate it.
func (c *PluginConfig) GetOpenCodeAPIKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.OpenCodeAPIKeys) > 0 {
		out := make([]string, len(c.OpenCodeAPIKeys))
		copy(out, c.OpenCodeAPIKeys)
		return out
	}
	if c.OpenCodeAPIKey != "" {
		return []string{c.OpenCodeAPIKey}
	}
	return nil
}

// normalizeOpenCodeKeys trims each key, drops empties and dedups while
// preserving the original order.
func normalizeOpenCodeKeys(keys []string) []string {
	out := make([]string, 0, len(keys))
	seen := make(map[string]bool, len(keys))
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, k)
	}
	return out
}

// GetOpenCodeAPIBase safely returns the OpenCode Go API base URL,
// falling back to DefaultOpenCodeAPIBase when unset.
func (c *PluginConfig) GetOpenCodeAPIBase() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.OpenCodeAPIBase == "" {
		return DefaultOpenCodeAPIBase
	}
	return c.OpenCodeAPIBase
}

// Plugin encapsulates the Command Code plugin instance.
type Plugin struct {
	config *PluginConfig
}

var (
	defaultPlugin = NewPlugin()
)

// DefaultPlugin returns the singleton plugin instance.
func DefaultPlugin() *Plugin {
	return defaultPlugin
}

// NewPlugin creates a new Plugin instance.
func NewPlugin() *Plugin {
	return &Plugin{
		config: &PluginConfig{
			APIBase: DefaultAPIBase,
		},
	}
}

// HandleMethod dispatches an ABI call to the corresponding handler.
func (p *Plugin) HandleMethod(method string, requestBytes []byte) ([]byte, error) {
	switch method {
	case "plugin.register":
		return p.handleRegister(requestBytes)
	case "plugin.reconfigure":
		return p.handleReconfigure(requestBytes)
	case "plugin.quiesce", "plugin.shutdown":
		return OkEnvelope(map[string]any{"shutdown": true})

	case "management.register":
		return p.handleManagementRegister()
	case "management.handle":
		return p.handleManagementHandle(requestBytes)

	default:
		return ErrorEnvelope("unknown_method", "unknown method: "+method), nil
	}
}

func (p *Plugin) handleRegister(raw []byte) ([]byte, error) {
	if len(raw) > 0 {
		var req LifecycleRequest
		if err := json.Unmarshal(raw, &req); err == nil && len(req.ConfigYAML) > 0 {
			_ = p.config.UpdateFromYAML(req.ConfigYAML)
		}
	}
	return OkEnvelope(Registration{
		SchemaVersion: SchemaVersion1,
		Metadata: Metadata{
			Name:             PluginName,
			Version:          PluginVersion,
			Author:           PluginAuthor,
			GitHubRepository: PluginRepo,
			Logo:             PluginLogo,
			ConfigFields: []ConfigField{
				{
					Name:        "session_token",
					Type:        "string",
					Description: "Command Code session token (__Secure-commandcode_prod_.session_token cookie value)",
				},
				{
					Name:        "api_base",
					Type:        "string",
					Description: "Command Code API base URL (default: https://api.commandcode.ai)",
				},
				{
					Name:        "opencode_api_key",
					Type:        "string",
					Description: "OpenCode Go API key (single Bearer token; degraded path when opencode_api_keys is unset)",
				},
				{
					Name:        "opencode_api_keys",
					Type:        "string",
					Description: "OpenCode Go API keys as a YAML list (e.g. opencode_api_keys: [\"sk-KEY1\", \"sk-KEY2\"]); takes precedence over opencode_api_key",
				},
				{
					Name:        "opencode_api_base",
					Type:        "string",
					Description: "OpenCode Go API base URL (default: https://opencode.ai/zen/go/v1)",
				},
			},
		},
		Capabilities: RegistrationCapability{
			ManagementAPI: true,
		},
	})
}

func (p *Plugin) handleReconfigure(raw []byte) ([]byte, error) {
	if len(raw) > 0 {
		var req LifecycleRequest
		if err := json.Unmarshal(raw, &req); err == nil && len(req.ConfigYAML) > 0 {
			_ = p.config.UpdateFromYAML(req.ConfigYAML)
		}
	}
	return p.handleRegister(raw)
}

func (p *Plugin) handleManagementRegister() ([]byte, error) {
	resp, err := RegisterManagement()
	if err != nil {
		return ErrorEnvelope("management_register_error", err.Error()), nil
	}
	return OkEnvelope(resp)
}

func (p *Plugin) handleManagementHandle(raw []byte) ([]byte, error) {
	var req ManagementRequest
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &req); err != nil {
			return ErrorEnvelope("invalid_request", "failed to parse ManagementRequest: "+err.Error()), nil
		}
	}

	resp, err := HandleManagement(context.Background(), req, p.config)
	if err != nil {
		return ErrorEnvelope("management_handle_error", err.Error()), nil
	}
	return OkEnvelope(resp)
}

// OkEnvelope builds a successful Envelope response.
func OkEnvelope(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Envelope{OK: true, Result: raw})
}

// ErrorEnvelope builds an error Envelope response.
func ErrorEnvelope(code, message string) []byte {
	raw, _ := json.Marshal(Envelope{
		OK: false,
		Error: &EnvelopeError{
			Code:    code,
			Message: message,
		},
	})
	return raw
}
