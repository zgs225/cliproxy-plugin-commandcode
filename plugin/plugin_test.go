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
	if !reg.Capabilities.AuthProvider {
		t.Errorf("Capabilities.AuthProvider = false, want true")
	}
	if !reg.Capabilities.ManagementAPI {
		t.Errorf("Capabilities.ManagementAPI = false, want true")
	}

	// Verify config fields
	if len(reg.Metadata.ConfigFields) != 2 {
		t.Fatalf("ConfigFields len = %d, want 2", len(reg.Metadata.ConfigFields))
	}
	fieldNames := map[string]bool{}
	for _, f := range reg.Metadata.ConfigFields {
		fieldNames[f.Name] = true
	}
	if !fieldNames["session_token"] || !fieldNames["api_base"] {
		t.Errorf("ConfigFields missing session_token or api_base: %+v", reg.Metadata.ConfigFields)
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

func TestPluginAuthIdentifier(t *testing.T) {
	p := NewPlugin()
	raw, err := p.HandleMethod("auth.identifier", nil)
	if err != nil {
		t.Fatalf("auth.identifier error: %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil || !env.OK {
		t.Fatalf("envelope error: %+v", env)
	}

	var idResp IdentifierResponse
	if err := json.Unmarshal(env.Result, &idResp); err != nil {
		t.Fatalf("unmarshal idResp error: %v", err)
	}
	if idResp.Identifier != PluginID {
		t.Errorf("Identifier = %q, want %q", idResp.Identifier, PluginID)
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
