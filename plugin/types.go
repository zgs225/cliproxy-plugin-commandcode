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
	AuthProvider  bool `json:"auth_provider"`
	ManagementAPI bool `json:"management_api"`
}

// IdentifierResponse is returned by auth.identifier.
type IdentifierResponse struct {
	Identifier string `json:"identifier"`
}

// AuthData describes a credential record.
type AuthData struct {
	Provider         string            `json:"Provider"`
	ID               string            `json:"ID"`
	FileName         string            `json:"FileName"`
	Label            string            `json:"Label"`
	Prefix           string            `json:"Prefix,omitempty"`
	ProxyURL         string            `json:"ProxyURL,omitempty"`
	Disabled         bool              `json:"Disabled,omitempty"`
	StorageJSON      []byte            `json:"StorageJSON"`
	Metadata         map[string]any    `json:"Metadata,omitempty"`
	Attributes       map[string]string `json:"Attributes,omitempty"`
	NextRefreshAfter time.Time         `json:"NextRefreshAfter,omitempty"`
}

// AuthParseRequest is passed to auth.parse.
type AuthParseRequest struct {
	Provider string         `json:"Provider"`
	Path     string         `json:"Path"`
	FileName string         `json:"FileName"`
	RawJSON  []byte         `json:"RawJSON"`
	Host     map[string]any `json:"Host,omitempty"`
}

// AuthParseResponse is returned by auth.parse.
type AuthParseResponse struct {
	Handled bool       `json:"Handled"`
	Auth    AuthData   `json:"Auth"`
	Auths   []AuthData `json:"Auths,omitempty"`
}

// AuthRefreshRequest is passed to auth.refresh.
type AuthRefreshRequest struct {
	AuthID       string            `json:"AuthID"`
	AuthProvider string            `json:"AuthProvider"`
	StorageJSON  []byte            `json:"StorageJSON"`
	Metadata     map[string]any    `json:"Metadata,omitempty"`
	Attributes   map[string]string `json:"Attributes,omitempty"`
}

// AuthRefreshResponse is returned by auth.refresh.
type AuthRefreshResponse struct {
	Auth             AuthData  `json:"Auth"`
	NextRefreshAfter time.Time `json:"NextRefreshAfter,omitempty"`
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

// FormattedUsageData is the complete formatted usage payload.
type FormattedUsageData struct {
	Credits      UsageCreditsData      `json:"credits"`
	WindowLimits UsageWindowLimitsData `json:"window_limits"`
	UpdatedAt    string                `json:"updated_at"`
}

// FormattedUsageResponse is returned by GET /plugins/commandcode/usage and POST /plugins/commandcode/usage.
type FormattedUsageResponse struct {
	OK           bool                  `json:"ok"`
	Data         FormattedUsageData    `json:"data"`
	Credits      UsageCreditsData      `json:"credits"`
	WindowLimits UsageWindowLimitsData `json:"window_limits"`
	UpdatedAt    string                `json:"updated_at"`
	Error        string                `json:"error,omitempty"`
}
