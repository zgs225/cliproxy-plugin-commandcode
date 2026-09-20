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
	DefaultAPIBase         = "https://api.commandcode.ai"
	DefaultOpenCodeAPIBase = "https://opencode.ai/zen/go/v1"
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

// fetchUpstream performs a GET on an internal Command Code endpoint, reusing
// the host.http.do bridge when available, else falling back to net/http.
func fetchUpstream(ctx context.Context, apiBase, endpoint, sessionToken, hostCallbackID string) ([]byte, int, error) {
	cleanToken := ExtractSessionToken(sessionToken)
	if cleanToken == "" {
		return nil, http.StatusBadRequest, errors.New("missing session_token: please provide a valid Command Code session token")
	}

	if apiBase == "" {
		apiBase = DefaultAPIBase
	}
	url := fmt.Sprintf("%s/%s", strings.TrimRight(apiBase, "/"), strings.TrimLeft(endpoint, "/"))
	cookieValue := FormatSessionCookie(cleanToken)

	headers := map[string][]string{
		"Cookie":     {cookieValue},
		"Accept":     {"application/json"},
		"User-Agent": {fmt.Sprintf("cliproxy-plugin-commandcode/%s", PluginVersion)},
	}
	return doUpstreamRequest(ctx, http.MethodGet, url, headers, hostCallbackID)
}

// doUpstreamRequest is the shared transport layer: it tries the host.http.do
// bridge first (when a host caller is registered) and falls back to net/http.
// Request semantics (method, URL, headers) are fully controlled by the caller.
func doUpstreamRequest(ctx context.Context, method, url string, headers map[string][]string, hostCallbackID string) ([]byte, int, error) {
	// 1. Try host.http.do if hostCaller is configured
	if hostCaller != nil {
		reqPayload := HostHTTPRequest{
			Method:         method,
			URL:            url,
			Headers:        headers,
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
	httpReq, errNew := http.NewRequestWithContext(ctx, method, url, nil)
	if errNew != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("create HTTP request: %w", errNew)
	}
	for key, values := range headers {
		for _, value := range values {
			httpReq.Header.Add(key, value)
		}
	}

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

// FetchCreditsRaw fetches raw upstream credit data via host.http.do or net/http fallback.
func FetchCreditsRaw(ctx context.Context, apiBase, sessionToken string, hostCallbackID string) ([]byte, int, error) {
	return fetchUpstream(ctx, apiBase, "internal/billing/credits", sessionToken, hostCallbackID)
}

// FetchUsageSummaryRaw fetches the billing-period (monthly) usage totals.
func FetchUsageSummaryRaw(ctx context.Context, apiBase, sessionToken string, hostCallbackID string) ([]byte, int, error) {
	return fetchUpstream(ctx, apiBase, "internal/usage/summary", sessionToken, hostCallbackID)
}

// FetchOpenCodeUsageRaw fetches raw OpenCode Go usage data from
// {apiBase}/usage with Bearer auth, via host.http.do or net/http fallback.
func FetchOpenCodeUsageRaw(ctx context.Context, apiBase, apiKey, hostCallbackID string) ([]byte, int, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, http.StatusBadRequest, errors.New("missing opencode_api_key: configure opencode_api_key in plugin config or pass it in the request")
	}

	if apiBase == "" {
		apiBase = DefaultOpenCodeAPIBase
	}
	url := strings.TrimRight(apiBase, "/") + "/usage"

	headers := map[string][]string{
		"Authorization": {"Bearer " + apiKey},
		"Accept":        {"application/json"},
		"User-Agent":    {fmt.Sprintf("cliproxy-plugin-commandcode/%s", PluginVersion)},
	}
	return doUpstreamRequest(ctx, http.MethodGet, url, headers, hostCallbackID)
}

// ParseOpenCodeUsage parses OpenCode Go usage JSON into the formatted response.
// Unknown status values are tolerated; a resetsAt that fails to parse is not
// fatal (ResetAt stays empty and ResetInSeconds stays 0).
func ParseOpenCodeUsage(raw []byte, now time.Time) (*OpenCodeFormattedUsageResponse, error) {
	if len(raw) == 0 {
		return nil, errors.New("empty response body from upstream")
	}

	var upstream OpenCodeUsageResponse
	if err := json.Unmarshal(raw, &upstream); err != nil {
		return nil, fmt.Errorf("unmarshal opencode usage response: %w", err)
	}

	if now.IsZero() {
		now = time.Now().UTC()
	}

	return &OpenCodeFormattedUsageResponse{
		OK:       true,
		Provider: "opencode_go",
		Windows: OpenCodeFormattedWindows{
			Rolling: formatOpenCodeWindow(upstream.Usage.Rolling, now),
			Weekly:  formatOpenCodeWindow(upstream.Usage.Weekly, now),
			Monthly: formatOpenCodeWindow(upstream.Usage.Monthly, now),
		},
		UpdatedAt: now.Format(time.RFC3339),
	}, nil
}

// formatOpenCodeWindow formats a single OpenCode Go usage window.
func formatOpenCodeWindow(w OpenCodeUsageWindow, now time.Time) OpenCodeFormattedWindow {
	percent := clampOpenCodePercent(w.Percent)
	out := OpenCodeFormattedWindow{
		Status:   w.Status,
		Percent:  percent,
		Exceeded: percent >= 100 || w.Status == "exceeded",
	}

	if w.ResetsAt != "" {
		if t, err := time.Parse(time.RFC3339, w.ResetsAt); err == nil {
			out.ResetAt = t.UTC().Format(time.RFC3339)
			if diff := t.UTC().Sub(now); diff > 0 {
				out.ResetInSeconds = int64(diff.Seconds())
			}
		}
		// Parse failure is not fatal: ResetAt stays empty, ResetInSeconds stays 0.
	}
	return out
}

// clampOpenCodePercent clamps a percentage to [0, 100] with 2-decimal rounding.
func clampOpenCodePercent(p float64) float64 {
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	return math.Round(p*100) / 100
}

// MaskAPIKey masks an OpenCode Go API key for display: first 4 + "…" + last 4
// characters (e.g. "sk-L…KqYB"). Keys shorter than 8 characters are fully
// masked as "***"; an empty key masks to "".
func MaskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) < 8 {
		return "***"
	}
	return key[:4] + "…" + key[len(key)-4:]
}

// QueryOpenCodeKeys queries OpenCode Go usage for each key sequentially and
// returns one typed result per key, in input order. A single key's failure is
// recorded only in that key's result and never aborts the loop.
func QueryOpenCodeKeys(ctx context.Context, apiBase string, keys []string, hostCallbackID string) []OpenCodeKeyResult {
	results := make([]OpenCodeKeyResult, 0, len(keys))
	for _, key := range keys {
		res, _ := queryOpenCodeKey(ctx, apiBase, key, hostCallbackID)
		results = append(results, res)
	}
	return results
}

// ParseAndFormatUsage parses upstream credits JSON into structured usage metrics.
// summary (optional) carries the billing-period usage totals used to derive the monthly window.
func ParseAndFormatUsage(raw []byte, summary *UpstreamUsageSummaryResponse, now time.Time) (*FormattedUsageResponse, error) {
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
	windowLimitsData.Monthly = formatMonthlyWindow(upstream.Credits, summary, now)

	// Inferred subscription plan
	planInfo := PlanFromWindowLimits(upstream.WindowLimits.FiveHour.Cap, upstream.WindowLimits.Weekly.Cap, upstream.WindowLimits.Limited)

	nowRFC := now.Format(time.RFC3339)
	data := FormattedUsageData{
		Plan:         planInfo,
		Credits:      creditsData,
		WindowLimits: windowLimitsData,
		UpdatedAt:    nowRFC,
	}

	return &FormattedUsageResponse{
		OK:           true,
		Plan:         planInfo,
		Data:         data,
		Credits:      creditsData,
		WindowLimits: windowLimitsData,
		UpdatedAt:    nowRFC,
	}, nil
}

// PlanFromWindowLimits infers the Command Code subscription plan based on window limit caps.
//
// Rules:
//   - windowLimits.limited == false -> Provider plan (pay-as-you-go)
//   - 5h cap=14 && weekly cap=35 -> GOAT plan
//   - 5h cap=16 && weekly cap=40 -> Pro plan
//   - 5h cap=90 && weekly cap=180 -> Max 20× plan
//   - 5h cap=45 && weekly cap=90 -> Max 10× plan
//   - 5h cap=3 && weekly cap=6 -> Go plan
//   - Otherwise -> Unknown
func PlanFromWindowLimits(fiveHourCap, weeklyCap float64, limited *bool) PlanInfo {
	if limited != nil && !*limited {
		return PlanInfo{
			Name: "Provider",
			Code: "provider",
		}
	}

	match := func(capVal, target float64) bool {
		return math.Abs(capVal-target) < 0.01
	}

	if match(fiveHourCap, 14) && match(weeklyCap, 35) {
		return PlanInfo{
			Name: "GOAT",
			Code: "goat",
		}
	}

	if match(fiveHourCap, 16) && match(weeklyCap, 40) {
		return PlanInfo{
			Name: "Pro",
			Code: "pro",
		}
	}

	if match(fiveHourCap, 90) && match(weeklyCap, 180) {
		return PlanInfo{
			Name: "Max 20×",
			Code: "max_20x",
		}
	}

	if match(fiveHourCap, 45) && match(weeklyCap, 90) {
		return PlanInfo{
			Name: "Max 10×",
			Code: "max_10x",
		}
	}

	if match(fiveHourCap, 3) && match(weeklyCap, 6) {
		return PlanInfo{
			Name: "Go",
			Code: "go",
		}
	}

	return PlanInfo{
		Name: "Unknown",
		Code: "unknown",
	}
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

// formatMonthlyWindow derives the monthly (billing period) window from the
// billing/credits response (remaining monthlyCredits) and the
// /internal/usage/summary response (totalMonthlyCredits consumed this period).
// cap = consumed + remaining, used = consumed, remaining = monthlyCredits.
// Returns a zero window when the summary (consumed totals) is unavailable,
// because a monthly window cannot be derived from remaining credits alone.
func formatMonthlyWindow(credits map[string]any, summary *UpstreamUsageSummaryResponse, now time.Time) UsageWindowLimitData {
	out := UsageWindowLimitData{}
	if summary == nil || summary.TotalMonthlyCredits <= 0 {
		return out
	}

	remaining := getFloatFromMap(credits, "monthlyCredits", "monthly_credits")
	used := summary.TotalMonthlyCredits

	capTotal := used + remaining
	if capTotal > 0 {
		percentage := (used / capTotal) * 100.0
		if percentage > 100.0 {
			percentage = 100.0
		}
		out.Percentage = math.Round(percentage*100) / 100
	}

	out.Used = used
	out.Cap = capTotal
	out.Remaining = remaining
	if out.Remaining < 0 {
		out.Remaining = 0
	}
	out.ResetAt = "" // 账单周期重置时间上游未提供
	out.ResetInSeconds = 0
	return out
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
