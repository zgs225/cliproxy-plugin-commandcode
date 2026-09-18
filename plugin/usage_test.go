package plugin

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFlexibleTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantYear int
	}{
		{
			name:     "unix seconds",
			input:    `{"resetAt": 1741123456}`,
			wantYear: 2025,
		},
		{
			name:     "unix milliseconds",
			input:    `{"resetAt": 1741123456000}`,
			wantYear: 2025,
		},
		{
			name:     "RFC3339 string",
			input:    `{"resetAt": "2025-06-15T12:00:00Z"}`,
			wantYear: 2025,
		},
		{
			name:     "null",
			input:    `{"resetAt": null}`,
			wantYear: 1, // Zero time year
		},
		{
			name:     "empty string",
			input:    `{"resetAt": ""}`,
			wantYear: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res struct {
				ResetAt FlexibleTime `json:"resetAt"`
			}
			if err := json.Unmarshal([]byte(tt.input), &res); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}
			if res.ResetAt.Year() != tt.wantYear {
				t.Errorf("Year = %d, want %d", res.ResetAt.Year(), tt.wantYear)
			}
		})
	}
}

func TestParseAndFormatUsage(t *testing.T) {
	raw := []byte(`{
		"credits": {
			"monthlyCredits": 1000.0,
			"opensourceMonthlyCredits": 500.0,
			"extraBonus": 50.0
		},
		"windowLimits": {
			"fiveHour": {
				"used": 25.0,
				"cap": 100.0,
				"exceeded": false,
				"resetAt": 1741123456
			},
			"weekly": {
				"used": 200.0,
				"cap": 1000.0,
				"exceeded": false,
				"resetAt": "2025-03-10T12:00:00Z"
			}
		}
	}`)

	now := time.Date(2025, 3, 4, 12, 0, 0, 0, time.UTC)

	t.Run("without summary monthly defaults empty", func(t *testing.T) {
		usage, err := ParseAndFormatUsage(raw, nil, now)
		if err != nil {
			t.Fatalf("ParseAndFormatUsage error: %v", err)
		}
		if !usage.OK {
			t.Fatal("expected OK=true")
		}
		if usage.WindowLimits.Monthly.Used != 0 || usage.WindowLimits.Monthly.Cap != 0 {
			t.Errorf("expected empty monthly without summary, got %+v", usage.WindowLimits.Monthly)
		}
	})

	t.Run("monthly window derived from summary + remaining credits", func(t *testing.T) {
		summary := &UpstreamUsageSummaryResponse{TotalMonthlyCredits: 30.0}
		// monthlyCredits in raw = 1000 remaining, so cap = 1030
		usage, err := ParseAndFormatUsage(raw, summary, now)
		if err != nil {
			t.Fatalf("ParseAndFormatUsage error: %v", err)
		}
		m := usage.WindowLimits.Monthly
		if m.Used != 30.0 {
			t.Errorf("monthly used = %v, want 30", m.Used)
		}
		if m.Cap != 1030.0 {
			t.Errorf("monthly cap = %v, want 1030", m.Cap)
		}
		if m.Remaining != 1000.0 {
			t.Errorf("monthly remaining = %v, want 1000", m.Remaining)
		}
		wantPct := math.Round((30.0/1030.0)*10000) / 100
		if m.Percentage != wantPct {
			t.Errorf("monthly percentage = %v, want %v", m.Percentage, wantPct)
		}
	})

	usage, err := ParseAndFormatUsage(raw, nil, now)
	if err != nil {
		t.Fatalf("ParseAndFormatUsage error: %v", err)
	}

	if !usage.OK {
		t.Fatal("expected OK=true")
	}

	// Verify credits
	if usage.Credits.MonthlyCredits != 1000.0 {
		t.Errorf("MonthlyCredits = %v, want 1000", usage.Credits.MonthlyCredits)
	}
	if usage.Credits.OpensourceMonthlyCredits != 500.0 {
		t.Errorf("OpensourceMonthlyCredits = %v, want 500", usage.Credits.OpensourceMonthlyCredits)
	}
	if usage.Credits.TotalCredits != 1500.0 {
		t.Errorf("TotalCredits = %v, want 1500", usage.Credits.TotalCredits)
	}

	// Verify 5-hour window
	fiveHour := usage.WindowLimits.FiveHour
	if fiveHour.Used != 25.0 {
		t.Errorf("FiveHour Used = %v, want 25", fiveHour.Used)
	}
	if fiveHour.Cap != 100.0 {
		t.Errorf("FiveHour Cap = %v, want 100", fiveHour.Cap)
	}
	if fiveHour.Remaining != 75.0 {
		t.Errorf("FiveHour Remaining = %v, want 75", fiveHour.Remaining)
	}
	if fiveHour.Percentage != 25.0 {
		t.Errorf("FiveHour Percentage = %v, want 25", fiveHour.Percentage)
	}
	if fiveHour.Exceeded {
		t.Errorf("FiveHour Exceeded = true, want false")
	}

	// Verify weekly window
	weekly := usage.WindowLimits.Weekly
	if weekly.Used != 200.0 {
		t.Errorf("Weekly Used = %v, want 200", weekly.Used)
	}
	if weekly.Cap != 1000.0 {
		t.Errorf("Weekly Cap = %v, want 1000", weekly.Cap)
	}
	if weekly.Remaining != 800.0 {
		t.Errorf("Weekly Remaining = %v, want 800", weekly.Remaining)
	}
	if weekly.Percentage != 20.0 {
		t.Errorf("Weekly Percentage = %v, want 20", weekly.Percentage)
	}
	if weekly.ResetAt != "2025-03-10T12:00:00Z" {
		t.Errorf("Weekly ResetAt = %v, want 2025-03-10T12:00:00Z", weekly.ResetAt)
	}

	// Verify inferred plan (cap 25 / 100 is unknown)
	if usage.Plan.Name != "Unknown" || usage.Plan.Code != "unknown" {
		t.Errorf("Plan = %+v, want Unknown", usage.Plan)
	}

	// Verify plan inference for GOAT
	rawGOAT := []byte(`{
		"credits": {"monthlyCredits": 100},
		"windowLimits": {
			"fiveHour": {"used": 2, "cap": 14},
			"weekly": {"used": 10, "cap": 35}
		}
	}`)
	usageGOAT, errGOAT := ParseAndFormatUsage(rawGOAT, nil, now)
	if errGOAT != nil {
		t.Fatalf("ParseAndFormatUsage GOAT error: %v", errGOAT)
	}
	if usageGOAT.Plan.Name != "GOAT" || usageGOAT.Plan.Code != "goat" {
		t.Errorf("GOAT Plan = %+v, want name=GOAT code=goat", usageGOAT.Plan)
	}
}

func TestFetchCreditsRaw_FallbackHTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/billing/credits" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		cookie := r.Header.Get("Cookie")
		expectedCookie := "__Secure-commandcode_prod_.session_token=test-session-123"
		if cookie != expectedCookie {
			t.Errorf("Cookie = %q, want %q", cookie, expectedCookie)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"credits":{"monthlyCredits":100},"windowLimits":{"fiveHour":{"used":1,"cap":10}}}`))
	}))
	defer ts.Close()

	// Ensure hostCaller is nil for fallback test
	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())

	body, status, err := FetchCreditsRaw(context.Background(), ts.URL, "test-session-123", "")
	if err != nil {
		t.Fatalf("FetchCreditsRaw error: %v", err)
	}
	if status != http.StatusOK {
		t.Errorf("status = %d, want 200", status)
	}
	if len(body) == 0 {
		t.Fatal("expected non-empty body")
	}

	usage, errParse := ParseAndFormatUsage(body, nil, time.Time{})
	if errParse != nil {
		t.Fatalf("ParseAndFormatUsage error: %v", errParse)
	}
	if usage.Credits.MonthlyCredits != 100 {
		t.Errorf("MonthlyCredits = %v, want 100", usage.Credits.MonthlyCredits)
	}
}

func TestFetchCreditsRaw_MissingToken(t *testing.T) {
	_, status, err := FetchCreditsRaw(context.Background(), "", "", "")
	if err == nil {
		t.Fatal("expected error for missing token")
	}
	if status != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", status)
	}
}

func TestFetchCreditsRaw_HostCaller(t *testing.T) {
	mockResponsePayload := []byte(`{"credits":{"monthlyCredits":500},"windowLimits":{"fiveHour":{"used":5,"cap":50}}}`)

	SetHostCaller(func(method string, payload []byte) ([]byte, error) {
		if method != "host.http.do" {
			t.Errorf("method = %s, want host.http.do", method)
		}
		hostResp := HostHTTPResponse{
			StatusCode: http.StatusOK,
			Body:       mockResponsePayload,
		}
		respJSON, _ := json.Marshal(hostResp)
		return json.Marshal(Envelope{OK: true, Result: respJSON})
	})
	defer SetHostCaller(nil)

	body, status, err := FetchCreditsRaw(context.Background(), "https://api.commandcode.ai", "mock-token", "cb-123")
	if err != nil {
		t.Fatalf("FetchCreditsRaw with hostCaller error: %v", err)
	}
	if status != http.StatusOK {
		t.Errorf("status = %d, want 200", status)
	}
	if string(body) != string(mockResponsePayload) {
		t.Errorf("body = %s, want %s", string(body), string(mockResponsePayload))
	}
}

func TestPlanFromWindowLimits(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name        string
		fiveHourCap float64
		weeklyCap   float64
		limited     *bool
		wantName    string
		wantCode    string
	}{
		{
			name:        "GOAT plan",
			fiveHourCap: 14,
			weeklyCap:   35,
			limited:     &trueVal,
			wantName:    "GOAT",
			wantCode:    "goat",
		},
		{
			name:        "Pro plan",
			fiveHourCap: 16,
			weeklyCap:   40,
			limited:     nil,
			wantName:    "Pro",
			wantCode:    "pro",
		},
		{
			name:        "Max 20x plan",
			fiveHourCap: 90,
			weeklyCap:   180,
			limited:     &trueVal,
			wantName:    "Max 20×",
			wantCode:    "max_20x",
		},
		{
			name:        "Max 10x plan",
			fiveHourCap: 45,
			weeklyCap:   90,
			limited:     nil,
			wantName:    "Max 10×",
			wantCode:    "max_10x",
		},
		{
			name:        "Go plan",
			fiveHourCap: 3,
			weeklyCap:   6,
			limited:     nil,
			wantName:    "Go",
			wantCode:    "go",
		},
		{
			name:        "Provider pay-as-you-go plan",
			fiveHourCap: 0,
			weeklyCap:   0,
			limited:     &falseVal,
			wantName:    "Provider",
			wantCode:    "provider",
		},
		{
			name:        "Provider plan with caps set but limited=false",
			fiveHourCap: 14,
			weeklyCap:   35,
			limited:     &falseVal,
			wantName:    "Provider",
			wantCode:    "provider",
		},
		{
			name:        "Float tolerance test",
			fiveHourCap: 13.999,
			weeklyCap:   35.001,
			limited:     nil,
			wantName:    "GOAT",
			wantCode:    "goat",
		},
		{
			name:        "Unknown caps",
			fiveHourCap: 10,
			weeklyCap:   20,
			limited:     nil,
			wantName:    "Unknown",
			wantCode:    "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PlanFromWindowLimits(tt.fiveHourCap, tt.weeklyCap, tt.limited)
			if got.Name != tt.wantName || got.Code != tt.wantCode {
				t.Errorf("PlanFromWindowLimits(%v, %v, %v) = %+v, want name=%q code=%q",
					tt.fiveHourCap, tt.weeklyCap, tt.limited, got, tt.wantName, tt.wantCode)
			}
		})
	}
}

func TestParseOpenCodeUsage(t *testing.T) {
	now := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)

	t.Run("normal payload from real upstream shape", func(t *testing.T) {
		raw := []byte(`{"usage":{
			"rolling": {"status":"ok","percent":4, "resetsAt":"2026-09-17T06:58:53.171Z"},
			"weekly":  {"status":"ok","percent":46,"resetsAt":"2026-09-21T00:00:00.000Z"},
			"monthly": {"status":"ok","percent":23,"resetsAt":"2026-10-14T09:13:49.000Z"}
		}}`)

		usage, err := ParseOpenCodeUsage(raw, now)
		if err != nil {
			t.Fatalf("ParseOpenCodeUsage error: %v", err)
		}
		if !usage.OK || usage.Provider != "opencode_go" {
			t.Fatalf("unexpected header: ok=%v provider=%q", usage.OK, usage.Provider)
		}
		if usage.UpdatedAt != "2026-09-16T12:00:00Z" {
			t.Errorf("UpdatedAt = %q", usage.UpdatedAt)
		}

		rolling := usage.Windows.Rolling
		if rolling.Percent != 4 || rolling.Status != "ok" || rolling.Exceeded {
			t.Errorf("rolling = %+v", rolling)
		}
		if rolling.ResetAt != "2026-09-17T06:58:53Z" {
			t.Errorf("rolling reset_at = %q", rolling.ResetAt)
		}
		if rolling.ResetInSeconds != 68333 {
			t.Errorf("rolling reset_in_seconds = %d, want 68333", rolling.ResetInSeconds)
		}

		weekly := usage.Windows.Weekly
		if weekly.Percent != 46 {
			t.Errorf("weekly percent = %v, want 46", weekly.Percent)
		}
		if weekly.ResetAt != "2026-09-21T00:00:00Z" {
			t.Errorf("weekly reset_at = %q, want 2026-09-21T00:00:00Z (.000Z tolerated)", weekly.ResetAt)
		}

		monthly := usage.Windows.Monthly
		if monthly.Percent != 23 {
			t.Errorf("monthly percent = %v, want 23", monthly.Percent)
		}
	})

	t.Run("float percent", func(t *testing.T) {
		raw := []byte(`{"usage":{"rolling":{"status":"ok","percent":12.345,"resetsAt":"2026-09-17T06:58:53Z"}}}`)
		usage, err := ParseOpenCodeUsage(raw, now)
		if err != nil {
			t.Fatalf("ParseOpenCodeUsage error: %v", err)
		}
		if got := usage.Windows.Rolling.Percent; got != 12.35 { // Round(x*100)/100
			t.Errorf("percent = %v, want 12.35", got)
		}
	})

	t.Run("unknown status tolerated", func(t *testing.T) {
		raw := []byte(`{"usage":{"rolling":{"status":"weird-status","percent":50,"resetsAt":"2026-09-17T06:58:53Z"}}}`)
		usage, err := ParseOpenCodeUsage(raw, now)
		if err != nil {
			t.Fatalf("ParseOpenCodeUsage error: %v", err)
		}
		if got := usage.Windows.Rolling; got.Status != "weird-status" || got.Exceeded {
			t.Errorf("rolling = %+v, want status kept and not exceeded", got)
		}
	})

	t.Run("exceeded status", func(t *testing.T) {
		raw := []byte(`{"usage":{"rolling":{"status":"exceeded","percent":99,"resetsAt":"2026-09-17T06:58:53Z"}}}`)
		usage, err := ParseOpenCodeUsage(raw, now)
		if err != nil {
			t.Fatalf("ParseOpenCodeUsage error: %v", err)
		}
		if !usage.Windows.Rolling.Exceeded {
			t.Error("expected Exceeded=true for status=exceeded")
		}
	})

	t.Run("percent 100 exceeded", func(t *testing.T) {
		raw := []byte(`{"usage":{"rolling":{"status":"ok","percent":100,"resetsAt":""}}}`)
		usage, err := ParseOpenCodeUsage(raw, now)
		if err != nil {
			t.Fatalf("ParseOpenCodeUsage error: %v", err)
		}
		if !usage.Windows.Rolling.Exceeded {
			t.Error("expected Exceeded=true for percent=100")
		}
		if usage.Windows.Rolling.ResetAt != "" || usage.Windows.Rolling.ResetInSeconds != 0 {
			t.Errorf("expected empty reset fields, got %+v", usage.Windows.Rolling)
		}
	})

	t.Run("percent above 100 clamped", func(t *testing.T) {
		raw := []byte(`{"usage":{"rolling":{"status":"ok","percent":150.5,"resetsAt":""}}}`)
		usage, err := ParseOpenCodeUsage(raw, now)
		if err != nil {
			t.Fatalf("ParseOpenCodeUsage error: %v", err)
		}
		if got := usage.Windows.Rolling.Percent; got != 100 {
			t.Errorf("percent = %v, want 100 (clamped)", got)
		}
		if !usage.Windows.Rolling.Exceeded {
			t.Error("expected Exceeded=true when clamped to 100")
		}
	})

	t.Run("malformed resetsAt not fatal", func(t *testing.T) {
		raw := []byte(`{"usage":{"rolling":{"status":"ok","percent":5,"resetsAt":"not-a-timestamp"}}}`)
		usage, err := ParseOpenCodeUsage(raw, now)
		if err != nil {
			t.Fatalf("ParseOpenCodeUsage must not fail on bad resetsAt: %v", err)
		}
		if got := usage.Windows.Rolling; got.ResetAt != "" || got.ResetInSeconds != 0 {
			t.Errorf("expected zero reset fields on parse failure, got %+v", got)
		}
	})

	t.Run("missing windows tolerated as zero values", func(t *testing.T) {
		raw := []byte(`{"usage":{}}`)
		usage, err := ParseOpenCodeUsage(raw, now)
		if err != nil {
			t.Fatalf("ParseOpenCodeUsage error: %v", err)
		}
		if usage.Windows.Rolling.Percent != 0 {
			t.Errorf("rolling percent = %v, want 0", usage.Windows.Rolling.Percent)
		}
	})

	t.Run("empty body", func(t *testing.T) {
		if _, err := ParseOpenCodeUsage(nil, now); err == nil {
			t.Fatal("expected error for empty body")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		if _, err := ParseOpenCodeUsage([]byte(`not-json`), now); err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestFetchOpenCodeUsageRaw_FallbackHTTP(t *testing.T) {
	var sawAuth, sawUA, sawAccept string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/usage" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		sawAuth = r.Header.Get("Authorization")
		sawUA = r.Header.Get("User-Agent")
		sawAccept = r.Header.Get("Accept")

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"usage":{"rolling":{"status":"ok","percent":4,"resetsAt":"2026-09-17T06:58:53.171Z"}}}`))
	}))
	defer ts.Close()

	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())
	defer func() {
		SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})
	}()

	body, status, err := FetchOpenCodeUsageRaw(context.Background(), ts.URL, "sk-test-key", "")
	if err != nil {
		t.Fatalf("FetchOpenCodeUsageRaw error: %v", err)
	}
	if status != http.StatusOK {
		t.Errorf("status = %d, want 200", status)
	}
	if len(body) == 0 {
		t.Fatal("expected non-empty body")
	}
	if sawAuth != "Bearer sk-test-key" {
		t.Errorf("Authorization = %q, want Bearer sk-test-key", sawAuth)
	}
	if sawAccept != "application/json" {
		t.Errorf("Accept = %q, want application/json", sawAccept)
	}
	if !strings.Contains(sawUA, "cliproxy-plugin-commandcode/") {
		t.Errorf("User-Agent = %q, want cliproxy-plugin-commandcode/<version>", sawUA)
	}

	usage, errParse := ParseOpenCodeUsage(body, time.Time{})
	if errParse != nil {
		t.Fatalf("ParseOpenCodeUsage error: %v", errParse)
	}
	if usage.Windows.Rolling.Percent != 4 {
		t.Errorf("rolling percent = %v, want 4", usage.Windows.Rolling.Percent)
	}
}

func TestFetchOpenCodeUsageRaw_BaseTrailingSlash(t *testing.T) {
	requests := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/usage" {
			t.Errorf("path = %q, want /usage (trailing slash trimmed)", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"usage":{"rolling":{"status":"ok","percent":1,"resetsAt":""}}}`))
	}))
	defer ts.Close()

	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())
	defer func() {
		SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})
	}()

	if _, _, err := FetchOpenCodeUsageRaw(context.Background(), ts.URL+"/", "sk-key", ""); err != nil {
		t.Fatalf("error: %v", err)
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want 1", requests)
	}
}

func TestFetchOpenCodeUsageRaw_MissingKey(t *testing.T) {
	_, status, err := FetchOpenCodeUsageRaw(context.Background(), "", "", "")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if status != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", status)
	}
}

func TestFetchOpenCodeUsageRaw_EmptyKeyAfterTrim(t *testing.T) {
	_, status, err := FetchOpenCodeUsageRaw(context.Background(), "", "   ", "")
	if err == nil {
		t.Fatal("expected error for whitespace-only key")
	}
	if status != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", status)
	}
}

func TestFetchOpenCodeUsageRaw_UpstreamNon200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid api key"}`))
	}))
	defer ts.Close()

	SetHostCaller(nil)
	SetDefaultHTTPClient(ts.Client())
	defer func() {
		SetDefaultHTTPClient(&http.Client{Timeout: 15 * time.Second})
	}()

	body, status, err := FetchOpenCodeUsageRaw(context.Background(), ts.URL, "sk-bad", "")
	if err != nil {
		t.Fatalf("expected nil transport error for non-200 upstream, got %v", err)
	}
	if status != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", status)
	}
	if string(body) != `{"error":"invalid api key"}` {
		t.Errorf("body = %q", string(body))
	}
}

func TestFetchOpenCodeUsageRaw_HostCaller(t *testing.T) {
	mockResponsePayload := []byte(`{"usage":{"rolling":{"status":"ok","percent":7,"resetsAt":"2026-09-17T06:58:53Z"}}}`)

	var sawMethod, sawURL string
	var sawHeaders map[string][]string
	SetHostCaller(func(method string, payload []byte) ([]byte, error) {
		if method != "host.http.do" {
			t.Errorf("method = %s, want host.http.do", method)
		}
		var req HostHTTPRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			t.Fatalf("unmarshal HostHTTPRequest error: %v", err)
		}
		sawMethod, sawURL, sawHeaders = req.Method, req.URL, req.Headers
		hostResp := HostHTTPResponse{
			StatusCode: http.StatusOK,
			Body:       mockResponsePayload,
		}
		respJSON, _ := json.Marshal(hostResp)
		return json.Marshal(Envelope{OK: true, Result: respJSON})
	})
	defer SetHostCaller(nil)

	body, status, err := FetchOpenCodeUsageRaw(context.Background(), "https://opencode.example/v1", "sk-host-key", "cb-123")
	if err != nil {
		t.Fatalf("FetchOpenCodeUsageRaw with hostCaller error: %v", err)
	}
	if status != http.StatusOK {
		t.Errorf("status = %d, want 200", status)
	}
	if string(body) != string(mockResponsePayload) {
		t.Errorf("body = %s, want %s", string(body), string(mockResponsePayload))
	}
	if sawMethod != http.MethodGet {
		t.Errorf("host request method = %s, want GET", sawMethod)
	}
	if sawURL != "https://opencode.example/v1/usage" {
		t.Errorf("host request url = %s, want https://opencode.example/v1/usage", sawURL)
	}
	auth := sawHeaders["Authorization"]
	if len(auth) == 0 || auth[0] != "Bearer sk-host-key" {
		t.Errorf("host request Authorization = %v, want Bearer sk-host-key", auth)
	}
}
