package plugin

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultAPIBase = "https://api.commandcode.ai"
)

// HTTPDoer abstracts HTTP requests for testing and fallback.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// HostCallerFunc is the signature for calling the host API via C ABI.
type HostCallerFunc func(method string, payload []byte) ([]byte, error)

var (
	defaultHTTPClient HTTPDoer       = &http.Client{Timeout: 15 * time.Second}
	hostCaller        HostCallerFunc // Set by main if host API is available
)

// SetHostCaller registers the host API callback runner.
func SetHostCaller(fn HostCallerFunc) {
	hostCaller = fn
}

// SetDefaultHTTPClient overrides the default standard HTTP client (useful for unit tests).
func SetDefaultHTTPClient(client HTTPDoer) {
	if client != nil {
		defaultHTTPClient = client
	}
}

// FetchCreditsRaw fetches raw upstream credit data via host.http.do or net/http fallback.
func FetchCreditsRaw(ctx context.Context, apiBase, sessionToken string, hostCallbackID string) ([]byte, int, error) {
	cleanToken := ExtractSessionToken(sessionToken)
	if cleanToken == "" {
		return nil, http.StatusBadRequest, errors.New("missing session_token: please provide a valid Command Code session token")
	}

	if apiBase == "" {
		apiBase = DefaultAPIBase
	}
	url := fmt.Sprintf("%s/internal/billing/credits", strings.TrimRight(apiBase, "/"))
	cookieValue := FormatSessionCookie(cleanToken)

	// 1. Try host.http.do if hostCaller is configured
	if hostCaller != nil {
		reqPayload := HostHTTPRequest{
			Method: http.MethodGet,
			URL:    url,
			Headers: map[string][]string{
				"Cookie":     {cookieValue},
				"Accept":     {"application/json"},
				"User-Agent": {"cliproxy-plugin-commandcode/0.1.0"},
			},
			HostCallbackID: hostCallbackID,
		}
		rawReq, errMarshal := json.Marshal(reqPayload)
		if errMarshal == nil {
			respBytes, errCall := hostCaller("host.http.do", rawReq)
			if errCall == nil && len(respBytes) > 0 {
				var env Envelope
				if errEnv := json.Unmarshal(respBytes, &env); errEnv == nil {
					if !env.OK {
						errMsg := "host HTTP request failed"
						if env.Error != nil {
							errMsg = fmt.Sprintf("%s: %s", env.Error.Code, env.Error.Message)
						}
						return nil, http.StatusBadGateway, fmt.Errorf("host.http.do error: %s", errMsg)
					}
					var hostResp HostHTTPResponse
					if errResp := json.Unmarshal(env.Result, &hostResp); errResp == nil {
						// hostResp.Body is automatically base64-decoded by json.Unmarshal for []byte
						status := hostResp.StatusCode
						if status == 0 {
							status = http.StatusOK
						}
						return hostResp.Body, status, nil
					}
				}
			}
		}
		// If hostCaller fails, seamlessly fallback to net/http
	}

	// 2. Fallback to Go net/http client
	httpReq, errNew := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if errNew != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("create HTTP request: %w", errNew)
	}
	httpReq.Header.Set("Cookie", cookieValue)
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "cliproxy-plugin-commandcode/0.1.0")

	res, errDo := defaultHTTPClient.Do(httpReq)
	if errDo != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("upstream request failed: %w", errDo)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	body, errRead := io.ReadAll(res.Body)
	if errRead != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("read upstream response body: %w", errRead)
	}

	return body, res.StatusCode, nil
}

// ParseAndFormatUsage parses upstream credits JSON into structured usage metrics.
func ParseAndFormatUsage(raw []byte, now time.Time) (*FormattedUsageResponse, error) {
	if len(raw) == 0 {
		return nil, errors.New("empty response body from upstream")
	}

	// First try to parse as UpstreamCreditsResponse
	var upstream UpstreamCreditsResponse
	if err := json.Unmarshal(raw, &upstream); err != nil {
		// Fallback: check if wrapped in { "data": { ... } }
		var wrapped struct {
			Data UpstreamCreditsResponse `json:"data"`
		}
		if errWrap := json.Unmarshal(raw, &wrapped); errWrap == nil && (len(wrapped.Data.Credits) > 0 || wrapped.Data.WindowLimits.FiveHour.Cap > 0) {
			upstream = wrapped.Data
		} else {
			return nil, fmt.Errorf("unmarshal upstream response: %w", err)
		}
	}

	if now.IsZero() {
		now = time.Now().UTC()
	}

	// Format credits
	creditsData := formatCredits(upstream.Credits)

	// Format window limits
	windowLimitsData := formatWindowLimits(upstream.WindowLimits, now)

	nowRFC := now.Format(time.RFC3339)
	data := FormattedUsageData{
		Credits:      creditsData,
		WindowLimits: windowLimitsData,
		UpdatedAt:    nowRFC,
	}

	return &FormattedUsageResponse{
		OK:           true,
		Data:         data,
		Credits:      creditsData,
		WindowLimits: windowLimitsData,
		UpdatedAt:    nowRFC,
	}, nil
}

func formatCredits(credits map[string]any) UsageCreditsData {
	data := UsageCreditsData{
		Details: credits,
	}
	if credits == nil {
		return data
	}

	data.MonthlyCredits = getFloatFromMap(credits, "monthlyCredits", "monthly_credits")
	data.OpensourceMonthlyCredits = getFloatFromMap(credits, "opensourceMonthlyCredits", "opensource_monthly_credits")
	data.TotalCredits = data.MonthlyCredits + data.OpensourceMonthlyCredits

	return data
}

func formatWindowLimits(upstream UpstreamWindowLimits, now time.Time) UsageWindowLimitsData {
	return UsageWindowLimitsData{
		FiveHour: formatSingleWindow(upstream.FiveHour, now),
		Weekly:   formatSingleWindow(upstream.Weekly, now),
	}
}

func formatSingleWindow(w UpstreamWindowLimit, now time.Time) UsageWindowLimitData {
	var remaining float64
	var percentage float64

	if w.Cap > 0 {
		remaining = w.Cap - w.Used
		if remaining < 0 {
			remaining = 0
		}
		percentage = (w.Used / w.Cap) * 100.0
		if percentage > 100.0 {
			percentage = 100.0
		}
		percentage = math.Round(percentage*100) / 100
	}

	var resetAtStr string
	var resetInSeconds int64
	if !w.ResetAt.IsZero() {
		resetAtStr = w.ResetAt.UTC().Format(time.RFC3339)
		diff := w.ResetAt.UTC().Sub(now)
		if diff > 0 {
			resetInSeconds = int64(diff.Seconds())
		} else {
			resetInSeconds = 0
		}
	}

	return UsageWindowLimitData{
		Used:           w.Used,
		Cap:            w.Cap,
		Remaining:      remaining,
		Percentage:     percentage,
		Exceeded:       w.Exceeded,
		ResetAt:        resetAtStr,
		ResetInSeconds: resetInSeconds,
	}
}

func getFloatFromMap(m map[string]any, keys ...string) float64 {
	for _, key := range keys {
		if val, exists := m[key]; exists && val != nil {
			switch v := val.(type) {
			case float64:
				return v
			case float32:
				return float64(v)
			case int:
				return float64(v)
			case int64:
				return float64(v)
			case json.Number:
				if f, err := v.Float64(); err == nil {
					return f
				}
			}
		}
	}
	return 0
}

// DecodeBase64OrRaw tries to decode base64, returning raw if not base64.
func DecodeBase64OrRaw(in []byte) []byte {
	decoded, err := base64.StdEncoding.DecodeString(string(in))
	if err == nil && len(decoded) > 0 {
		return decoded
	}
	return in
}
