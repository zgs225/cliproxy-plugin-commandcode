package plugin

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Envelope matches the CLIProxyAPI ABI JSON Envelope.
type Envelope struct {
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *EnvelopeError  `json:"error,omitempty"`
}

// EnvelopeError represents an error inside the ABI Envelope.
type EnvelopeError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Retryable  bool   `json:"retryable,omitempty"`
	HTTPStatus int    `json:"http_status,omitempty"`
}

// LifecycleRequest represents the payload for plugin.register or plugin.reconfigure.
type LifecycleRequest struct {
	ConfigYAML []byte `json:"config_yaml"`
}

// Registration describes the plugin registration response.
type Registration struct {
	SchemaVersion uint32                 `json:"schema_version"`
	Metadata      Metadata               `json:"metadata"`
	Capabilities  RegistrationCapability `json:"capabilities"`
}

// Metadata describes the plugin metadata.
type Metadata struct {
	Name             string        `json:"Name"`
	Version          string        `json:"Version"`
	Author           string        `json:"Author,omitempty"`
	GitHubRepository string        `json:"GitHubRepository,omitempty"`
	Logo             string        `json:"Logo,omitempty"`
	ConfigFields     []ConfigField `json:"ConfigFields,omitempty"`
}

// ConfigField describes one configuration field for the plugin.
type ConfigField struct {
	Name        string   `json:"Name"`
	Type        string   `json:"Type"`
	EnumValues  []string `json:"EnumValues,omitempty"`
	Description string   `json:"Description"`
}

// RegistrationCapability declares the capabilities implemented by this plugin.
type RegistrationCapability struct {
	AuthProvider  bool `json:"auth_provider,omitempty"`
	ManagementAPI bool `json:"management_api,omitempty"`
}

// ManagementRegistrationResponse is returned by management.register.
type ManagementRegistrationResponse struct {
	Routes    []ManagementRoute `json:"routes,omitempty"`
	Resources []ResourceRoute   `json:"resources,omitempty"`
}

// ManagementRoute describes one Management API route.
type ManagementRoute struct {
	Method      string `json:"Method"`
	Path        string `json:"Path"`
	Menu        string `json:"Menu,omitempty"`
	Description string `json:"Description,omitempty"`
}

// ResourceRoute describes one browser-navigable resource route.
type ResourceRoute struct {
	Path        string `json:"Path"`
	Menu        string `json:"Menu"`
	Description string `json:"Description"`
}

// ManagementRequest is received by management.handle.
type ManagementRequest struct {
	Method         string              `json:"Method"`
	Path           string              `json:"Path"`
	Headers        map[string][]string `json:"Headers"`
	Query          map[string][]string `json:"Query"`
	Body           []byte              `json:"Body"`
	HostCallbackID string              `json:"host_callback_id,omitempty"`
}

// ManagementResponse is returned by management.handle.
type ManagementResponse struct {
	StatusCode int                 `json:"StatusCode"`
	Headers    map[string][]string `json:"Headers"`
	Body       []byte              `json:"Body"`
}

// HostHTTPRequest describes a request dispatched through host.http.do.
type HostHTTPRequest struct {
	Method         string              `json:"method"`
	URL            string              `json:"url"`
	Headers        map[string][]string `json:"headers,omitempty"`
	Body           []byte              `json:"body,omitempty"`
	HostCallbackID string              `json:"host_callback_id,omitempty"`
}

// HostHTTPResponse describes a response received from host.http.do.
type HostHTTPResponse struct {
	StatusCode int                 `json:"StatusCode"`
	Headers    map[string][]string `json:"Headers"`
	Body       []byte              `json:"Body"`
}

// FlexibleTime handles parsing timestamps from upstream that may be unix seconds, unix milliseconds, or RFC3339 strings.
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON parses various timestamp formats.
func (ft *FlexibleTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\" \t\r\n")
	if s == "" || s == "null" || s == "0" {
		ft.Time = time.Time{}
		return nil
	}

	// Try parsing as integer / float number (Unix timestamp)
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		if n > 1e11 {
			// Milliseconds
			ft.Time = time.UnixMilli(n).UTC()
		} else {
			// Seconds
			ft.Time = time.Unix(n, 0).UTC()
		}
		return nil
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		sec := int64(f)
		if sec > 1e11 {
			ft.Time = time.UnixMilli(sec).UTC()
		} else {
			ft.Time = time.Unix(sec, 0).UTC()
		}
		return nil
	}

	// Try standard RFC3339 / ISO8601 layouts
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}
	for _, layout := range formats {
		if t, err := time.Parse(layout, s); err == nil {
			ft.Time = t.UTC()
			return nil
		}
	}

	return fmt.Errorf("cannot parse %q as FlexibleTime", string(b))
}

// UpstreamCreditsResponse reflects the payload returned by Command Code's /internal/billing/credits.
type UpstreamCreditsResponse struct {
	Credits      map[string]any       `json:"credits"`
	WindowLimits UpstreamWindowLimits `json:"windowLimits"`
}

// UpstreamWindowLimits carries fiveHour and weekly window metrics.
type UpstreamWindowLimits struct {
	Limited  *bool               `json:"limited,omitempty"`
	FiveHour UpstreamWindowLimit `json:"fiveHour"`
	Weekly   UpstreamWindowLimit `json:"weekly"`
}

// UpstreamWindowLimit represents one quota window from upstream.
type UpstreamWindowLimit struct {
	Used     float64      `json:"used"`
	Cap      float64      `json:"cap"`
	Exceeded bool         `json:"exceeded"`
	ResetAt  FlexibleTime `json:"resetAt"`
}

// UsageCreditsData is the formatted credits section.
type UsageCreditsData struct {
	MonthlyCredits           float64        `json:"monthly_credits"`
	OpensourceMonthlyCredits float64        `json:"opensource_monthly_credits"`
	TotalCredits             float64        `json:"total_credits"`
	Details                  map[string]any `json:"details,omitempty"`
}

// UsageWindowLimitData is the formatted window limit section.
type UsageWindowLimitData struct {
	Used           float64 `json:"used"`
	Cap            float64 `json:"cap"`
	Remaining      float64 `json:"remaining"`
	Percentage     float64 `json:"percentage"`
	Exceeded       bool    `json:"exceeded"`
	ResetAt        string  `json:"reset_at"`
	ResetInSeconds int64   `json:"reset_in_seconds"`
}

// UpstreamUsageSummaryResponse reflects Command Code's /internal/usage/summary payload,
// which aggregates usage over the current billing period (monthly).
type UpstreamUsageSummaryResponse struct {
	TotalCount            int64   `json:"totalCount"`
	TotalCost             float64 `json:"totalCost"`
	TotalCredits          float64 `json:"totalCredits"`
	TotalMonthlyCredits   float64 `json:"totalMonthlyCredits"`
	TotalPurchasedCredits float64 `json:"totalPurchasedCredits"`
	PeriodBasis           string  `json:"periodBasis"`
}

// UsageWindowLimitsData contains both windows.
type UsageWindowLimitsData struct {
	Monthly  UsageWindowLimitData `json:"monthly"`
	FiveHour UsageWindowLimitData `json:"five_hour"`
	Weekly   UsageWindowLimitData `json:"weekly"`
}

// PlanInfo represents inferred Command Code subscription plan details.
type PlanInfo struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// FormattedUsageData is the complete formatted usage payload.
type FormattedUsageData struct {
	Plan         PlanInfo              `json:"plan"`
	Credits      UsageCreditsData      `json:"credits"`
	WindowLimits UsageWindowLimitsData `json:"window_limits"`
	UpdatedAt    string                `json:"updated_at"`
}

// FormattedUsageResponse is returned by GET /plugins/commandcode/usage and POST /plugins/commandcode/usage.
type FormattedUsageResponse struct {
	OK           bool                  `json:"ok"`
	Plan         PlanInfo              `json:"plan"`
	Data         FormattedUsageData    `json:"data"`
	Credits      UsageCreditsData      `json:"credits"`
	WindowLimits UsageWindowLimitsData `json:"window_limits"`
	UpdatedAt    string                `json:"updated_at"`
	Error        string                `json:"error,omitempty"`
}

// OpenCodeUsageResponse reflects GET {opencode_api_base}/usage from OpenCode Go.
type OpenCodeUsageResponse struct {
	Usage OpenCodeUsageWindows `json:"usage"`
}

// OpenCodeUsageWindows carries the three usage windows returned by OpenCode Go.
type OpenCodeUsageWindows struct {
	Rolling OpenCodeUsageWindow `json:"rolling"`
	Weekly  OpenCodeUsageWindow `json:"weekly"`
	Monthly OpenCodeUsageWindow `json:"monthly"`
}

// OpenCodeUsageWindow represents one quota window from OpenCode Go.
// Percent is int in the observed upstream payload but parsed as float64 for tolerance.
type OpenCodeUsageWindow struct {
	Status   string  `json:"status"`
	Percent  float64 `json:"percent"`
	ResetsAt string  `json:"resetsAt"` // RFC3339 UTC
}

// OpenCodeFormattedWindows is the formatted OpenCode Go window section.
type OpenCodeFormattedWindows struct {
	Rolling OpenCodeFormattedWindow `json:"rolling"`
	Weekly  OpenCodeFormattedWindow `json:"weekly"`
	Monthly OpenCodeFormattedWindow `json:"monthly"`
}

// OpenCodeFormattedWindow is one formatted OpenCode Go window.
type OpenCodeFormattedWindow struct {
	Status         string  `json:"status"`
	Percent        float64 `json:"percent"`
	Exceeded       bool    `json:"exceeded"`
	ResetAt        string  `json:"reset_at"`
	ResetInSeconds int64   `json:"reset_in_seconds"`
}

// OpenCodeFormattedUsageResponse is the formatted OpenCode Go usage payload.
type OpenCodeFormattedUsageResponse struct {
	OK        bool                     `json:"ok"`
	Provider  string                   `json:"provider"` // "opencode_go"
	Windows   OpenCodeFormattedWindows `json:"windows"`
	UpdatedAt string                   `json:"updated_at"`
	Error     string                   `json:"error,omitempty"`
}

// OpenCodeKeyResult is the per-key outcome of a multi-key OpenCode Go query
// (v0.4.0). Windows is a pointer so failed keys omit the field entirely
// instead of marshaling a zero-value struct with "status":"" noise.
type OpenCodeKeyResult struct {
	KeyID      string                    `json:"key_id"`
	OK         bool                      `json:"ok"`
	Windows    *OpenCodeFormattedWindows `json:"windows,omitempty"`
	StatusCode int                       `json:"status_code"`
	Error      string                    `json:"error,omitempty"`
	UpdatedAt  string                    `json:"updated_at,omitempty"`
}

// OpenCodeMultiKeyResponse is the multi-key OpenCode Go usage payload returned
// by /plugins/commandcode/opencode/usage and the opencode field of /all.
// Top-level Error is non-empty only when no key is configured at all.
type OpenCodeMultiKeyResponse struct {
	OK        bool                `json:"ok"`
	Provider  string              `json:"provider"` // "opencode_go"
	Keys      []OpenCodeKeyResult `json:"keys"`
	UpdatedAt string              `json:"updated_at"`
	Error     string              `json:"error,omitempty"`
}

// AllUsageResponse aggregates both providers for /plugins/commandcode/all.
// Partial failure semantics: each provider's payload is present only on success;
// failures are reported in Errors. CommandCode carries the raw JSON of
// FormattedUsageResponse; OpenCode carries the raw JSON of
// OpenCodeMultiKeyResponse (v0.4.0 breaking change: no longer the single-key
// OpenCodeFormattedUsageResponse).
type AllUsageResponse struct {
	OK          bool              `json:"ok"` // at least one provider succeeded
	CommandCode json.RawMessage   `json:"commandcode,omitempty"`
	OpenCode    json.RawMessage   `json:"opencode,omitempty"`
	Errors      map[string]string `json:"errors,omitempty"`
	UpdatedAt   string            `json:"updated_at"`
}
