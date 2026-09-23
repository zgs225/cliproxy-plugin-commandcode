package plugin

import (
	"encoding/json"
	"testing"
)

func TestPluginRegister_And_Reconfigure(t *testing.T) {
	p := NewPlugin()

	configYAML := []byte(`
session_token: "my-yaml-token"
api_base: "https://custom-api.commandcode.ai"
`)
	lifecycleReq, _ := json.Marshal(LifecycleRequest{ConfigYAML: configYAML})

	// Test plugin.register
	regBytes, err := p.HandleMethod("plugin.register", lifecycleReq)
	if err != nil {
		t.Fatalf("handleMethod(plugin.register) error: %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(regBytes, &env); err != nil {
		t.Fatalf("unmarshal envelope error: %v", err)
	}
	if !env.OK {
		t.Fatalf("expected env.OK=true, got false: %+v", env.Error)
	}

	var reg Registration
	if err := json.Unmarshal(env.Result, &reg); err != nil {
		t.Fatalf("unmarshal registration error: %v", err)
	}

	if reg.Metadata.Name != PluginName {
		t.Errorf("Metadata.Name = %q, want %q", reg.Metadata.Name, PluginName)
	}
	if reg.Metadata.Version != PluginVersion {
		t.Errorf("Metadata.Version = %q, want %q", reg.Metadata.Version, PluginVersion)
	}
	if reg.Capabilities.AuthProvider {
		t.Errorf("Capabilities.AuthProvider = true, want false")
	}
	if !reg.Capabilities.ManagementAPI {
		t.Errorf("Capabilities.ManagementAPI = false, want true")
	}

	// Verify config fields (v0.5.0: 5 → 6, adds commandcode_api_key)
	if len(reg.Metadata.ConfigFields) != 6 {
		t.Fatalf("ConfigFields len = %d, want 6", len(reg.Metadata.ConfigFields))
	}
	fieldNames := map[string]bool{}
	for _, f := range reg.Metadata.ConfigFields {
		fieldNames[f.Name] = true
	}
	if !fieldNames["session_token"] || !fieldNames["commandcode_api_key"] || !fieldNames["api_base"] || !fieldNames["opencode_api_key"] || !fieldNames["opencode_api_keys"] || !fieldNames["opencode_api_base"] {
		t.Errorf("ConfigFields missing expected fields: %+v", reg.Metadata.ConfigFields)
	}

	// Verify config parsed
	if p.config.GetSessionToken() != "my-yaml-token" {
		t.Errorf("SessionToken = %q, want my-yaml-token", p.config.GetSessionToken())
	}
	if p.config.GetAPIBase() != "https://custom-api.commandcode.ai" {
		t.Errorf("APIBase = %q, want https://custom-api.commandcode.ai", p.config.GetAPIBase())
	}
	// Test plugin.reconfigure
	reconfYAML := []byte(`
session_token: "new-token-abc"
`)
	reconfReq, _ := json.Marshal(LifecycleRequest{ConfigYAML: reconfYAML})
	reconfBytes, err := p.HandleMethod("plugin.reconfigure", reconfReq)
	if err != nil {
		t.Fatalf("handleMethod(plugin.reconfigure) error: %v", err)
	}
	var reconfEnv Envelope
	if err := json.Unmarshal(reconfBytes, &reconfEnv); err != nil || !reconfEnv.OK {
		t.Fatalf("reconfigure failed: %+v", reconfEnv)
	}
	if p.config.GetSessionToken() != "new-token-abc" {
		t.Errorf("SessionToken after reconfigure = %q, want new-token-abc", p.config.GetSessionToken())
	}
}

func TestPluginConfig_CommandCodeAPIKey(t *testing.T) {
	p := NewPlugin()
	// Trimmed, and NOT run through ExtractSessionToken (it is a plain
	// Bearer token, not a Command Code cookie string).
	configYAML := []byte("commandcode_api_key: \"  user_abc123xyz  \"\n")
	lifecycleReq, _ := json.Marshal(LifecycleRequest{ConfigYAML: configYAML})
	if _, err := p.HandleMethod("plugin.register", lifecycleReq); err != nil {
		t.Fatalf("handleMethod(plugin.register) error: %v", err)
	}
	if got := p.config.GetCommandCodeAPIKey(); got != "user_abc123xyz" {
		t.Errorf("CommandCodeAPIKey = %q, want user_abc123xyz (trimmed, raw)", got)
	}

	// Whitespace-only value clears the field.
	p2 := NewPlugin()
	spaceReq, _ := json.Marshal(LifecycleRequest{ConfigYAML: []byte("commandcode_api_key: \"   \"\n")})
	if _, err := p2.HandleMethod("plugin.register", spaceReq); err != nil {
		t.Fatalf("handleMethod(plugin.register) error: %v", err)
	}
	if got := p2.config.GetCommandCodeAPIKey(); got != "" {
		t.Errorf("CommandCodeAPIKey = %q, want empty (whitespace-only)", got)
	}

	// Omitting the key in a reconfigure must not clear a configured value.
	reconfReq, _ := json.Marshal(LifecycleRequest{ConfigYAML: []byte("session_token: \"tok\"\n")})
	if _, err := p.HandleMethod("plugin.reconfigure", reconfReq); err != nil {
		t.Fatalf("handleMethod(plugin.reconfigure) error: %v", err)
	}
	if got := p.config.GetCommandCodeAPIKey(); got != "user_abc123xyz" {
		t.Errorf("CommandCodeAPIKey after reconfigure = %q, want kept user_abc123xyz", got)
	}

	// Default is empty.
	if got := NewPlugin().config.GetCommandCodeAPIKey(); got != "" {
		t.Errorf("default CommandCodeAPIKey = %q, want empty", got)
	}
}

func TestPluginAuthIdentifier_NotHandled(t *testing.T) {
	p := NewPlugin()
	raw, err := p.HandleMethod("auth.identifier", nil)
	if err != nil {
		t.Fatalf("auth.identifier error: %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("envelope error: %+v", env)
	}
	if env.OK {
		t.Fatal("expected env.OK=false for auth.identifier")
	}
	if env.Error == nil || env.Error.Code != "unknown_method" {
		t.Errorf("Error = %+v, want code=unknown_method", env.Error)
	}
}

func TestPluginUnknownMethod(t *testing.T) {
	p := NewPlugin()
	raw, err := p.HandleMethod("unknown.method.test", nil)
	if err != nil {
		t.Fatalf("expected no go error, got %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if env.OK {
		t.Fatal("expected env.OK=false for unknown method")
	}
	if env.Error == nil || env.Error.Code != "unknown_method" {
		t.Errorf("Error = %+v, want code=unknown_method", env.Error)
	}
}

func TestEnvelopeError(t *testing.T) {
	raw := ErrorEnvelope("test_code", "test error message")
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if env.OK {
		t.Fatal("expected OK=false")
	}
	if env.Error.Code != "test_code" || env.Error.Message != "test error message" {
		t.Errorf("env.Error = %+v", env.Error)
	}
}

func TestPluginConfig_OpenCode(t *testing.T) {
	p := NewPlugin()
	configYAML := []byte("opencode_api_key: \" sk-opencode-123 \"\nopencode_api_base: \"https://custom.oc.example/v1/\"\n")
	lifecycleReq, _ := json.Marshal(LifecycleRequest{ConfigYAML: configYAML})

	if _, err := p.HandleMethod("plugin.register", lifecycleReq); err != nil {
		t.Fatalf("handleMethod(plugin.register) error: %v", err)
	}

	if got := p.config.GetOpenCodeAPIKey(); got != "sk-opencode-123" {
		t.Errorf("OpenCodeAPIKey = %q, want sk-opencode-123", got)
	}
	if got := p.config.GetOpenCodeAPIBase(); got != "https://custom.oc.example/v1" {
		t.Errorf("OpenCodeAPIBase = %q, want https://custom.oc.example/v1 (trailing slash trimmed)", got)
	}

	// A Command Code cookie string must NOT be run through ExtractSessionToken.
	cookieLike := []byte("opencode_api_key: \"sk-raw-bearer-value\"\n")
	req2, _ := json.Marshal(LifecycleRequest{ConfigYAML: cookieLike})
	if _, err := p.HandleMethod("plugin.reconfigure", req2); err != nil {
		t.Fatalf("handleMethod(plugin.reconfigure) error: %v", err)
	}
	if got := p.config.GetOpenCodeAPIKey(); got != "sk-raw-bearer-value" {
		t.Errorf("OpenCodeAPIKey = %q, want sk-raw-bearer-value (raw, no cookie extraction)", got)
	}

	// Empty config falls back to the default base.
	empty := NewPlugin()
	if got := empty.config.GetOpenCodeAPIBase(); got != DefaultOpenCodeAPIBase {
		t.Errorf("default OpenCodeAPIBase = %q, want %q", got, DefaultOpenCodeAPIBase)
	}
	if got := empty.config.GetOpenCodeAPIKey(); got != "" {
		t.Errorf("default OpenCodeAPIKey = %q, want empty", got)
	}
	if got := empty.config.GetOpenCodeAPIKeys(); len(got) != 0 {
		t.Errorf("default GetOpenCodeAPIKeys = %v, want empty", got)
	}
}

func TestPluginConfig_OpenCodeAPIKeys(t *testing.T) {
	newCfg := func(t *testing.T, yaml string) *PluginConfig {
		t.Helper()
		cfg := &PluginConfig{}
		if err := cfg.UpdateFromYAML([]byte(yaml)); err != nil {
			t.Fatalf("UpdateFromYAML error: %v", err)
		}
		return cfg
	}

	t.Run("list takes precedence over scalar", func(t *testing.T) {
		cfg := newCfg(t, `
opencode_api_key: "sk-scalar"
opencode_api_keys:
  - " sk-key1 "
  - "sk-key2"
`)
		got := cfg.GetOpenCodeAPIKeys()
		if len(got) != 2 || got[0] != "sk-key1" || got[1] != "sk-key2" {
			t.Errorf("GetOpenCodeAPIKeys = %v, want [sk-key1 sk-key2] (list wins, trimmed)", got)
		}
		if cfg.GetOpenCodeAPIKey() != "sk-scalar" {
			t.Errorf("GetOpenCodeAPIKey = %q, want sk-scalar (scalar field kept)", cfg.GetOpenCodeAPIKey())
		}
	})

	t.Run("scalar degrades to single-key list", func(t *testing.T) {
		cfg := newCfg(t, `
opencode_api_key: " sk-only "
`)
		got := cfg.GetOpenCodeAPIKeys()
		if len(got) != 1 || got[0] != "sk-only" {
			t.Errorf("GetOpenCodeAPIKeys = %v, want [sk-only]", got)
		}
	})

	t.Run("dedup preserve order and drop empties", func(t *testing.T) {
		cfg := newCfg(t, `
opencode_api_keys:
  - "sk-b"
  - ""
  - "  "
  - "sk-a"
  - "sk-b"
  - "sk-c"
  - "sk-a"
`)
		got := cfg.GetOpenCodeAPIKeys()
		want := []string{"sk-b", "sk-a", "sk-c"}
		if len(got) != len(want) {
			t.Fatalf("GetOpenCodeAPIKeys = %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("GetOpenCodeAPIKeys[%d] = %q, want %q (order preserved, deduped)", i, got[i], want[i])
			}
		}
	})

	t.Run("all empty yields no keys", func(t *testing.T) {
		for _, yaml := range []string{
			`opencode_api_key: ""`,
			"opencode_api_keys: []\nopencode_api_key: \"   \"",
			"opencode_api_keys:\n  - \"\"\n  - \"   \"",
		} {
			cfg := newCfg(t, yaml)
			if got := cfg.GetOpenCodeAPIKeys(); len(got) != 0 {
				t.Errorf("yaml %q: GetOpenCodeAPIKeys = %v, want empty", yaml, got)
			}
		}
	})

	t.Run("getter returns a copy", func(t *testing.T) {
		cfg := newCfg(t, "opencode_api_keys:\n  - sk-a\n  - sk-b\n")
		got := cfg.GetOpenCodeAPIKeys()
		got[0] = "mutated"
		again := cfg.GetOpenCodeAPIKeys()
		if again[0] != "sk-a" {
			t.Errorf("GetOpenCodeAPIKeys not a copy: after mutation got %q", again[0])
		}
	})
}
