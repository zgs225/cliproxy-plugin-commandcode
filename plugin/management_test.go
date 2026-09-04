package plugin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterManagement(t *testing.T) {
	resp, err := RegisterManagement()
	if err != nil {
		t.Fatalf("RegisterManagement error: %v", err)
	}

	if len(resp.Routes) != 2 {
		t.Fatalf("len(Routes) = %d, want 2", len(resp.Routes))
	}
	if resp.Routes[0].Method != http.MethodGet || resp.Routes[0].Path != "/plugins/commandcode/usage" {
		t.Errorf("Route 0 mismatch: %+v", resp.Routes[0])
	}
	if resp.Routes[1].Method != http.MethodPost || resp.Routes[1].Path != "/plugins/commandcode/usage" {
		t.Errorf("Route 1 mismatch: %+v", resp.Routes[1])
	}

	if len(resp.Resources) != 1 {
		t.Fatalf("len(Resources) = %d, want 1", len(resp.Resources))
	}
	if resp.Resources[0].Path != "/quota" || resp.Resources[0].Menu != "Command Code 配额" {
		t.Errorf("Resource 0 mismatch: %+v", resp.Resources[0])
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
		if !strings.Contains(bodyStr, "Command Code 配额") {
			t.Errorf("Body does not contain expected title")
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
