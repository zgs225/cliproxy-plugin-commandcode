package plugin

import (
	"context"
	"encoding/json"
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
		_, _ = w.Write([]byte(mockOpencodeUsageJSON))
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
			var usage OpenCodeFormattedUsageResponse
			if err := json.Unmarshal(resp.Body, &usage); err != nil {
				t.Fatalf("unmarshal body error: %v", err)
			}
			if !usage.OK || usage.Provider != "opencode_go" {
				t.Fatalf("unexpected response: ok=%v provider=%q", usage.OK, usage.Provider)
			}
			if usage.Windows.Rolling.Percent != 4 || usage.Windows.Weekly.Percent != 46 || usage.Windows.Monthly.Percent != 23 {
				t.Errorf("windows percents = %v/%v/%v, want 4/46/23",
					usage.Windows.Rolling.Percent, usage.Windows.Weekly.Percent, usage.Windows.Monthly.Percent)
			}
			if usage.Windows.Weekly.ResetInSeconds <= 0 {
				t.Errorf("weekly reset_in_seconds = %d, want > 0", usage.Windows.Weekly.ResetInSeconds)
			}
		})
	}

	if !sawAuthHeader {
		t.Fatal("upstream never received Authorization header")
	}
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
	var ocUsage OpenCodeFormattedUsageResponse
	if err := json.Unmarshal(all.OpenCode, &ocUsage); err != nil || !ocUsage.OK {
		t.Errorf("opencode payload invalid: err=%v usage=%+v", err, ocUsage)
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
	if !strings.Contains(all.Errors["opencode"], "opencode upstream returned 500") {
		t.Errorf("errors[opencode] = %q, want it to mention 'opencode upstream returned 500'", all.Errors["opencode"])
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
