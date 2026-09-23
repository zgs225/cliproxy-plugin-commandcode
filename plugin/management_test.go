package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRegisterManagement(t *testing.T) {
	resp, err := RegisterManagement()
	if err != nil {
		t.Fatalf("RegisterManagement error: %v", err)
	}

	if len(resp.Routes) != 6 {
		t.Fatalf("len(Routes) = %d, want 6", len(resp.Routes))
	}
	if resp.Routes[0].Method != http.MethodGet || resp.Routes[0].Path != "/plugins/commandcode/usage" {
		t.Errorf("Route 0 mismatch: %+v", resp.Routes[0])
	}
	if resp.Routes[1].Method != http.MethodPost || resp.Routes[1].Path != "/plugins/commandcode/usage" {
		t.Errorf("Route 1 mismatch: %+v", resp.Routes[1])
	}
	wantOpencode := []struct{ method, path string }{
		{http.MethodGet, "/plugins/commandcode/opencode/usage"},
		{http.MethodPost, "/plugins/commandcode/opencode/usage"},
		{http.MethodGet, "/plugins/commandcode/all"},
		{http.MethodPost, "/plugins/commandcode/all"},
	}
	for i, w := range wantOpencode {
		if resp.Routes[2+i].Method != w.method || resp.Routes[2+i].Path != w.path {
			t.Errorf("Route %d mismatch: got %+v, want %s %s", 2+i, resp.Routes[2+i], w.method, w.path)
		}
	}

	if len(resp.Resources) != 1 {
		t.Fatalf("len(Resources) = %d, want 1", len(resp.Resources))
	}
	if resp.Resources[0].Path != "/quota" || resp.Resources[0].Menu != "用量配额" {
		t.Errorf("Resource 0 mismatch: %+v", resp.Resources[0])
	}
	if resp.Resources[0].Description != "Command Code + OpenCode Go 用量与限额卡片" {
		t.Errorf("Resource Description mismatch: %+v", resp.Resources[0])
	}
}

func TestHandleManagement_QuotaResource(t *testing.T) {
	paths := []string{
		"/quota",
		"/v0/resource/plugins/commandcode/quota",
	}

	for _, p := range paths {
		req := ManagementRequest{
			Method: http.MethodGet,
			Path:   p,
		}
		resp, err := HandleManagement(context.Background(), req, nil)
		if err != nil {
			t.Fatalf("HandleManagement error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
		}
		ct := resp.Headers["Content-Type"]
		if len(ct) == 0 || !strings.Contains(ct[0], "text/html") {
			t.Errorf("Content-Type = %v, want text/html", ct)
		}
		bodyStr := string(resp.Body)
		if !strings.Contains(bodyStr, "用量配额") {
			t.Errorf("Body does not contain expected menu text 用量配额")
		}
		if !strings.Contains(bodyStr, "v0.5.0") {
			t.Errorf("Body does not contain version badge v0.5.0")
		}
	}
}

func TestHandleManagement_GetUsage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"credits": {"monthlyCredits": 888},
			"windowLimits": {"fiveHour": {"used": 2, "cap": 20}}
		}`))
	}))
	defer ts.Close()

	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())

	cfg := &PluginConfig{
		SessionToken: "configured-token",
		APIBase:      ts.URL,
	}

	req := ManagementRequest{
		Method: http.MethodGet,
		Path:   "/v0/management/plugins/commandcode/usage",
	}

	resp, err := HandleManagement(context.Background(), req, cfg)
	if err != nil {
		t.Fatalf("HandleManagement error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200, body=%s", resp.StatusCode, string(resp.Body))
	}

	var usage FormattedUsageResponse
	if err := json.Unmarshal(resp.Body, &usage); err != nil {
		t.Fatalf("Unmarshal body error: %v", err)
	}
	if !usage.OK {
		t.Fatal("expected OK=true")
	}
	if usage.Credits.MonthlyCredits != 888 {
		t.Errorf("MonthlyCredits = %v, want 888", usage.Credits.MonthlyCredits)
	}
}

func TestHandleManagement_PostUsage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie := r.Header.Get("Cookie")
		if !strings.Contains(cookie, "post-token-999") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"credits": {"monthlyCredits": 666},
			"windowLimits": {"fiveHour": {"used": 1, "cap": 10}}
		}`))
	}))
	defer ts.Close()

	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())

	reqBody, _ := json.Marshal(map[string]string{
		"session_token": "post-token-999",
		"api_base":      ts.URL,
	})

	req := ManagementRequest{
		Method: http.MethodPost,
		Path:   "/plugins/commandcode/usage",
		Body:   reqBody,
	}

	resp, err := HandleManagement(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("HandleManagement error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200, body=%s", resp.StatusCode, string(resp.Body))
	}

	var usage FormattedUsageResponse
	if err := json.Unmarshal(resp.Body, &usage); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if usage.Credits.MonthlyCredits != 666 {
		t.Errorf("MonthlyCredits = %v, want 666", usage.Credits.MonthlyCredits)
	}
}

const mockOpencodeUsageJSON = `{"usage":{
	"rolling": {"status":"ok","percent":4, "resetsAt":"2026-09-17T06:58:53.171Z"},
	"weekly":  {"status":"ok","percent":46,"resetsAt":"2026-09-21T00:00:00.000Z"},
	"monthly": {"status":"ok","percent":23,"resetsAt":"2026-10-14T09:13:49.000Z"}
}}`

// mockOpencodeUsageJSONFuture returns the same usage envelope but with reset
// timestamps relative to now. The hardcoded dates in mockOpencodeUsageJSON
// eventually fall into the past (weekly 2026-09-21 did), which makes
// weekly.ResetInSeconds = 0 and breaks the `want > 0` assertion in
// TestHandleManagement_OpencodeUsageRoute. Use this fixture for tests that
// assert positive reset_in_seconds.
func mockOpencodeUsageJSONFuture() string {
	now := time.Now().UTC()
	ts := func(d time.Duration) string {
		return now.Add(d).Format("2006-01-02T15:04:05.000Z")
	}
	return fmt.Sprintf(`{"usage":{
	"rolling": {"status":"ok","percent":4, "resetsAt":"%s"},
	"weekly":  {"status":"ok","percent":46,"resetsAt":"%s"},
	"monthly": {"status":"ok","percent":23,"resetsAt":"%s"}
}}`, ts(2*time.Hour), ts(72*time.Hour), ts(30*24*time.Hour))
}

// Verifies that /plugins/commandcode/opencode/usage is matched by the dedicated
// OpenCode handler and NOT swallowed by the generic "/usage" suffix match
// (which would route it to the Command Code handler).
func TestHandleManagement_OpencodeUsageRoute(t *testing.T) {
	var sawAuthHeader bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/usage" {
			t.Errorf("unexpected path %s, want /usage (Command Code handler must not be hit)", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk-opencode-override" {
			t.Errorf("Authorization = %q, want Bearer sk-opencode-override", got)
		}
		if r.Header.Get("Cookie") != "" {
			t.Errorf("unexpected Cookie header on opencode request: %q", r.Header.Get("Cookie"))
		}
		sawAuthHeader = true
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(mockOpencodeUsageJSONFuture()))
	}))
	defer ts.Close()

	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())
	defer func() {
		SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})
	}()

	reqBody, _ := json.Marshal(map[string]string{
		"opencode_api_key":  "sk-opencode-override",
		"opencode_api_base": ts.URL,
	})

	for _, tc := range []struct {
		method string
		body   []byte
	}{
		{http.MethodPost, reqBody},
		// GET with configured plugin config (no query override by design).
	} {
		t.Run(tc.method, func(t *testing.T) {
			cfg := &PluginConfig{
				OpenCodeAPIKey:  "sk-configured",
				OpenCodeAPIBase: ts.URL,
			}
			req := ManagementRequest{
				Method: tc.method,
				Path:   "/v0/management/plugins/commandcode/opencode/usage",
				Body:   tc.body,
			}
			resp, err := HandleManagement(context.Background(), req, cfg)
			if err != nil {
				t.Fatalf("HandleManagement error: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("StatusCode = %d, want 200, body=%s", resp.StatusCode, string(resp.Body))
			}
			// v0.4.0: the response is the multi-key envelope even for a single key.
			var usage OpenCodeMultiKeyResponse
			if err := json.Unmarshal(resp.Body, &usage); err != nil {
				t.Fatalf("unmarshal body error: %v", err)
			}
			if !usage.OK || usage.Provider != "opencode_go" {
				t.Fatalf("unexpected response: ok=%v provider=%q", usage.OK, usage.Provider)
			}
			if len(usage.Keys) != 1 || !usage.Keys[0].OK {
				t.Fatalf("expected exactly one successful key, got %+v", usage.Keys)
			}
			if usage.Keys[0].Windows == nil {
				t.Fatal("keys[0].windows = nil, want non-nil on success")
			}
			if usage.Keys[0].Windows.Rolling.Percent != 4 || usage.Keys[0].Windows.Weekly.Percent != 46 || usage.Keys[0].Windows.Monthly.Percent != 23 {
				t.Errorf("windows percents = %v/%v/%v, want 4/46/23",
					usage.Keys[0].Windows.Rolling.Percent, usage.Keys[0].Windows.Weekly.Percent, usage.Keys[0].Windows.Monthly.Percent)
			}
			if usage.Keys[0].Windows.Weekly.ResetInSeconds <= 0 {
				t.Errorf("weekly reset_in_seconds = %d, want > 0", usage.Keys[0].Windows.Weekly.ResetInSeconds)
			}
			// The raw key from the POST body must never appear in the response.
			if strings.Contains(string(resp.Body), "sk-opencode-override") && tc.method == http.MethodPost {
				t.Errorf("response leaks the raw override key: %s", string(resp.Body))
			}
		})
	}

	if !sawAuthHeader {
		t.Fatal("upstream never received Authorization header")
	}
}

// /opencode/usage status matrix (v0.4.0): >=1 key success → 200; one success
// + one 401 → 200 with keys[1].ok=false and no windows; all 401 → 502; no
// keys configured → 400 with the "no opencode api keys configured" prefix.
func TestHandleManagement_OpencodeUsage_MultiKeyMatrix(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get("Authorization") {
		case "Bearer sk-good-AAAA", "Bearer sk-good-ZZZZ":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(mockOpencodeUsageJSON))
		default:
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid api key"}`))
		}
	}))
	defer ts.Close()

	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())
	defer func() {
		SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})
	}()

	postKeys := func(keys ...string) ManagementRequest {
		body, _ := json.Marshal(map[string]any{
			"opencode_api_keys": keys,
			"opencode_api_base": ts.URL,
		})
		return ManagementRequest{
			Method: http.MethodPost,
			Path:   "/plugins/commandcode/opencode/usage",
			Body:   body,
		}
	}

	t.Run("both keys succeed → 200", func(t *testing.T) {
		resp, err := HandleManagement(context.Background(), postKeys("sk-good-AAAA", "sk-good-ZZZZ"), nil)
		if err != nil {
			t.Fatalf("HandleManagement error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("StatusCode = %d, want 200, body=%s", resp.StatusCode, string(resp.Body))
		}
		var usage OpenCodeMultiKeyResponse
		if err := json.Unmarshal(resp.Body, &usage); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if !usage.OK || len(usage.Keys) != 2 || !usage.Keys[0].OK || !usage.Keys[1].OK {
			t.Errorf("unexpected response: %+v", usage)
		}
		if usage.Error != "" {
			t.Errorf("top-level error = %q, want empty when keys are configured", usage.Error)
		}
	})

	t.Run("one success one 401 → 200 with failed key isolated", func(t *testing.T) {
		resp, err := HandleManagement(context.Background(), postKeys("sk-good-AAAA", "sk-bad-BBBB"), nil)
		if err != nil {
			t.Fatalf("HandleManagement error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("StatusCode = %d, want 200 (>=1 key succeeded), body=%s", resp.StatusCode, string(resp.Body))
		}
		var usage OpenCodeMultiKeyResponse
		if err := json.Unmarshal(resp.Body, &usage); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if !usage.OK || len(usage.Keys) != 2 {
			t.Fatalf("unexpected response: %+v", usage)
		}
		if !usage.Keys[0].OK || usage.Keys[0].Windows == nil {
			t.Errorf("keys[0] = %+v, want ok with windows", usage.Keys[0])
		}
		if usage.Keys[1].OK || usage.Keys[1].Windows != nil {
			t.Errorf("keys[1] = %+v, want not-ok with nil windows", usage.Keys[1])
		}
		if usage.Keys[1].StatusCode != http.StatusUnauthorized {
			t.Errorf("keys[1].status_code = %d, want 401", usage.Keys[1].StatusCode)
		}
		// windows must be omitted from the JSON for the failed key, not
		// serialized as null or a zero-value struct.
		var raw struct {
			Keys []struct {
				Windows json.RawMessage `json:"windows"`
			} `json:"keys"`
		}
		if err := json.Unmarshal(resp.Body, &raw); err != nil {
			t.Fatalf("unmarshal raw error: %v", err)
		}
		if len(raw.Keys[1].Windows) != 0 {
			t.Errorf("keys[1].windows in JSON = %s, want omitted", string(raw.Keys[1].Windows))
		}
		if !strings.Contains(usage.Keys[1].Error, "opencode upstream returned 401") {
			t.Errorf("keys[1].error = %q, want upstream 401 mention", usage.Keys[1].Error)
		}
		if strings.Contains(string(resp.Body), "sk-bad-BBBB") {
			t.Errorf("response leaks the raw key: %s", string(resp.Body))
		}
	})

	t.Run("all keys 401 → 502", func(t *testing.T) {
		resp, err := HandleManagement(context.Background(), postKeys("sk-bad-CCCC", "sk-bad-DDDD"), nil)
		if err != nil {
			t.Fatalf("HandleManagement error: %v", err)
		}
		if resp.StatusCode != http.StatusBadGateway {
			t.Fatalf("StatusCode = %d, want 502 (all keys upstream-failed), body=%s", resp.StatusCode, string(resp.Body))
		}
		var usage OpenCodeMultiKeyResponse
		if err := json.Unmarshal(resp.Body, &usage); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if usage.OK {
			t.Errorf("OK = true, want false when all keys fail")
		}
		if usage.Error != "" {
			t.Errorf("top-level error = %q, want empty (per-key errors carry the detail)", usage.Error)
		}
	})

	t.Run("no keys configured → 400", func(t *testing.T) {
		req := ManagementRequest{
			Method: http.MethodGet,
			Path:   "/v0/management/plugins/commandcode/opencode/usage",
		}
		resp, err := HandleManagement(context.Background(), req, &PluginConfig{})
		if err != nil {
			t.Fatalf("HandleManagement error: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("StatusCode = %d, want 400, body=%s", resp.StatusCode, string(resp.Body))
		}
		var usage OpenCodeMultiKeyResponse
		if err := json.Unmarshal(resp.Body, &usage); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if !strings.HasPrefix(usage.Error, "no opencode api keys configured") {
			t.Errorf("top-level error = %q, want prefix 'no opencode api keys configured'", usage.Error)
		}
	})

	t.Run("POST body opencode_api_keys overrides config and wins over scalar", func(t *testing.T) {
		body, _ := json.Marshal(map[string]any{
			"opencode_api_key":  "sk-scalar-must-lose",
			"opencode_api_keys": []string{"sk-good-AAAA"},
			"opencode_api_base": ts.URL,
		})
		req := ManagementRequest{
			Method: http.MethodPost,
			Path:   "/plugins/commandcode/opencode/usage",
			Body:   body,
		}
		cfg := &PluginConfig{OpenCodeAPIKey: "sk-config-must-lose", OpenCodeAPIBase: ts.URL}
		resp, err := HandleManagement(context.Background(), req, cfg)
		if err != nil {
			t.Fatalf("HandleManagement error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("StatusCode = %d, want 200, body=%s", resp.StatusCode, string(resp.Body))
		}
		var usage OpenCodeMultiKeyResponse
		if err := json.Unmarshal(resp.Body, &usage); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if len(usage.Keys) != 1 || usage.Keys[0].KeyID != MaskAPIKey("sk-good-AAAA") {
			t.Errorf("keys = %+v, want only the body-list key (list wins over scalar and config)", usage.Keys)
		}
	})
}

// Regression: /plugins/commandcode/all must not be swallowed by the generic
// "/usage" suffix match nor miss its dedicated handler.
func TestHandleManagement_AllRoute_BothProvidersOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/usage" && r.Header.Get("Authorization") != "":
			_, _ = w.Write([]byte(mockOpencodeUsageJSON))
		case r.URL.Path == "/internal/billing/credits":
			_, _ = w.Write([]byte(`{"credits":{"monthlyCredits":888},"windowLimits":{"fiveHour":{"used":2,"cap":20}}}`))
		case r.URL.Path == "/internal/usage/summary":
			_, _ = w.Write([]byte(`{"totalMonthlyCredits": 100}`))
		default:
			t.Errorf("unexpected upstream request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())
	defer func() {
		SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})
	}()

	cfg := &PluginConfig{
		SessionToken:    "configured-token",
		APIBase:         ts.URL,
		OpenCodeAPIKey:  "sk-configured",
		OpenCodeAPIBase: ts.URL,
	}

	req := ManagementRequest{
		Method: http.MethodGet,
		Path:   "/v0/management/plugins/commandcode/all",
	}
	resp, err := HandleManagement(context.Background(), req, cfg)
	if err != nil {
		t.Fatalf("HandleManagement error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200, body=%s", resp.StatusCode, string(resp.Body))
	}

	var all AllUsageResponse
	if err := json.Unmarshal(resp.Body, &all); err != nil {
		t.Fatalf("unmarshal body error: %v", err)
	}
	if !all.OK {
		t.Fatal("expected ok=true when both providers succeed")
	}
	if len(all.CommandCode) == 0 || len(all.OpenCode) == 0 {
		t.Fatalf("expected both provider payloads, got commandcode=%d bytes opencode=%d bytes",
			len(all.CommandCode), len(all.OpenCode))
	}
	if len(all.Errors) != 0 {
		t.Errorf("expected empty errors map, got %v", all.Errors)
	}

	var ccUsage FormattedUsageResponse
	if err := json.Unmarshal(all.CommandCode, &ccUsage); err != nil || !ccUsage.OK {
		t.Errorf("commandcode payload invalid: err=%v usage=%+v", err, ccUsage)
	}
	// v0.4.0: the opencode field carries the multi-key envelope.
	var ocUsage OpenCodeMultiKeyResponse
	if err := json.Unmarshal(all.OpenCode, &ocUsage); err != nil || !ocUsage.OK {
		t.Errorf("opencode payload invalid: err=%v usage=%+v", err, ocUsage)
	}
	if len(ocUsage.Keys) != 1 || !ocUsage.Keys[0].OK || ocUsage.Keys[0].Windows == nil {
		t.Errorf("opencode keys = %+v, want one successful key with windows", ocUsage.Keys)
	}
}

// Partial failure: one provider fails upstream → ok stays true, the failed
// provider's field is omitted and the error lands in the errors map.
func TestHandleManagement_AllUsage_PartialFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/usage" && r.Header.Get("Authorization") != "":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"upstream exploded"}`))
		case r.URL.Path == "/internal/billing/credits":
			_, _ = w.Write([]byte(`{"credits":{"monthlyCredits":888},"windowLimits":{"fiveHour":{"used":2,"cap":20}}}`))
		case r.URL.Path == "/internal/usage/summary":
			_, _ = w.Write([]byte(`{"totalMonthlyCredits": 100}`))
		default:
			t.Errorf("unexpected upstream request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())
	defer func() {
		SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})
	}()

	cfg := &PluginConfig{
		SessionToken:    "configured-token",
		APIBase:         ts.URL,
		OpenCodeAPIKey:  "sk-configured",
		OpenCodeAPIBase: ts.URL,
	}

	req := ManagementRequest{
		Method: http.MethodPost,
		Path:   "/plugins/commandcode/all",
	}
	resp, err := HandleManagement(context.Background(), req, cfg)
	if err != nil {
		t.Fatalf("HandleManagement error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200 (partial failure), body=%s", resp.StatusCode, string(resp.Body))
	}

	var all AllUsageResponse
	if err := json.Unmarshal(resp.Body, &all); err != nil {
		t.Fatalf("unmarshal body error: %v", err)
	}
	if !all.OK {
		t.Error("expected ok=true despite one provider failing")
	}
	if len(all.CommandCode) == 0 {
		t.Error("expected successful commandcode payload to be present")
	}
	if _, present := all.Errors["opencode"]; !present {
		t.Errorf("expected errors[opencode] to be set, got %v", all.Errors)
	}
	// v0.4.0: a configured-but-failed key is an upstream failure; with the
	// single configured key failing, the aggregate message is "all N keys failed".
	if !strings.Contains(all.Errors["opencode"], "all 1 opencode keys failed") {
		t.Errorf("errors[opencode] = %q, want it to mention 'all 1 opencode keys failed'", all.Errors["opencode"])
	}
	// opencode field must be omitted (omitempty), not serialized as "null".
	if strings.Contains(string(resp.Body), `"opencode":null`) {
		t.Errorf("opencode field serialized as null: %s", string(resp.Body))
	}
}

// All providers fail because credentials are missing → 400.
func TestHandleManagement_AllUsage_AllMissingConfig(t *testing.T) {
	SetHostCaller(nil)
	SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})

	req := ManagementRequest{
		Method: http.MethodGet,
		Path:   "/v0/management/plugins/commandcode/all",
	}
	resp, err := HandleManagement(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("HandleManagement error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("StatusCode = %d, want 400, body=%s", resp.StatusCode, string(resp.Body))
	}

	var all AllUsageResponse
	if err := json.Unmarshal(resp.Body, &all); err != nil {
		t.Fatalf("unmarshal body error: %v", err)
	}
	if all.OK {
		t.Error("expected ok=false")
	}
	if _, present := all.Errors["commandcode"]; !present {
		t.Errorf("expected errors[commandcode], got %v", all.Errors)
	}
	if _, present := all.Errors["opencode"]; !present {
		t.Errorf("expected errors[opencode], got %v", all.Errors)
	}
}

// Regression for the acceptance review finding: an upstream 400 passed
// through by executeUsageQuery must NOT be classified as a local
// configuration problem — all-upstream-failure must yield 502, not 400.
func TestHandleManagement_AllUsage_Upstream400NotMisclassified(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request from upstream"}`))
	}))
	defer ts.Close()
	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())
	defer SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})

	cfgYAML := []byte("session_token: testtoken\napi_base: " + ts.URL + "\nopencode_api_key: sk-test\nopencode_api_base: " + ts.URL + "\n")
	cfg := NewPlugin()
	if err := cfg.config.UpdateFromYAML(cfgYAML); err != nil {
		t.Fatalf("UpdateFromYAML: %v", err)
	}

	req := ManagementRequest{
		Method: http.MethodGet,
		Path:   "/v0/management/plugins/commandcode/all",
	}
	resp, err := HandleManagement(context.Background(), req, cfg.config)
	if err != nil {
		t.Fatalf("HandleManagement error: %v", err)
	}
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("StatusCode = %d, want 502 (upstream 400 must not be misread as local missing config), body=%s", resp.StatusCode, string(resp.Body))
	}
}

// /all matrix (v0.4.0): Command Code upstream down + all opencode keys 401
// → every failure is upstream → 502, with the aggregate "all N keys failed"
// message in errors["opencode"].
func TestHandleManagement_AllUsage_CCUpstreamDown_OpenCodeAll401(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/usage":
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid api key"}`))
		default: // commandcode internal endpoints
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"cc exploded"}`))
		}
	}))
	defer ts.Close()

	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())
	defer func() {
		SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})
	}()

	cfg := &PluginConfig{
		SessionToken:    "configured-token",
		APIBase:         ts.URL,
		OpenCodeAPIKeys: []string{"sk-bad-AAAA", "sk-bad-BBBB"},
		OpenCodeAPIBase: ts.URL,
	}

	req := ManagementRequest{
		Method: http.MethodGet,
		Path:   "/v0/management/plugins/commandcode/all",
	}
	resp, err := HandleManagement(context.Background(), req, cfg)
	if err != nil {
		t.Fatalf("HandleManagement error: %v", err)
	}
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("StatusCode = %d, want 502 (all failures upstream), body=%s", resp.StatusCode, string(resp.Body))
	}

	var all AllUsageResponse
	if err := json.Unmarshal(resp.Body, &all); err != nil {
		t.Fatalf("unmarshal body error: %v", err)
	}
	if all.OK {
		t.Error("expected ok=false")
	}
	if !strings.Contains(all.Errors["opencode"], "all 2 opencode keys failed") {
		t.Errorf("errors[opencode] = %q, want 'all 2 opencode keys failed'", all.Errors["opencode"])
	}
}

// /all matrix: Command Code upstream down + one opencode key succeeds → 200
// (partial failure); the multi-key opencode payload is inlined.
func TestHandleManagement_AllUsage_CCUpstreamDown_OpenCodeOneOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/usage" && r.Header.Get("Authorization") == "Bearer sk-good-AAAA":
			_, _ = w.Write([]byte(mockOpencodeUsageJSON))
		case r.URL.Path == "/usage":
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid api key"}`))
		default: // commandcode internal endpoints
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"cc exploded"}`))
		}
	}))
	defer ts.Close()

	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())
	defer func() {
		SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})
	}()

	cfg := &PluginConfig{
		SessionToken:    "configured-token",
		APIBase:         ts.URL,
		OpenCodeAPIKeys: []string{"sk-good-AAAA", "sk-bad-BBBB"},
		OpenCodeAPIBase: ts.URL,
	}

	req := ManagementRequest{
		Method: http.MethodGet,
		Path:   "/v0/management/plugins/commandcode/all",
	}
	resp, err := HandleManagement(context.Background(), req, cfg)
	if err != nil {
		t.Fatalf("HandleManagement error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200 (opencode partial success), body=%s", resp.StatusCode, string(resp.Body))
	}

	var all AllUsageResponse
	if err := json.Unmarshal(resp.Body, &all); err != nil {
		t.Fatalf("unmarshal body error: %v", err)
	}
	if !all.OK {
		t.Error("expected ok=true (>=1 provider succeeded)")
	}
	if _, present := all.Errors["commandcode"]; !present {
		t.Errorf("expected errors[commandcode], got %v", all.Errors)
	}
	if _, present := all.Errors["opencode"]; present {
		t.Errorf("errors[opencode] must be absent on partial success, got %q", all.Errors["opencode"])
	}
	var oc OpenCodeMultiKeyResponse
	if err := json.Unmarshal(all.OpenCode, &oc); err != nil || !oc.OK {
		t.Fatalf("opencode payload invalid: err=%v oc=%+v", err, oc)
	}
	if len(oc.Keys) != 2 || !oc.Keys[0].OK || oc.Keys[1].OK {
		t.Errorf("opencode keys = %+v, want [ok, failed]", oc.Keys)
	}
	if strings.Contains(string(resp.Body), "sk-good-AAAA") || strings.Contains(string(resp.Body), "sk-bad-BBBB") {
		t.Errorf("/all response leaks a raw opencode key: %s", string(resp.Body))
	}
}

func TestIsLocalCredentialError(t *testing.T) {
	local := []string{
		"session_token is required. Configure ...",
		"opencode_api_key is required. Configure ...",
		// v0.4.0 prefix: zero opencode keys configured is a local problem.
		"no opencode api keys configured. Configure opencode_api_keys (YAML list) ...",
	}
	for _, msg := range local {
		if !isLocalCredentialError(msg) {
			t.Errorf("isLocalCredentialError(%q) = false, want true", msg)
		}
	}

	upstream := []string{
		"opencode upstream returned 401: check opencode_api_key",
		"opencode upstream request failed: dial tcp: connection refused",
		"all 2 opencode keys failed",
		"upstream returned non-200 status",
		"failed to parse opencode upstream usage: unexpected end of JSON input",
		"",
	}
	for _, msg := range upstream {
		if isLocalCredentialError(msg) {
			t.Errorf("isLocalCredentialError(%q) = true, want false", msg)
		}
	}
}

// Unknown path after the new routes still 404s.
func TestHandleManagement_UnknownPath(t *testing.T) {
	req := ManagementRequest{
		Method: http.MethodGet,
		Path:   "/v0/management/plugins/commandcode/nonsense",
	}
	resp, err := HandleManagement(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("HandleManagement error: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("StatusCode = %d, want 404", resp.StatusCode)
	}
}

// ---- v0.5.0: commandcode_api_key → /alpha + Bearer path ----

// mockAlphaCreditsJSON omits opensourceMonthlyCredits, matching the real
// /alpha/billing/credits payload shape (field difference vs /internal).
const mockAlphaCreditsJSON = `{"credits":{"monthlyCredits":555},"windowLimits":{"fiveHour":{"used":1,"cap":10}}}`

// newAlphaTestServer: /alpha/billing/credits and /alpha/usage/summary both
// assert Bearer auth and the absence of a Cookie header; every other path
// fails the test (regression guard against falling back to /internal).
func newAlphaTestServer(t *testing.T, wantKey string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer "+wantKey {
			t.Errorf("upstream Authorization = %q, want Bearer %s", got, wantKey)
		}
		if got := r.Header.Get("Cookie"); got != "" {
			t.Errorf("upstream Cookie = %q, want none on the /alpha path", got)
		}
		switch r.URL.Path {
		case "/alpha/billing/credits":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(mockAlphaCreditsJSON))
		case "/alpha/usage/summary":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"totalMonthlyCredits": 40}`))
		default:
			t.Errorf("unexpected upstream request: %s %s (internal path must not be hit when commandcode_api_key is set)", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
}

// GET /usage with commandcode_api_key configured → both upstream calls hit
// /alpha/* with Bearer auth and no Cookie; response parses with the alpha
// payload (total = monthly when opensource field is absent) and the monthly
// window is derived from the /alpha summary.
func TestHandleManagement_GetUsage_AlphaKey(t *testing.T) {
	ts := newAlphaTestServer(t, "user_cfg-key")
	defer ts.Close()

	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())
	defer func() {
		SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})
	}()

	cfg := &PluginConfig{
		CommandCodeAPIKey: "user_cfg-key",
		APIBase:           ts.URL,
	}
	req := ManagementRequest{
		Method: http.MethodGet,
		Path:   "/v0/management/plugins/commandcode/usage",
	}
	resp, err := HandleManagement(context.Background(), req, cfg)
	if err != nil {
		t.Fatalf("HandleManagement error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200, body=%s", resp.StatusCode, string(resp.Body))
	}

	var usage FormattedUsageResponse
	if err := json.Unmarshal(resp.Body, &usage); err != nil {
		t.Fatalf("unmarshal body error: %v", err)
	}
	if !usage.OK {
		t.Fatal("expected OK=true")
	}
	if usage.Credits.MonthlyCredits != 555 || usage.Credits.TotalCredits != 555 {
		t.Errorf("credits = %+v, want monthly=555 total=555 (opensource absent in alpha payload)", usage.Credits)
	}
}

// POST /usage: body commandcode_api_key overrides the configured key (the
// config key gets a 401 from the mock, so a 200 proves the body key won).
func TestHandleManagement_PostUsage_AlphaKeyBodyOverridesConfig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get("Authorization") {
		case "Bearer user_body-key":
			switch r.URL.Path {
			case "/alpha/billing/credits":
				_, _ = w.Write([]byte(mockAlphaCreditsJSON))
			case "/alpha/usage/summary":
				_, _ = w.Write([]byte(`{"totalMonthlyCredits": 40}`))
			default:
				t.Errorf("unexpected path: %s", r.URL.Path)
				http.NotFound(w, r)
			}
		default:
			t.Errorf("upstream got Authorization %q — config key must lose to the POST body key", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer ts.Close()

	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())
	defer func() {
		SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})
	}()

	reqBody, _ := json.Marshal(map[string]string{
		"commandcode_api_key": "user_body-key",
		"api_base":            ts.URL,
	})
	req := ManagementRequest{
		Method: http.MethodPost,
		Path:   "/plugins/commandcode/usage",
		Body:   reqBody,
	}
	cfg := &PluginConfig{CommandCodeAPIKey: "user_cfg-must-lose", APIBase: ts.URL}
	resp, err := HandleManagement(context.Background(), req, cfg)
	if err != nil {
		t.Fatalf("HandleManagement error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200 (body key overrides config), body=%s", resp.StatusCode, string(resp.Body))
	}
	var usage FormattedUsageResponse
	if err := json.Unmarshal(resp.Body, &usage); err != nil || !usage.OK {
		t.Errorf("unexpected response: err=%v usage=%+v", err, usage)
	}
}

// Regression: without commandcode_api_key the legacy /internal + Cookie path
// is preserved; the /alpha endpoints must never be requested.
func TestHandleManagement_Usage_FallbackToInternalCookie(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/alpha/") {
			t.Errorf("unexpected /alpha request %s — must stay on /internal without commandcode_api_key", r.URL.Path)
		}
		switch r.URL.Path {
		case "/internal/billing/credits":
			if !strings.Contains(r.Header.Get("Cookie"), "legacy-cookie-token") {
				t.Errorf("Cookie = %q, want the session token cookie", r.Header.Get("Cookie"))
			}
			_, _ = w.Write([]byte(`{"credits":{"monthlyCredits": 321},"windowLimits":{"fiveHour":{"used":1,"cap":10}}}`))
		case "/internal/usage/summary":
			_, _ = w.Write([]byte(`{"totalMonthlyCredits": 11}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())
	defer func() {
		SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})
	}()

	cfg := &PluginConfig{SessionToken: "legacy-cookie-token", APIBase: ts.URL}

	t.Run("GET falls back to internal", func(t *testing.T) {
		req := ManagementRequest{Method: http.MethodGet, Path: "/plugins/commandcode/usage"}
		resp, err := HandleManagement(context.Background(), req, cfg)
		if err != nil {
			t.Fatalf("HandleManagement error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("StatusCode = %d, want 200, body=%s", resp.StatusCode, string(resp.Body))
		}
	})

	t.Run("POST falls back to internal", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]string{"session_token": "legacy-cookie-token", "api_base": ts.URL})
		req := ManagementRequest{Method: http.MethodPost, Path: "/plugins/commandcode/usage", Body: reqBody}
		resp, err := HandleManagement(context.Background(), req, &PluginConfig{})
		if err != nil {
			t.Fatalf("HandleManagement error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("StatusCode = %d, want 200, body=%s", resp.StatusCode, string(resp.Body))
		}
	})
}

// GET must NOT honor a commandcode_api_key query parameter (same
// secrets-out-of-URLs policy as the OpenCode handler): the config's
// session_token path is used and the query-provided key is ignored.
func TestHandleManagement_GetUsage_NoQueryKeyOverride(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/alpha/billing/credits" || r.URL.Path == "/alpha/usage/summary" {
			t.Errorf("unexpected /alpha request %s — query override of commandcode_api_key is not supported", r.URL.Path)
		}
		if r.URL.Path != "/internal/billing/credits" {
			return
		}
		if r.Header.Get("Cookie") == "" {
			t.Errorf("Cookie = %q, want the configured session token cookie", r.Header.Get("Cookie"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"credits":{"monthlyCredits": 77},"windowLimits":{"fiveHour":{"used":1,"cap":10}}}`))
	}))
	defer ts.Close()

	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())
	defer func() {
		SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})
	}()

	cfg := &PluginConfig{SessionToken: "configured-cookie-token", APIBase: ts.URL}
	req := ManagementRequest{
		Method: http.MethodGet,
		Path:   "/v0/management/plugins/commandcode/usage",
		Query: map[string][]string{
			"commandcode_api_key": {"user_query-must-be-ignored"},
		},
	}
	resp, err := HandleManagement(context.Background(), req, cfg)
	if err != nil {
		t.Fatalf("HandleManagement error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200, body=%s", resp.StatusCode, string(resp.Body))
	}
}

// Alpha upstream non-200 → statusCode passed through, message points at
// commandcode_api_key, and the upstream body is NOT echoed (unlike the
// internal path).
func TestHandleManagement_Usage_AlphaUpstreamNon200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"upstream secret detail xyzzy"}`))
	}))
	defer ts.Close()

	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())
	defer func() {
		SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})
	}()

	cfg := &PluginConfig{CommandCodeAPIKey: "user_bad-key", APIBase: ts.URL}
	resp, err := HandleManagement(context.Background(), ManagementRequest{
		Method: http.MethodGet,
		Path:   "/plugins/commandcode/usage",
	}, cfg)
	if err != nil {
		t.Fatalf("HandleManagement error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("StatusCode = %d, want 401 (upstream status passed through)", resp.StatusCode)
	}

	var errResp struct {
		OK         bool   `json:"ok"`
		StatusCode int    `json:"status_code"`
		Error      string `json:"error"`
		Body       string `json:"body"`
	}
	if err := json.Unmarshal(resp.Body, &errResp); err != nil {
		t.Fatalf("unmarshal error body: %v", err)
	}
	if errResp.OK || errResp.StatusCode != http.StatusUnauthorized {
		t.Errorf("error payload = %+v, want ok=false status_code=401", errResp)
	}
	if !strings.Contains(errResp.Error, "commandcode upstream returned 401: check commandcode_api_key") {
		t.Errorf("error = %q, want the commandcode_api_key hint message", errResp.Error)
	}
	if strings.Contains(string(resp.Body), "upstream secret detail") {
		t.Errorf("alpha branch must not echo the upstream body, got: %s", string(resp.Body))
	}
}

// Neither credential → 400 with the preserved "session_token is required"
// prefix (isLocalCredentialError in /all depends on it).
func TestHandleManagement_Usage_NoCredentialsStillSessionTokenMessage(t *testing.T) {
	SetHostCaller(nil)
	SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})
	defer func() {
		SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})
	}()

	resp, err := HandleManagement(context.Background(), ManagementRequest{
		Method: http.MethodGet,
		Path:   "/plugins/commandcode/usage",
	}, &PluginConfig{})
	if err != nil {
		t.Fatalf("HandleManagement error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("StatusCode = %d, want 400, body=%s", resp.StatusCode, string(resp.Body))
	}
	if msg := extractErrorResponseMessage(resp.Body); !strings.HasPrefix(msg, "session_token is required") {
		t.Errorf("error = %q, want the preserved 'session_token is required' prefix", msg)
	}
}

// /all: POST body commandcode_api_key overrides the configured session_token
// (and config key) — the Command Code provider goes through /alpha + Bearer.
func TestHandleManagement_AllUsage_CommandCodeAPIKeyOverride(t *testing.T) {
	ts := newAlphaTestServer(t, "user_all-key")
	defer ts.Close()

	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())
	defer func() {
		SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})
	}()

	cfg := &PluginConfig{
		SessionToken: "cfg-token-must-lose",
		APIBase:      ts.URL,
	}

	reqBody, _ := json.Marshal(map[string]string{"commandcode_api_key": "user_all-key"})
	req := ManagementRequest{
		Method: http.MethodPost,
		Path:   "/plugins/commandcode/all",
		Body:   reqBody,
	}
	resp, err := HandleManagement(context.Background(), req, cfg)
	if err != nil {
		t.Fatalf("HandleManagement error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200 (commandcode succeeded via /alpha), body=%s", resp.StatusCode, string(resp.Body))
	}

	var all AllUsageResponse
	if err := json.Unmarshal(resp.Body, &all); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !all.OK {
		t.Error("expected ok=true")
	}
	var ccUsage FormattedUsageResponse
	if err := json.Unmarshal(all.CommandCode, &ccUsage); err != nil || !ccUsage.OK {
		t.Errorf("commandcode payload invalid: err=%v usage=%+v", err, ccUsage)
	}
	if ccUsage.Credits.TotalCredits != 555 {
		t.Errorf("commandcode total_credits = %v, want 555 (alpha payload)", ccUsage.Credits.TotalCredits)
	}
	// OpenCode key missing → local-credential error for that provider only.
	if _, present := all.Errors["opencode"]; !present {
		t.Errorf("expected errors[opencode], got %v", all.Errors)
	}
}
