package plugin

import (
	"context"
	"encoding/json"
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
		},
		Resources: []ResourceRoute{
			{
				Path:        "/quota",
				Menu:        "Command Code 配额",
				Description: "Command Code 用量与限额卡片",
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

	// 2. Serve Usage API (GET / POST)
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

	usage, errParse := ParseAndFormatUsage(raw, time.Now().UTC())
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
