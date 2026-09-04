package plugin

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
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
