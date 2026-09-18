package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// RegisterManagement handles management.register method.
func RegisterManagement() (ManagementRegistrationResponse, error) {
	return ManagementRegistrationResponse{
		Routes: []ManagementRoute{
			{
				Method:      http.MethodGet,
				Path:        "/plugins/commandcode/usage",
				Description: "Query Command Code credits and window limits usage",
			},
			{
				Method:      http.MethodPost,
				Path:        "/plugins/commandcode/usage",
				Description: "Query Command Code credits and window limits usage with custom session_token",
			},
			{
				Method:      http.MethodGet,
				Path:        "/plugins/commandcode/opencode/usage",
				Description: "Query OpenCode Go usage windows (rolling/weekly/monthly)",
			},
			{
				Method:      http.MethodPost,
				Path:        "/plugins/commandcode/opencode/usage",
				Description: "Query OpenCode Go usage windows with custom opencode_api_key",
			},
			{
				Method:      http.MethodGet,
				Path:        "/plugins/commandcode/all",
				Description: "Query both Command Code and OpenCode Go usage (aggregated, partial failures reported in errors map)",
			},
			{
				Method:      http.MethodPost,
				Path:        "/plugins/commandcode/all",
				Description: "Query both providers with custom credentials in request body",
			},
		},
		Resources: []ResourceRoute{
			{
				Path:        "/quota",
				Menu:        "用量配额",
				Description: "Command Code + OpenCode Go 用量与限额卡片",
			},
		},
	}, nil
}

// HandleManagement processes management.handle requests for API routes and resource pages.
func HandleManagement(ctx context.Context, req ManagementRequest, cfg *PluginConfig) (ManagementResponse, error) {
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	path := strings.TrimSpace(req.Path)

	// 1. Serve Quota Resource Page
	if method == http.MethodGet && (strings.HasSuffix(path, "/quota") || strings.HasSuffix(path, "/quota/")) {
		return ManagementResponse{
			StatusCode: http.StatusOK,
			Headers: map[string][]string{
				"Content-Type": {"text/html; charset=utf-8"},
			},
			Body: GetQuotaPageHTML(),
		}, nil
	}

	// 2. OpenCode Go usage API — MUST be matched before the generic /usage
	// suffix match below, otherwise "/plugins/commandcode/opencode/usage"
	// would be swallowed by the Command Code handler.
	if strings.HasSuffix(path, "/plugins/commandcode/opencode/usage") {
		switch method {
		case http.MethodGet, http.MethodPost:
			return handleOpenCodeUsage(ctx, req, cfg)
		default:
			return ManagementResponse{
				StatusCode: http.StatusMethodNotAllowed,
				Headers: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: []byte(`{"ok":false,"error":"method not allowed"}`),
			}, nil
		}
	}

	// 3. Aggregated usage API (both providers) — does not end with "/usage",
	// but registered before the generic match for clarity and future safety.
	if strings.HasSuffix(path, "/plugins/commandcode/all") {
		switch method {
		case http.MethodGet, http.MethodPost:
			return handleAllUsage(ctx, req, cfg)
		default:
			return ManagementResponse{
				StatusCode: http.StatusMethodNotAllowed,
				Headers: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: []byte(`{"ok":false,"error":"method not allowed"}`),
			}, nil
		}
	}

	// 4. Command Code usage API (GET / POST) — generic suffix match kept as-is.
	if strings.HasSuffix(path, "/plugins/commandcode/usage") || strings.HasSuffix(path, "/usage") {
		switch method {
		case http.MethodGet:
			return handleGetUsage(ctx, req, cfg)
		case http.MethodPost:
			return handlePostUsage(ctx, req, cfg)
		default:
			return ManagementResponse{
				StatusCode: http.StatusMethodNotAllowed,
				Headers: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: []byte(`{"ok":false,"error":"method not allowed"}`),
			}, nil
		}
	}

	// Unknown path
	return ManagementResponse{
		StatusCode: http.StatusNotFound,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: []byte(`{"ok":false,"error":"not found"}`),
	}, nil
}

func handleGetUsage(ctx context.Context, req ManagementRequest, cfg *PluginConfig) (ManagementResponse, error) {
	sessionToken := ""
	apiBase := ""

	// Check query params
	if len(req.Query) > 0 {
		if tokens, ok := req.Query["session_token"]; ok && len(tokens) > 0 {
			sessionToken = tokens[0]
		} else if tokens, ok := req.Query["token"]; ok && len(tokens) > 0 {
			sessionToken = tokens[0]
		}
		if bases, ok := req.Query["api_base"]; ok && len(bases) > 0 {
			apiBase = bases[0]
		}
	}

	// Fallback to plugin config
	if sessionToken == "" && cfg != nil {
		sessionToken = cfg.GetSessionToken()
	}
	if apiBase == "" && cfg != nil {
		apiBase = cfg.GetAPIBase()
	}

	return executeUsageQuery(ctx, apiBase, sessionToken, req.HostCallbackID)
}

func handlePostUsage(ctx context.Context, req ManagementRequest, cfg *PluginConfig) (ManagementResponse, error) {
	var body struct {
		SessionToken string `json:"session_token"`
		Token        string `json:"token"`
		APIBase      string `json:"api_base"`
	}

	if len(req.Body) > 0 {
		_ = json.Unmarshal(req.Body, &body)
	}

	sessionToken := body.SessionToken
	if sessionToken == "" {
		sessionToken = body.Token
	}
	apiBase := body.APIBase

	// Fallback to plugin config if body didn't specify
	if sessionToken == "" && cfg != nil {
		sessionToken = cfg.GetSessionToken()
	}
	if apiBase == "" && cfg != nil {
		apiBase = cfg.GetAPIBase()
	}

	return executeUsageQuery(ctx, apiBase, sessionToken, req.HostCallbackID)
}

func executeUsageQuery(ctx context.Context, apiBase, sessionToken, hostCallbackID string) (ManagementResponse, error) {
	if strings.TrimSpace(sessionToken) == "" {
		resBytes, _ := json.Marshal(map[string]any{
			"ok":    false,
			"error": "session_token is required. Configure session_token in plugin config, provide a credential file, or pass session_token in request",
		})
		return ManagementResponse{
			StatusCode: http.StatusBadRequest,
			Headers: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: resBytes,
		}, nil
	}

	raw, statusCode, errFetch := FetchCreditsRaw(ctx, apiBase, sessionToken, hostCallbackID)
	if errFetch != nil {
		resBytes, _ := json.Marshal(map[string]any{
			"ok":          false,
			"status_code": statusCode,
			"error":       errFetch.Error(),
		})
		if statusCode == 0 || statusCode == http.StatusOK {
			statusCode = http.StatusBadGateway
		}
		return ManagementResponse{
			StatusCode: statusCode,
			Headers: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: resBytes,
		}, nil
	}

	if statusCode != http.StatusOK {
		resBytes, _ := json.Marshal(map[string]any{
			"ok":          false,
			"status_code": statusCode,
			"error":       "upstream returned non-200 status",
			"body":        string(raw),
		})
		return ManagementResponse{
			StatusCode: statusCode,
			Headers: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: resBytes,
		}, nil
	}

	// Fetch billing-period (monthly) usage totals; non-fatal if unavailable.
	var summary *UpstreamUsageSummaryResponse
	if sumRaw, sumStatus, sumErr := FetchUsageSummaryRaw(ctx, apiBase, sessionToken, hostCallbackID); sumErr == nil && sumStatus == http.StatusOK {
		var parsed UpstreamUsageSummaryResponse
		if errSum := json.Unmarshal(sumRaw, &parsed); errSum == nil && parsed.TotalMonthlyCredits > 0 {
			summary = &parsed
		}
	}

	usage, errParse := ParseAndFormatUsage(raw, summary, time.Now().UTC())
	if errParse != nil {
		resBytes, _ := json.Marshal(map[string]any{
			"ok":    false,
			"error": "failed to parse upstream usage: " + errParse.Error(),
		})
		return ManagementResponse{
			StatusCode: http.StatusBadGateway,
			Headers: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: resBytes,
		}, nil
	}

	resBytes, _ := json.Marshal(usage)
	return ManagementResponse{
		StatusCode: http.StatusOK,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: resBytes,
	}, nil
}

// handleOpenCodeUsage serves GET/POST /plugins/commandcode/opencode/usage.
// Credentials can be overridden via POST body only (opencode_api_key / api_key);
// GET queries are read-only against the plugin config — query parameter
// overrides are intentionally not supported to keep secrets out of URLs.
func handleOpenCodeUsage(ctx context.Context, req ManagementRequest, cfg *PluginConfig) (ManagementResponse, error) {
	apiKey := ""
	apiBase := ""

	if strings.EqualFold(strings.ToUpper(strings.TrimSpace(req.Method)), http.MethodPost) && len(req.Body) > 0 {
		var body struct {
			OpenCodeAPIKey  string `json:"opencode_api_key"`
			APIKey          string `json:"api_key"`
			OpenCodeAPIBase string `json:"opencode_api_base"`
		}
		_ = json.Unmarshal(req.Body, &body)
		apiKey = body.OpenCodeAPIKey
		if apiKey == "" {
			apiKey = body.APIKey
		}
		apiBase = body.OpenCodeAPIBase
	}

	// Fallback to plugin config
	if apiKey == "" && cfg != nil {
		apiKey = cfg.GetOpenCodeAPIKey()
	}
	if apiBase == "" && cfg != nil {
		apiBase = cfg.GetOpenCodeAPIBase()
	}

	return handleOpenCodeUsageWithKey(ctx, apiBase, apiKey, req.HostCallbackID)
}

// handleAllUsage serves GET/POST /plugins/commandcode/all: it queries both
// providers sequentially (no goroutines — the host.http.do bridge's host-side
// concurrency safety cannot be verified and shared maps would race under -race).
// Partial failure: OK=true as long as at least one provider succeeds; failures
// land in the Errors map and successful fields are omitted when absent.
// HTTP status: any success → 200; all failed due to missing local credentials → 400;
// all failed due to upstream errors → 502.
func handleAllUsage(ctx context.Context, req ManagementRequest, cfg *PluginConfig) (ManagementResponse, error) {
	sessionToken := ""
	opencodeKey := ""

	if strings.EqualFold(strings.ToUpper(strings.TrimSpace(req.Method)), http.MethodPost) && len(req.Body) > 0 {
		var body struct {
			SessionToken   string `json:"session_token"`
			OpencodeAPIKey string `json:"opencode_api_key"`
		}
		_ = json.Unmarshal(req.Body, &body)
		sessionToken = body.SessionToken
		opencodeKey = body.OpencodeAPIKey
	}

	// Fallback to plugin config
	if sessionToken == "" && cfg != nil {
		sessionToken = cfg.GetSessionToken()
	}
	if opencodeKey == "" && cfg != nil {
		opencodeKey = cfg.GetOpenCodeAPIKey()
	}
	apiBase := ""
	if cfg != nil {
		apiBase = cfg.GetAPIBase()
	}
	ocAPIBase := ""
	if cfg != nil {
		ocAPIBase = cfg.GetOpenCodeAPIBase()
	}

	now := time.Now().UTC()
	resp := AllUsageResponse{OK: false, UpdatedAt: now.Format(time.RFC3339)}
	errs := make(map[string]string)
	localMissing := 0
	upstreamFailed := 0
	succeeded := 0

	// Provider 1: Command Code (reuses executeUsageQuery).
	ccResp, _ := executeUsageQuery(ctx, apiBase, sessionToken, req.HostCallbackID)
	if ccResp.StatusCode == http.StatusOK {
		resp.CommandCode = ccResp.Body
		succeeded++
	} else {
		ccErr := extractErrorResponseMessage(ccResp.Body)
		errs["commandcode"] = ccErr
		if isLocalCredentialError(ccErr) {
			localMissing++
		} else {
			upstreamFailed++
		}
	}

	// Provider 2: OpenCode Go (same classification via isLocalCredentialError,
	// not by HTTP 400 alone: upstream 4xx may be passed through and must not
	// be misclassified as a local configuration problem).
	if strings.TrimSpace(opencodeKey) != "" {
		ocResp, _ := handleOpenCodeUsageWithKey(ctx, ocAPIBase, opencodeKey, req.HostCallbackID)
		if ocResp.StatusCode == http.StatusOK {
			resp.OpenCode = ocResp.Body
			succeeded++
		} else {
			ocErr := extractErrorResponseMessage(ocResp.Body)
			errs["opencode"] = ocErr
			if isLocalCredentialError(ocErr) {
				localMissing++
			} else {
				upstreamFailed++
			}
		}
	} else {
		errs["opencode"] = "missing opencode_api_key: configure opencode_api_key in plugin config or pass it in the request body"
		localMissing++
	}

	if len(errs) > 0 {
		resp.Errors = errs
	}
	resp.OK = succeeded > 0

	statusCode := http.StatusOK
	if !resp.OK {
		if upstreamFailed == 0 && localMissing == len(errs) {
			statusCode = http.StatusBadRequest
		} else {
			statusCode = http.StatusBadGateway
		}
	}

	resBytes, _ := json.Marshal(resp)
	return ManagementResponse{
		StatusCode: statusCode,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: resBytes,
	}, nil
}

// isLocalCredentialError reports whether an /all provider error is a local
// configuration problem (missing credential in plugin config), as opposed to
// an upstream failure. Local-credential errors carry fixed message prefixes;
// upstream 4xx/5xx never match them, so the /all 400-vs-502 classification
// does not rely on the HTTP status alone.
func isLocalCredentialError(msg string) bool {
	for _, prefix := range []string{
		"session_token is required",
		"opencode_api_key is required",
	} {
		if strings.HasPrefix(msg, prefix) {
			return true
		}
	}
	return false
}

// handleOpenCodeUsageWithKey runs the OpenCode usage query with an explicit
// credential, shared by handleOpenCodeUsage and handleAllUsage.
func handleOpenCodeUsageWithKey(ctx context.Context, apiBase, apiKey, hostCallbackID string) (ManagementResponse, error) {
	if strings.TrimSpace(apiKey) == "" {
		resBytes, _ := json.Marshal(map[string]any{
			"ok":    false,
			"error": "opencode_api_key is required. Configure opencode_api_key in plugin config or pass it in the request body",
		})
		return ManagementResponse{
			StatusCode: http.StatusBadRequest,
			Headers: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: resBytes,
		}, nil
	}

	raw, statusCode, errFetch := FetchOpenCodeUsageRaw(ctx, apiBase, apiKey, hostCallbackID)
	if errFetch != nil {
		errMsg := fmt.Sprintf("opencode upstream request failed: %s", errFetch.Error())
		resBytes, _ := json.Marshal(map[string]any{
			"ok":          false,
			"status_code": statusCode,
			"error":       errMsg,
		})
		if statusCode == 0 || statusCode == http.StatusOK {
			statusCode = http.StatusBadGateway
		}
		return ManagementResponse{
			StatusCode: statusCode,
			Headers: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: resBytes,
		}, nil
	}

	if statusCode != http.StatusOK {
		resBytes, _ := json.Marshal(map[string]any{
			"ok":          false,
			"status_code": statusCode,
			"error":       fmt.Sprintf("opencode upstream returned %d: check opencode_api_key", statusCode),
		})
		return ManagementResponse{
			StatusCode: statusCode,
			Headers: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: resBytes,
		}, nil
	}

	usage, errParse := ParseOpenCodeUsage(raw, time.Now().UTC())
	if errParse != nil {
		resBytes, _ := json.Marshal(map[string]any{
			"ok":    false,
			"error": "failed to parse opencode upstream usage: " + errParse.Error(),
		})
		return ManagementResponse{
			StatusCode: http.StatusBadGateway,
			Headers: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: resBytes,
		}, nil
	}

	resBytes, _ := json.Marshal(usage)
	return ManagementResponse{
		StatusCode: http.StatusOK,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: resBytes,
	}, nil
}

// extractErrorResponseMessage pulls the "error" field out of a JSON error body.
func extractErrorResponseMessage(body []byte) string {
	var parsed struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Error != "" {
		return parsed.Error
	}
	return "unknown error"
}
