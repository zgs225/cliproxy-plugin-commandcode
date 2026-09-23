package plugin

import (
	"encoding/base64"
)

// QuotaPageHTML contains the full standalone HTML/CSS/JS single-file QuotaCard UI.
const QuotaPageHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>用量配额 - CLIProxyAPI</title>
  <style>
    :root {
      --bg-page: #f8fafc;
      --bg-card: #ffffff;
      --bg-subtle: #f1f5f9;
      --border-color: #e2e8f0;
      --text-main: #0f172a;
      --text-muted: #64748b;
      --text-dim: #94a3b8;
      --primary: #3b82f6;
      --primary-hover: #2563eb;
      --primary-subtle: #eff6ff;
      --success: #10b981;
      --success-subtle: #ecfdf5;
      --warning: #f59e0b;
      --warning-subtle: #fffbeb;
      --danger: #ef4444;
      --danger-subtle: #fef2f2;
      --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
      --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
      --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
      --radius-sm: 6px;
      --radius-md: 10px;
      --radius-lg: 16px;
    }

    /* Auto mode fallback: when no explicit data-theme is set, follow OS dark preference */
    @media (prefers-color-scheme: dark) {
      :root:not([data-theme="light"]):not([data-theme="white"]) {
        --bg-page: #0b0f19;
        --bg-card: #151d30;
        --bg-subtle: #1e293b;
        --border-color: #334155;
        --text-main: #f8fafc;
        --text-muted: #94a3b8;
        --text-dim: #64748b;
        --primary: #60a5fa;
        --primary-hover: #3b82f6;
        --primary-subtle: rgba(59, 130, 246, 0.12);
        --success: #34d399;
        --success-subtle: rgba(16, 185, 129, 0.12);
        --warning: #fbbf24;
        --warning-subtle: rgba(245, 158, 11, 0.12);
        --danger: #f87171;
        --danger-subtle: rgba(239, 68, 68, 0.12);
        --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.3);
        --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.4);
        --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.5);
      }
    }

    /* Explicit dark theme from Management Center */
    :root[data-theme="dark"] {
      --bg-page: #0b0f19;
      --bg-card: #151d30;
      --bg-subtle: #1e293b;
      --border-color: #334155;
      --text-main: #f8fafc;
      --text-muted: #94a3b8;
      --text-dim: #64748b;
      --primary: #60a5fa;
      --primary-hover: #3b82f6;
      --primary-subtle: rgba(59, 130, 246, 0.12);
      --success: #34d399;
      --success-subtle: rgba(16, 185, 129, 0.12);
      --warning: #fbbf24;
      --warning-subtle: rgba(245, 158, 11, 0.12);
      --danger: #f87171;
      --danger-subtle: rgba(239, 68, 68, 0.12);
      --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.3);
      --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.4);
      --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.5);
    }

    /* Explicit white / light theme from Management Center: overrides OS dark preference */
    :root[data-theme="white"],
    :root[data-theme="light"] {
      --bg-page: #f8fafc;
      --bg-card: #ffffff;
      --bg-subtle: #f1f5f9;
      --border-color: #e2e8f0;
      --text-main: #0f172a;
      --text-muted: #64748b;
      --text-dim: #94a3b8;
      --primary: #3b82f6;
      --primary-hover: #2563eb;
      --primary-subtle: #eff6ff;
      --success: #10b981;
      --success-subtle: #ecfdf5;
      --warning: #f59e0b;
      --warning-subtle: #fffbeb;
      --danger: #ef4444;
      --danger-subtle: #fef2f2;
      --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
      --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
      --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
    }

    * {
      box-sizing: border-box;
      margin: 0;
      padding: 0;
    }

    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
      background-color: var(--bg-page);
      color: var(--text-main);
      min-height: 100vh;
      padding: 24px 16px;
      display: flex;
      justify-content: center;
      align-items: flex-start;
      line-height: 1.5;
    }

    .container {
      width: 100%;
      max-width: 860px;
      margin: 0 auto;
    }

    .header-card {
      background: var(--bg-card);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-lg);
      padding: 20px 24px;
      margin-bottom: 20px;
      box-shadow: var(--shadow-sm);
      display: flex;
      justify-content: space-between;
      align-items: center;
      flex-wrap: wrap;
      gap: 16px;
    }

    .brand-section {
      display: flex;
      align-items: center;
      gap: 14px;
    }

    .brand-icon {
      width: 44px;
      height: 44px;
      border-radius: var(--radius-md);
      background: linear-gradient(135deg, #3b82f6, #8b5cf6);
      display: flex;
      align-items: center;
      justify-content: center;
      color: white;
      box-shadow: 0 4px 10px rgba(59, 130, 246, 0.3);
    }

    .brand-icon svg {
      width: 26px;
      height: 26px;
    }

    .brand-title {
      font-size: 20px;
      font-weight: 700;
      color: var(--text-main);
      display: flex;
      align-items: center;
      gap: 8px;
    }

    .version-tag {
      font-size: 11px;
      font-weight: 600;
      padding: 2px 8px;
      border-radius: 9999px;
      background: var(--primary-subtle);
      color: var(--primary);
      border: 1px solid var(--primary);
    }

    .plan-tag {
      font-size: 11px;
      font-weight: 700;
      padding: 2px 10px;
      border-radius: 9999px;
      background: linear-gradient(135deg, rgba(139, 92, 246, 0.15), rgba(59, 130, 246, 0.15));
      color: #8b5cf6;
      border: 1px solid rgba(139, 92, 246, 0.35);
      display: inline-flex;
      align-items: center;
      gap: 4px;
      letter-spacing: 0.3px;
    }

    :root[data-theme="dark"] .plan-tag {
      background: linear-gradient(135deg, rgba(167, 139, 250, 0.2), rgba(96, 165, 250, 0.2));
      color: #c084fc;
      border-color: rgba(167, 139, 250, 0.45);
    }

    @media (prefers-color-scheme: dark) {
      :root:not([data-theme="light"]):not([data-theme="white"]) .plan-tag {
        background: linear-gradient(135deg, rgba(167, 139, 250, 0.2), rgba(96, 165, 250, 0.2));
        color: #c084fc;
        border-color: rgba(167, 139, 250, 0.45);
      }
    }

    .action-group {
      display: flex;
      align-items: center;
      gap: 10px;
    }

    .status-badge {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      font-size: 12px;
      font-weight: 600;
      padding: 6px 12px;
      border-radius: 9999px;
      background: var(--bg-subtle);
      color: var(--text-muted);
      border: 1px solid var(--border-color);
    }

    .status-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: var(--text-dim);
    }

    .status-badge.online .status-dot { background: var(--success); }
    .status-badge.online { background: var(--success-subtle); color: var(--success); border-color: rgba(16, 185, 129, 0.3); }

    .status-badge.warning .status-dot { background: var(--warning); }
    .status-badge.warning { background: var(--warning-subtle); color: var(--warning); border-color: rgba(245, 158, 11, 0.3); }

    .status-badge.exceeded .status-dot { background: var(--danger); animation: pulse 1.5s infinite; }
    .status-badge.exceeded { background: var(--danger-subtle); color: var(--danger); border-color: rgba(239, 68, 68, 0.3); }

    @keyframes pulse {
      0% { transform: scale(0.95); opacity: 0.8; }
      50% { transform: scale(1.2); opacity: 1; }
      100% { transform: scale(0.95); opacity: 0.8; }
    }

    .btn {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      font-size: 13px;
      font-weight: 600;
      padding: 8px 14px;
      border-radius: var(--radius-md);
      cursor: pointer;
      transition: all 0.2s ease;
      border: 1px solid var(--border-color);
      background: var(--bg-card);
      color: var(--text-main);
      outline: none;
    }

    .btn:hover {
      background: var(--bg-subtle);
      border-color: var(--primary);
    }

    .btn-primary {
      background: var(--primary);
      color: white;
      border-color: var(--primary);
    }

    .btn-primary:hover {
      background: var(--primary-hover);
      border-color: var(--primary-hover);
    }

    .btn svg {
      width: 15px;
      height: 15px;
    }

    .spin {
      animation: spin 1s linear infinite;
    }

    @keyframes spin {
      from { transform: rotate(0deg); }
      to { transform: rotate(360deg); }
    }

    /* Error / Warning Alert Banner */
    .alert {
      display: none;
      padding: 14px 18px;
      border-radius: var(--radius-md);
      margin-bottom: 20px;
      font-size: 13px;
      align-items: center;
      gap: 10px;
      border: 1px solid transparent;
    }

    .alert.show {
      display: flex;
    }

    .alert-danger {
      background: var(--danger-subtle);
      color: var(--danger);
      border-color: rgba(239, 68, 68, 0.3);
    }

    .alert-warning {
      background: var(--warning-subtle);
      color: var(--warning);
      border-color: rgba(245, 158, 11, 0.3);
    }

    /* Drawer / Settings */
    .settings-drawer {
      display: none;
      background: var(--bg-card);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-lg);
      padding: 20px;
      margin-bottom: 20px;
      box-shadow: var(--shadow-sm);
    }

    .settings-drawer.open {
      display: block;
    }

    .drawer-title {
      font-size: 15px;
      font-weight: 700;
      margin-bottom: 14px;
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .form-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 14px;
    }

    @media (max-width: 640px) {
      .form-grid {
        grid-template-columns: 1fr;
      }
    }

    .form-group {
      display: flex;
      flex-direction: column;
      gap: 6px;
    }

    .form-group label {
      font-size: 12px;
      font-weight: 600;
      color: var(--text-muted);
    }

    .form-control {
      background: var(--bg-subtle);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-sm);
      padding: 8px 12px;
      font-size: 13px;
      color: var(--text-main);
      outline: none;
      font-family: monospace;
    }

    .form-control:focus {
      border-color: var(--primary);
    }

    /* Metrics Overview Grid */
    .metrics-grid {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 16px;
      margin-bottom: 20px;
    }

    @media (max-width: 700px) {
      .metrics-grid {
        grid-template-columns: 1fr;
      }
    }

    .metric-card {
      background: var(--bg-card);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-lg);
      padding: 20px;
      box-shadow: var(--shadow-sm);
      display: flex;
      flex-direction: column;
      gap: 8px;
      transition: transform 0.2s ease, box-shadow 0.2s ease;
    }

    .metric-card:hover {
      box-shadow: var(--shadow-md);
      transform: translateY(-2px);
    }

    .metric-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .metric-title {
      font-size: 13px;
      font-weight: 600;
      color: var(--text-muted);
    }

    .metric-icon {
      width: 28px;
      height: 28px;
      border-radius: var(--radius-sm);
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 14px;
      background: var(--bg-subtle);
      color: var(--text-muted);
    }

    .metric-value {
      font-size: 28px;
      font-weight: 800;
      color: var(--text-main);
      font-feature-settings: "tnum";
    }

    .metric-desc {
      font-size: 12px;
      color: var(--text-dim);
    }

    /* Window Limit Quota Cards */
    .quota-section {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 20px;
      margin-bottom: 24px;
    }

    @media (max-width: 768px) {
      .quota-section {
        grid-template-columns: 1fr;
      }
    }

    .quota-card {
      background: var(--bg-card);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-lg);
      padding: 24px;
      box-shadow: var(--shadow-sm);
      display: flex;
      flex-direction: column;
      gap: 18px;
      position: relative;
      overflow: hidden;
    }

    .quota-card.is-exceeded {
      border-color: rgba(239, 68, 68, 0.4);
    }

    .quota-card.is-exceeded::before {
      content: "";
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      height: 4px;
      background: var(--danger);
    }

    .quota-card-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
    }

    .quota-tag {
      font-size: 11px;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.5px;
      padding: 3px 8px;
      border-radius: 4px;
      background: var(--primary-subtle);
      color: var(--primary);
    }

    .quota-name {
      font-size: 18px;
      font-weight: 700;
      color: var(--text-main);
      margin-top: 4px;
    }

    .quota-percent-badge {
      font-size: 15px;
      font-weight: 800;
      padding: 4px 10px;
      border-radius: var(--radius-md);
      background: var(--bg-subtle);
      color: var(--text-main);
    }

    .quota-stats-row {
      display: flex;
      justify-content: space-between;
      align-items: baseline;
      font-size: 13px;
      color: var(--text-muted);
    }

    .quota-usage-num {
      font-size: 24px;
      font-weight: 800;
      color: var(--text-main);
      font-feature-settings: "tnum";
    }

    .quota-cap-num {
      font-size: 14px;
      color: var(--text-muted);
      font-weight: 500;
    }

    /* Progress bar */
    .progress-track {
      width: 100%;
      height: 12px;
      background: var(--bg-subtle);
      border-radius: 9999px;
      overflow: hidden;
      position: relative;
    }

    .progress-bar {
      height: 100%;
      border-radius: 9999px;
      transition: width 0.6s cubic-bezier(0.4, 0, 0.2, 1), background 0.3s ease;
      width: 0%;
      background: linear-gradient(90deg, #10b981, #06b6d4);
    }

    .progress-bar.warning {
      background: linear-gradient(90deg, #f59e0b, #fbbf24);
    }

    .progress-bar.danger {
      background: linear-gradient(90deg, #ef4444, #f87171);
    }

    /* Reset countdown */
    .reset-box {
      background: var(--bg-subtle);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-md);
      padding: 12px 14px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 10px;
    }

    .reset-label {
      display: flex;
      align-items: center;
      gap: 6px;
      font-size: 12px;
      color: var(--text-muted);
      font-weight: 600;
    }

    .reset-label svg {
      width: 14px;
      height: 14px;
    }

    .countdown-timer {
      font-size: 13px;
      font-weight: 700;
      color: var(--text-main);
      font-family: monospace;
      letter-spacing: 0.5px;
    }

    /* Tab Bar (Command Code | OpenCode Go | All) */
    .tab-bar {
      display: flex;
      gap: 6px;
      background: var(--bg-card);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-lg);
      padding: 6px;
      margin-bottom: 20px;
      box-shadow: var(--shadow-sm);
    }

    .tab-btn {
      flex: 1;
      border: none;
      background: transparent;
      color: var(--text-muted);
      font-size: 13px;
      font-weight: 600;
      padding: 8px 12px;
      border-radius: var(--radius-md);
      cursor: pointer;
      transition: all 0.2s ease;
      font-family: inherit;
    }

    .tab-btn:hover {
      color: var(--text-main);
      background: var(--bg-subtle);
    }

    .tab-btn.active {
      background: var(--primary);
      color: #fff;
    }

    .tab-section {
      display: none;
    }

    .tab-section.active {
      display: block;
    }

    /* Command Code tab head: title + plan badge (planBadge moved out of header) */
    .cc-tab-head {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;
      gap: 10px;
    }

    .cc-tab-title {
      font-size: 15px;
      font-weight: 700;
      color: var(--text-main);
    }

    /* OpenCode multi-key groups: one group per key, 3 compact window rows each */
    .oc-key-group {
      background: var(--bg-card);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-lg);
      padding: 20px;
      box-shadow: var(--shadow-sm);
      display: flex;
      flex-direction: column;
      gap: 12px;
      margin-bottom: 16px;
    }

    .oc-key-head {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 10px;
    }

    .oc-key-id {
      font-size: 13px;
      font-family: monospace;
      color: var(--text-muted);
    }

    .oc-key-body {
      display: flex;
      flex-direction: column;
      gap: 10px;
    }

    .oc-win-row {
      display: grid;
      grid-template-columns: 92px 1fr 52px 84px;
      align-items: center;
      gap: 12px;
    }

    @media (max-width: 640px) {
      .oc-win-row {
        grid-template-columns: 86px 1fr 48px;
      }
      .oc-win-reset {
        display: none;
      }
    }

    .oc-win-name {
      font-size: 12px;
      font-weight: 600;
      color: var(--text-muted);
      white-space: nowrap;
    }

    /* .progress-track combo class: 8px mini bar, same as .all-grid */
    .oc-win-bar {
      height: 8px;
    }

    .oc-win-pct {
      font-size: 13px;
      font-weight: 700;
      color: var(--text-main);
      white-space: nowrap;
      font-feature-settings: "tnum";
    }

    .oc-win-pct.warn { color: var(--warning); }
    .oc-win-pct.bad { color: var(--danger); }

    /* .countdown-timer combo class: smaller reset countdown */
    .oc-win-reset {
      font-size: 12px;
    }

    .oc-key-error {
      font-size: 12px;
      color: var(--danger);
      word-break: break-all;
    }

    .all-key-id {
      font-family: monospace;
      font-size: 12px;
    }

    textarea.form-control {
      resize: vertical;
    }

    /* All tab: two provider cards side by side */
    .all-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 20px;
      margin-bottom: 24px;
    }

    @media (max-width: 768px) {
      .all-grid {
        grid-template-columns: 1fr;
      }
    }

    .all-group-title {
      font-size: 13px;
      font-weight: 700;
      color: var(--text-muted);
      margin: 8px 0 10px 2px;
    }

    .all-group-title:first-child {
      margin-top: 0;
    }

    .all-provider-card {
      background: var(--bg-card);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-lg);
      padding: 20px;
      box-shadow: var(--shadow-sm);
      display: flex;
      flex-direction: column;
      gap: 12px;
    }

    .all-provider-head {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 10px;
    }

    .all-provider-name {
      font-size: 15px;
      font-weight: 700;
      color: var(--text-main);
    }

    .all-provider-body {
      display: flex;
      flex-direction: column;
      gap: 8px;
    }

    .all-summary-row {
      display: flex;
      justify-content: space-between;
      align-items: baseline;
      gap: 10px;
      font-size: 13px;
      color: var(--text-muted);
    }

    .all-summary-value {
      font-weight: 700;
      color: var(--text-main);
      font-feature-settings: "tnum";
    }

    .all-grid .progress-track {
      height: 8px;
    }

    /* Per-tab provider failure error cards */
    .error-card {
      display: none;
      background: var(--danger-subtle);
      border: 1px solid rgba(239, 68, 68, 0.3);
      border-radius: var(--radius-lg);
      padding: 20px;
      margin-bottom: 20px;
      flex-direction: column;
      gap: 8px;
    }

    .error-card.show {
      display: flex;
    }

    .error-card-title {
      color: var(--danger);
      font-weight: 700;
      font-size: 14px;
      display: flex;
      align-items: center;
      gap: 8px;
    }

    .error-card-msg {
      color: var(--danger);
      font-size: 13px;
      word-break: break-all;
    }

    .form-hint {
      font-size: 11px;
      color: var(--text-dim);
    }

    /* Footer */
    .footer-bar {
      display: flex;
      justify-content: space-between;
      align-items: center;
      flex-wrap: wrap;
      gap: 12px;
      padding: 12px 8px;
      font-size: 12px;
      color: var(--text-dim);
    }

    .footer-links {
      display: flex;
      gap: 16px;
    }

    .footer-links a {
      color: var(--text-muted);
      text-decoration: none;
    }

    .footer-links a:hover {
      color: var(--primary);
    }
  </style>
</head>
<body>
  <div class="container">
    <!-- Header Card -->
    <div class="header-card">
      <div class="brand-section">
        <div class="brand-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="16 18 22 12 16 6"></polyline>
            <polyline points="8 6 2 12 8 18"></polyline>
          </svg>
        </div>
        <div>
          <div class="brand-title">
            用量配额
            <span class="version-tag">v0.4.5</span>
          </div>
        </div>
      </div>

      <div class="action-group">
        <button id="btnSettings" class="btn" title="配置选项">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="3"></circle>
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
          </svg>
        </button>

        <button id="btnRefresh" class="btn btn-primary">
          <svg id="refreshIcon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="23 4 23 10 17 10"></polyline>
            <polyline points="1 20 1 14 7 14"></polyline>
            <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
          </svg>
          刷新
        </button>
      </div>
    </div>

    <!-- Tab Bar -->
    <div id="tabBar" class="tab-bar">
      <button type="button" class="tab-btn active" data-tab="all">All</button>
      <button type="button" class="tab-btn" data-tab="commandcode">Command Code</button>
      <button type="button" class="tab-btn" data-tab="opencode">OpenCode Go</button>
    </div>

    <!-- Alert Message -->
    <div id="alertBox" class="alert alert-danger">
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
      <span id="alertMsg">加载失败</span>
    </div>

    <!-- Settings Drawer -->
    <div id="settingsDrawer" class="settings-drawer">
      <div class="drawer-title">
        <span>诊断与访问配置</span>
        <button id="btnCloseSettings" class="btn" style="padding:4px 8px;">关闭</button>
      </div>
      <div class="form-grid">
        <div class="form-group">
          <label for="inputMgmtKey">CLIProxyAPI 管理密钥 (Management Key)</label>
          <input type="password" id="inputMgmtKey" class="form-control" placeholder="留空则自动从浏览器 localStorage 探测" />
        </div>
        <div class="form-group">
          <label for="inputSessionToken">Command Code 会话 Token (Session Token 测试)</label>
          <input type="password" id="inputSessionToken" class="form-control" placeholder="覆盖测试: __Secure-commandcode_prod_.session_token" />
        </div>
        <div class="form-group">
          <label for="inputOpenCodeKeys">OpenCode Go API Keys (测试覆盖)</label>
          <textarea id="inputOpenCodeKeys" class="form-control" rows="3" placeholder="每行一个 sk-..."></textarea>
          <span class="form-hint">每行一个 key；仅当次请求生效，不持久化</span>
        </div>
      </div>
      <div style="margin-top: 14px; display: flex; justify-content: flex-end; gap: 10px;">
        <button id="btnSaveConfig" class="btn btn-primary">保存并重新获取用量</button>
      </div>
    </div>

    <!-- Tab: Command Code -->
    <div id="sectionCommandcode" class="tab-section">
      <div id="ccErrorCard" class="error-card">
        <div class="error-card-title">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
          <span>Command Code 查询失败</span>
        </div>
        <div id="ccErrorMsg" class="error-card-msg">-</div>
      </div>

      <div id="ccContent">
        <div class="cc-tab-head">
          <span class="cc-tab-title">Command Code</span>
          <span id="planBadge" class="plan-tag" style="display:none;">Plan: -</span>
        </div>

    <!-- Overview Metrics -->
    <div class="metrics-grid">
      <div class="metric-card">
        <div class="metric-header">
          <span class="metric-title">月度额度 (Monthly Credits)</span>
          <div class="metric-icon">💰</div>
        </div>
        <div id="valMonthlyCredits" class="metric-value">-</div>
        <div class="metric-desc">当前周期 Command Code 基础配额</div>
      </div>

      <div class="metric-card">
        <div class="metric-header">
          <span class="metric-title">开源奖励额度 (Open Source)</span>
          <div class="metric-icon">🎁</div>
        </div>
        <div id="valOpensourceCredits" class="metric-value">-</div>
        <div class="metric-desc">开源贡献者赠送与奖励额度</div>
      </div>

      <div class="metric-card">
        <div class="metric-header">
          <span class="metric-title">可用额度总计 (Total Credits)</span>
          <div class="metric-icon">✨</div>
        </div>
        <div id="valTotalCredits" class="metric-value">-</div>
        <div class="metric-desc">月度额度与奖励额度汇总</div>
      </div>
    </div>

    <!-- Double Window Limits -->
    <div class="quota-section">
      <!-- Monthly Window -->
      <div id="cardMonthly" class="quota-card">
        <div class="quota-card-header">
          <div>
            <span class="quota-tag" style="color:#f59e0b; background:rgba(245,158,11,0.12)">账单周期</span>
            <div class="quota-name">月度额度 (Monthly Window)</div>
          </div>
          <div id="badgeMonthly" class="quota-percent-badge">- %</div>
        </div>

        <div class="quota-stats-row">
          <div>
            <span id="usedMonthly" class="quota-usage-num">-</span>
            <span id="capMonthly" class="quota-cap-num">/ -</span>
          </div>
          <div>剩余可用: <strong id="remainMonthly">-</strong></div>
        </div>

        <div class="progress-track">
          <div id="barMonthly" class="progress-bar"></div>
        </div>

        <div class="reset-box">
          <div class="reset-label">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
            <span>账单周期重置</span>
          </div>
          <div id="timerMonthly" class="countdown-timer">--:--:--</div>
        </div>
      </div>

      <!-- 5-Hour Window -->
      <div id="cardFiveHour" class="quota-card">
        <div class="quota-card-header">
          <div>
            <span class="quota-tag">短期滑动窗口</span>
            <div class="quota-name">5 小时限制 (5-Hour Window)</div>
          </div>
          <div id="badgeFiveHour" class="quota-percent-badge">- %</div>
        </div>

        <div class="quota-stats-row">
          <div>
            <span id="usedFiveHour" class="quota-usage-num">-</span>
            <span id="capFiveHour" class="quota-cap-num">/ -</span>
          </div>
          <div>剩余可用: <strong id="remainFiveHour">-</strong></div>
        </div>

        <div class="progress-track">
          <div id="barFiveHour" class="progress-bar"></div>
        </div>

        <div class="reset-box">
          <div class="reset-label">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
            <span>重置倒计时</span>
          </div>
          <div id="timerFiveHour" class="countdown-timer">--:--:--</div>
        </div>
      </div>

      <!-- Weekly Window -->
      <div id="cardWeekly" class="quota-card">
        <div class="quota-card-header">
          <div>
            <span class="quota-tag" style="color:#8b5cf6; background:rgba(139,92,246,0.12)">周度窗口</span>
            <div class="quota-name">每周限制 (Weekly Window)</div>
          </div>
          <div id="badgeWeekly" class="quota-percent-badge">- %</div>
        </div>

        <div class="quota-stats-row">
          <div>
            <span id="usedWeekly" class="quota-usage-num">-</span>
            <span id="capWeekly" class="quota-cap-num">/ -</span>
          </div>
          <div>剩余可用: <strong id="remainWeekly">-</strong></div>
        </div>

        <div class="progress-track">
          <div id="barWeekly" class="progress-bar"></div>
        </div>

        <div class="reset-box">
          <div class="reset-label">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
            <span>重置倒计时</span>
          </div>
          <div id="timerWeekly" class="countdown-timer">--:--:--</div>
        </div>
      </div>
    </div>

      </div>
    </div>
    <!-- /Tab: Command Code -->

    <!-- Tab: OpenCode Go -->
    <div id="sectionOpencode" class="tab-section">
      <div id="ocErrorCard" class="error-card">
        <div class="error-card-title">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
          <span>OpenCode Go 查询失败</span>
        </div>
        <div id="ocErrorMsg" class="error-card-msg">-</div>
      </div>
      <div id="ocContent"></div>
    </div>
    <!-- /Tab: OpenCode Go -->

    <!-- Tab: All -->
    <div id="sectionAll" class="tab-section active">
      <!-- All tab: vertical groups, one group title per provider -->
      <div class="all-group-title">Command Code</div>
      <div class="all-provider-card">
        <div class="all-provider-head">
          <span id="allBadgeCommandcode" class="status-badge"><span class="status-dot"></span><span id="allBadgeTextCommandcode">-</span></span>
        </div>
        <div id="allBodyCommandcode" class="all-provider-body">尚未加载</div>
      </div>

      <div class="all-group-title">OpenCode Go</div>
      <div class="all-provider-card">
        <div class="all-provider-head">
          <span id="allBadgeOpencode" class="status-badge"><span class="status-dot"></span><span id="allBadgeTextOpencode">-</span></span>
        </div>
        <div id="allBodyOpencode" class="all-provider-body">尚未加载</div>
      </div>
    </div>
    <!-- /Tab: All -->

    <!-- Footer Status -->
    <div class="footer-bar">
      <div>最后同步时间: <span id="lastUpdated">-</span></div>
      <div class="footer-links">
        <span id="authInfo">Provider: commandcode + opencode_go</span>
      </div>
    </div>
  </div>

  <script>
    (function () {
      const ALL_ENDPOINT = "/v0/management/plugins/commandcode/all";

      // Elements
      const btnRefresh = document.getElementById("btnRefresh");
      const refreshIcon = document.getElementById("refreshIcon");
      const btnSettings = document.getElementById("btnSettings");
      const btnCloseSettings = document.getElementById("btnCloseSettings");
      const btnSaveConfig = document.getElementById("btnSaveConfig");
      const settingsDrawer = document.getElementById("settingsDrawer");
      const inputMgmtKey = document.getElementById("inputMgmtKey");
      const inputSessionToken = document.getElementById("inputSessionToken");
      const inputOpenCodeKeys = document.getElementById("inputOpenCodeKeys");
      const alertBox = document.getElementById("alertBox");
      const alertMsg = document.getElementById("alertMsg");
      const planBadge = document.getElementById("planBadge");

      // Tab sections and per-provider error cards
      const ccContent = document.getElementById("ccContent");
      const ccErrorCard = document.getElementById("ccErrorCard");
      const ccErrorMsg = document.getElementById("ccErrorMsg");
      const ocContent = document.getElementById("ocContent");
      const ocErrorCard = document.getElementById("ocErrorCard");
      const ocErrorMsg = document.getElementById("ocErrorMsg");

      // All tab elements
      const allBodyCommandcode = document.getElementById("allBodyCommandcode");
      const allBadgeCommandcode = document.getElementById("allBadgeCommandcode");
      const allBadgeTextCommandcode = document.getElementById("allBadgeTextCommandcode");
      const allBodyOpencode = document.getElementById("allBodyOpencode");
      const allBadgeOpencode = document.getElementById("allBadgeOpencode");
      const allBadgeTextOpencode = document.getElementById("allBadgeTextOpencode");

      const valMonthlyCredits = document.getElementById("valMonthlyCredits");
      const valOpensourceCredits = document.getElementById("valOpensourceCredits");
      const valTotalCredits = document.getElementById("valTotalCredits");

      const cardMonthly = document.getElementById("cardMonthly");
      const badgeMonthly = document.getElementById("badgeMonthly");
      const usedMonthly = document.getElementById("usedMonthly");
      const capMonthly = document.getElementById("capMonthly");
      const remainMonthly = document.getElementById("remainMonthly");
      const barMonthly = document.getElementById("barMonthly");
      const timerMonthly = document.getElementById("timerMonthly");

      const cardFiveHour = document.getElementById("cardFiveHour");
      const badgeFiveHour = document.getElementById("badgeFiveHour");
      const usedFiveHour = document.getElementById("usedFiveHour");
      const capFiveHour = document.getElementById("capFiveHour");
      const remainFiveHour = document.getElementById("remainFiveHour");
      const barFiveHour = document.getElementById("barFiveHour");
      const timerFiveHour = document.getElementById("timerFiveHour");

      const cardWeekly = document.getElementById("cardWeekly");
      const badgeWeekly = document.getElementById("badgeWeekly");
      const usedWeekly = document.getElementById("usedWeekly");
      const capWeekly = document.getElementById("capWeekly");
      const remainWeekly = document.getElementById("remainWeekly");
      const barWeekly = document.getElementById("barWeekly");
      const timerWeekly = document.getElementById("timerWeekly");
      const lastUpdated = document.getElementById("lastUpdated");

      let fiveHourTargetTime = null;
      let monthlyTargetTime = null;
      
      let weeklyTargetTime = null;
      let timerInterval = null;

      // Per-provider render state (drives tab badge aggregation)
      const providerState = {
        commandcode: { level: "unknown", err: null, data: null },
        opencode: { level: "unknown", err: null, data: null }
      };
      let activeTab = "all";

      function getStoredManagementKey() {
        if (inputMgmtKey.value.trim()) {
          return inputMgmtKey.value.trim();
        }
        const keys = ["management_key", "managementKey", "cpa_management_key", "token"];
        for (const k of keys) {
          const val = localStorage.getItem(k);
          if (val && val.trim()) {
            return val.trim();
          }
        }
        return "";
      }

      function showAlert(message, isWarning = false) {
        alertMsg.textContent = message;
        alertBox.className = isWarning ? "alert alert-warning show" : "alert alert-danger show";
      }

      function hideAlert() {
        alertBox.className = "alert";
      }

      // Command Code 额度按美元(USD)计价：金额统一加 $ 前缀，按美元格式输出（千分位 + 两位小数）
      function formatUSD(num) {
        if (num === null || num === undefined || isNaN(num)) return "$0.00";
        return Number(num).toLocaleString("en-US", {
          style: "currency",
          currency: "USD",
          minimumFractionDigits: 2,
          maximumFractionDigits: 2
        });
      }

      function formatCountdown(targetDate) {
        if (!targetDate) return "--:--:--";
        const now = new Date().getTime();
        const target = targetDate.getTime();
        const diff = target - now;
        if (diff <= 0) return "已就绪 (可重置)";

        const days = Math.floor(diff / (1000 * 60 * 60 * 24));
        const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
        const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
        const seconds = Math.floor((diff % (1000 * 60)) / 1000);

        if (days > 0) {
          return days + "天 " + pad(hours) + ":" + pad(minutes) + ":" + pad(seconds);
        }
        return pad(hours) + ":" + pad(minutes) + ":" + pad(seconds);
      }

      function pad(n) {
        return n < 10 ? "0" + n : n;
      }

      function esc(s) {
        return String(s)
          .replace(/&/g, "&amp;")
          .replace(/</g, "&lt;")
          .replace(/>/g, "&gt;")
          .replace(/"/g, "&quot;");
      }

      // Returns a clamped number in [0,100], or null when missing/invalid
      function clampPercent(v) {
        if (v === null || v === undefined || v === "") return null;
        const n = Number(v);
        if (!isFinite(n)) return null;
        return Math.min(100, Math.max(0, n));
      }

      // Unknown status values never error: only explicit exceeded flag,
      // status "exceeded", or percent >= 100 count as exceeded
      function levelOf(pct, exceeded, status) {
        if (exceeded || status === "exceeded" || (pct !== null && pct >= 100)) return "exceeded";
        if (pct !== null && pct >= 80) return "warning";
        return "online";
      }

      const LEVEL_RANK = { unknown: -1, online: 0, warning: 1, error: 2, exceeded: 3 };
      const BADGE_TEXT = {
        unknown: "正在检查...",
        online: "正常运行 (Normal)",
        warning: "配额紧张 (Warning)",
        exceeded: "已达限额 (Exceeded)",
        error: "查询错误"
      };

      function badgeVisualClass(level) {
        // error reuses the danger/exceeded look
        return level === "error" ? "exceeded" : level === "unknown" ? "" : level;
      }

      function setProviderStatus(provider, level) {
        providerState[provider].level = level;
        providerState[provider].err = null;
      }

      function summaryRow(label, value) {
        return '<div class="all-summary-row"><span class="all-summary-label">' + label + '</span><span class="all-summary-value">' + value + '</span></div>';
      }

      function miniBar(pct, level) {
        const cls = level === "exceeded" ? " danger" : level === "warning" ? " warning" : "";
        const w = pct === null ? 0 : pct;
        return '<div class="progress-track"><div class="progress-bar' + cls + '" style="width:' + w + '%"></div></div>';
      }

      function showProviderError(provider, msg) {
        providerState[provider].level = "error";
        providerState[provider].err = msg;
        providerState[provider].data = null;
        if (provider === "commandcode") {
          ccContent.style.display = "none";
          ccErrorMsg.textContent = msg;
          ccErrorCard.classList.add("show");
        } else {
          ocContent.style.display = "none";
          ocErrorMsg.textContent = msg;
          ocErrorCard.classList.add("show");
        }
      }

      // Multi-window countdown elements carry their reset info in data
      // attributes (data-reset-at ISO string, or data-reset-secs seconds),
      // so updateTimers() can walk every tab容器's .oc-win-reset uniformly
      function updateTimers() {
        if (monthlyTargetTime) {
          timerMonthly.textContent = formatCountdown(monthlyTargetTime);
        }
        if (fiveHourTargetTime) {
          timerFiveHour.textContent = formatCountdown(fiveHourTargetTime);
        }
        if (weeklyTargetTime) {
          timerWeekly.textContent = formatCountdown(weeklyTargetTime);
        }
        const allResetEls = document.querySelectorAll(".oc-win-reset");
        for (let i = 0; i < allResetEls.length; i++) {
          const el = allResetEls[i];
          let target = null;
          const at = el.getAttribute("data-reset-at");
          if (at) {
            const t = new Date(at);
            if (!isNaN(t.getTime())) target = t;
          }
          if (!target) {
            const secs = Number(el.getAttribute("data-reset-secs"));
            if (secs > 0) target = new Date(Date.now() + secs * 1000);
          }
          el.textContent = target ? formatCountdown(target) : "-";
        }
      }

      function renderUsage(data) {
        providerState.commandcode.err = null;
        providerState.commandcode.data = data;
        ccContent.style.display = "";
        ccErrorCard.classList.remove("show");
        const credits = data.credits || (data.data && data.data.credits) || {};
        const limits = data.window_limits || (data.data && data.data.window_limits) || {};

        // Plan
        const plan = data.plan || (data.data && data.data.plan);
        if (plan) {
          const planName = plan.name || (typeof plan === "string" ? plan : "Unknown");
          planBadge.textContent = "Plan: " + planName;
          planBadge.style.display = "inline-flex";
        } else {
          planBadge.style.display = "none";
        }

        // Credits
        valMonthlyCredits.textContent = formatUSD(credits.monthly_credits);
        valOpensourceCredits.textContent = formatUSD(credits.opensource_monthly_credits);
        valTotalCredits.textContent = formatUSD(credits.total_credits);

        // Monthly (billing period)
        const monthly = limits.monthly || {};
        const pMonth = Math.min(100, Math.max(0, monthly.percentage || 0));
        badgeMonthly.textContent = pMonth.toFixed(1) + "%";
        usedMonthly.textContent = formatUSD(monthly.used);
        capMonthly.textContent = "/ " + formatUSD(monthly.cap);
        remainMonthly.textContent = formatUSD(monthly.remaining);
        barMonthly.style.width = pMonth + "%";

        barMonthly.className = "progress-bar" + (pMonth >= 90 || monthly.exceeded ? " danger" : pMonth >= 70 ? " warning" : "");
        cardMonthly.className = "quota-card" + (monthly.exceeded ? " is-exceeded" : "");
        monthlyTargetTime = null;
        timerMonthly.textContent = "账单周期";

        // Five Hour
        const fiveHour = limits.five_hour || {};
        const pFive = Math.min(100, Math.max(0, fiveHour.percentage || 0));
        badgeFiveHour.textContent = pFive.toFixed(1) + "%";
        usedFiveHour.textContent = formatUSD(fiveHour.used);
        capFiveHour.textContent = "/ " + formatUSD(fiveHour.cap);
        remainFiveHour.textContent = formatUSD(fiveHour.remaining);
        barFiveHour.style.width = pFive + "%";

        barFiveHour.className = "progress-bar" + (pFive >= 90 || fiveHour.exceeded ? " danger" : pFive >= 70 ? " warning" : "");
        cardFiveHour.className = "quota-card" + (fiveHour.exceeded ? " is-exceeded" : "");

        if (fiveHour.reset_at) {
          fiveHourTargetTime = new Date(fiveHour.reset_at);
        } else if (fiveHour.reset_in_seconds) {
          fiveHourTargetTime = new Date(Date.now() + fiveHour.reset_in_seconds * 1000);
        } else {
          fiveHourTargetTime = null;
        }

        // Weekly
        const weekly = limits.weekly || {};
        const pWeek = Math.min(100, Math.max(0, weekly.percentage || 0));
        badgeWeekly.textContent = pWeek.toFixed(1) + "%";
        usedWeekly.textContent = formatUSD(weekly.used);
        capWeekly.textContent = "/ " + formatUSD(weekly.cap);
        remainWeekly.textContent = formatUSD(weekly.remaining);
        barWeekly.style.width = pWeek + "%";

        barWeekly.className = "progress-bar" + (pWeek >= 90 || weekly.exceeded ? " danger" : pWeek >= 70 ? " warning" : "");
        cardWeekly.className = "quota-card" + (weekly.exceeded ? " is-exceeded" : "");

        if (weekly.reset_at) {
          weeklyTargetTime = new Date(weekly.reset_at);
        } else if (weekly.reset_in_seconds) {
          weeklyTargetTime = new Date(Date.now() + weekly.reset_in_seconds * 1000);
        } else {
          weeklyTargetTime = null;
        }

        // Per-provider status; consumed by All tab provider badges
        let ccLevel;
        if (fiveHour.exceeded || weekly.exceeded || monthly.exceeded) {
          ccLevel = "exceeded";
        } else if (pFive >= 80 || pWeek >= 80 || pMonth >= 80) {
          ccLevel = "warning";
        } else {
          ccLevel = "online";
        }
        setProviderStatus("commandcode", ccLevel);

        updateTimers();
      }
      function renderOpencode(data) {
        providerState.opencode.err = null;
        providerState.opencode.data = data;

        // 多 key 契约：keys[] 为空/缺失视为 provider 级失败，走全局错误卡片
        const keys = data && Array.isArray(data.keys) ? data.keys : [];
        if (keys.length === 0) {
          showProviderError("opencode", (data && data.error) || "OpenCode 未返回任何 key 数据");
          return;
        }

        ocContent.style.display = "";
        ocErrorCard.classList.remove("show");

        // All tab 的 OpenCode 组与本 tab 共用同一套逐 key 渲染逻辑
        const rendered = renderOpenCodeKeyGroups(keys);
        ocContent.innerHTML = rendered.html;
        updateTimers();

        setProviderStatus("opencode", rendered.worst);
      }

      // Shared per-key group renderer: the OpenCode tab and the All tab's
      // OpenCode group both consume this to avoid logic drift. Returns
      // { html, worst }. Countdown info is stamped into data-reset-at (ISO
      // string) or data-reset-secs (seconds) attributes on .oc-win-reset
      // elements, so updateTimers() uniformly walks every tab container.
      function renderOpenCodeKeyGroups(keys) {
        const WIN_DEFS = [
          { name: "Rolling 5h", key: "rolling" },
          { name: "Weekly", key: "weekly" },
          { name: "Monthly", key: "monthly" }
        ];
        const rank = { online: 0, warning: 1, exceeded: 2 };
        // provider 级别 = 全部 key 最差（失败 key 按 exceeded 视觉计入）
        let worst = "online";
        let html = "";

        keys.forEach(function (k) {
          const keyId = esc(k.key_id || "***");
          if (k.ok && k.windows) {
            const windows = k.windows;
            let keyWorst = "online";
            let rowsHtml = "";

            WIN_DEFS.forEach(function (def) {
              const w = windows[def.key] || {};
              const pct = clampPercent(w.percent);
              const level = levelOf(pct, w.exceeded, w.status);
              if (rank[level] > rank[keyWorst]) keyWorst = level;
              if (rank[level] > rank[worst]) worst = level;

              // reset_at 优先；缺失/不可解析时回退 reset_in_seconds；皆无显 "-"
              let resetAt = "";
              let resetSecs = "";
              if (w.reset_at) {
                const t = new Date(w.reset_at);
                if (!isNaN(t.getTime())) resetAt = t.toISOString();
              }
              if (!resetAt && w.reset_in_seconds > 0) {
                resetSecs = String(w.reset_in_seconds);
              }
              const resetAttr = resetAt
                ? " data-reset-at=\"" + esc(resetAt) + "\""
                : (resetSecs ? " data-reset-secs=\"" + esc(resetSecs) + "\"" : "");

              const pctText = pct === null ? "-" : Math.round(pct) + "%";
              const pctCls = level === "exceeded" ? " bad" : pct !== null && pct >= 80 ? " warn" : "";
              const barCls = level === "exceeded" ? " danger" : level === "warning" ? " warning" : "";

              rowsHtml += '<div class="oc-win-row">' +
                '<span class="oc-win-name">' + def.name + '</span>' +
                '<div class="progress-track oc-win-bar"><div class="progress-bar' + barCls + '" style="width:' + (pct === null ? 0 : pct) + '%"></div></div>' +
                '<span class="oc-win-pct' + pctCls + '">' + pctText + '</span>' +
                '<span class="oc-win-reset countdown-timer"' + resetAttr + '>-</span>' +
                '</div>';
            });

            const chipText = keyWorst === "exceeded" ? "超限" : keyWorst === "warning" ? "紧张" : "正常";
            html += '<div class="oc-key-group">' +
              '<div class="oc-key-head"><span class="oc-key-id">' + keyId + '</span>' +
              '<span class="oc-key-chip status-badge ' + keyWorst + '"><span class="status-dot"></span><span>' + chipText + '</span></span></div>' +
              '<div class="oc-key-body">' + rowsHtml + '</div>' +
              '</div>';
          } else {
            // 失败 key 无 windows 字段：只渲染 key 头 + 错误行
            if (rank.exceeded > rank[worst]) worst = "exceeded";
            const is401 = k.status_code === 401;
            const label = is401 ? "凭据无效" : "查询错误";
            const statusPart = k.status_code ? "上游 " + esc(String(k.status_code)) : "查询失败";
            const errTail = k.error ? "：" + esc(k.error) : "";
            const tip = is401 ? "，请检查该 key 或从配置中移除" : "";
            html += '<div class="oc-key-group">' +
              '<div class="oc-key-head"><span class="oc-key-id">' + keyId + '</span>' +
              '<span class="oc-key-chip status-badge exceeded"><span class="status-dot"></span><span>' + label + '</span></span></div>' +
              '<div class="oc-key-error">' + label + '（' + statusPart + '）' + errTail + tip + '</div>' +
              '</div>';
          }
        });

        return { html: html, worst: worst };
      }

      function renderAllTab() {
        const cc = providerState.commandcode;
        const oc = providerState.opencode;

        // Command Code summary card
        if (cc.err) {
          allBadgeCommandcode.className = "status-badge exceeded";
          allBadgeTextCommandcode.textContent = "查询错误";
          allBodyCommandcode.innerHTML = '<div class="error-card-title">Command Code 查询失败</div><div class="error-card-msg">' + esc(cc.err) + '</div>';
        } else if (cc.data) {
          const data = cc.data;
          const credits = data.credits || (data.data && data.data.credits) || {};
          const limits = data.window_limits || (data.data && data.data.window_limits) || {};
          const plan = data.plan || (data.data && data.data.plan);
          const planName = plan ? (plan.name || (typeof plan === "string" ? plan : "Unknown")) : "-";
          const windowsDef = [
            { name: "Monthly", o: limits.monthly || {} },
            { name: "5-Hour", o: limits.five_hour || {} },
            { name: "Weekly", o: limits.weekly || {} }
          ];
          let worstName = "-";
          let worstPct = null;
          let worstLevel = "online";
          windowsDef.forEach(function (d) {
            const pct = clampPercent(d.o.percentage);
            const level = levelOf(pct, d.o.exceeded, "");
            if (worstPct === null || (pct !== null && pct > worstPct) || level === "exceeded") {
              worstName = d.name;
              worstPct = pct === null ? 0 : pct;
              worstLevel = level;
            }
          });
          allBadgeCommandcode.className = "status-badge " + badgeVisualClass(cc.level);
          allBadgeTextCommandcode.textContent = BADGE_TEXT[cc.level] || "-";
          // Plan + 三个限额窗口（百分比/进度条/重置时间），行结构与 OpenCode 卡片一致，
          // 倒计时走 data-reset-at/data-reset-secs 属性，由 updateTimers() 统一刷新
          const ccWins = [
            { name: "Rolling 5h", o: limits.five_hour || {} },
            { name: "Weekly", o: limits.weekly || {} },
            { name: "Monthly", o: limits.monthly || {} }
          ];
          let ccRows = "";
          ccWins.forEach(function (d) {
            const w = d.o;
            const pct = clampPercent(w.percentage);
            const level = levelOf(pct, w.exceeded, "");
            let resetAt = "";
            let resetSecs = "";
            if (w.reset_at) {
              const t = new Date(w.reset_at);
              if (!isNaN(t.getTime())) resetAt = t.toISOString();
            }
            if (!resetAt && w.reset_in_seconds > 0) {
              resetSecs = String(w.reset_in_seconds);
            }
            const resetAttr = resetAt
              ? " data-reset-at=\"" + esc(resetAt) + "\""
              : (resetSecs ? " data-reset-secs=\"" + esc(resetSecs) + "\"" : "");
            const pctText = pct === null ? "-" : Math.round(pct) + "%";
            const pctCls = level === "exceeded" ? " bad" : pct !== null && pct >= 80 ? " warn" : "";
            const barCls = level === "exceeded" ? " danger" : level === "warning" ? " warning" : "";
            ccRows += '<div class="oc-win-row">' +
              '<span class="oc-win-name">' + d.name + '</span>' +
              '<div class="progress-track oc-win-bar"><div class="progress-bar' + barCls + '" style="width:' + (pct === null ? 0 : pct) + '%"></div></div>' +
              '<span class="oc-win-pct' + pctCls + '">' + pctText + '</span>' +
              '<span class="oc-win-reset countdown-timer"' + resetAttr + '>-</span>' +
              '</div>';
          });
          allBodyCommandcode.innerHTML =
            summaryRow("Plan", esc(planName)) +
            '<div class="oc-key-body">' + ccRows + '</div>';
        } else {
          allBadgeCommandcode.className = "status-badge";
          allBadgeTextCommandcode.textContent = "尚未加载";
          allBodyCommandcode.innerHTML = '<div class="all-summary-value">尚未加载</div>';
        }

        // OpenCode Go summary card
        if (oc.err) {
          allBadgeOpencode.className = "status-badge exceeded";
          allBadgeTextOpencode.textContent = "查询错误";
          allBodyOpencode.innerHTML = '<div class="error-card-title">OpenCode Go 查询失败</div><div class="error-card-msg">' + esc(oc.err) + '</div>';
        } else if (oc.data) {
          const data = oc.data;
          // 多 key 契约：逐 key 完整卡片，与 OpenCode tab 共用同一渲染函数
          const keys = Array.isArray(data.keys) ? data.keys : [];
          const rendered = renderOpenCodeKeyGroups(keys);
          const html = rendered.html || '<div class="all-summary-value">无 key 数据</div>';
          allBadgeOpencode.className = "status-badge " + badgeVisualClass(oc.level);
          allBadgeTextOpencode.textContent = BADGE_TEXT[oc.level] || "-";
          allBodyOpencode.innerHTML = html;
        } else {
          allBadgeOpencode.className = "status-badge";
          allBadgeTextOpencode.textContent = "尚未加载";
          allBodyOpencode.innerHTML = '<div class="all-summary-value">尚未加载</div>';
        }
      }

      function setActiveTab(tab, updateHash) {
        activeTab = tab;
        const tabBtns = document.querySelectorAll(".tab-btn");
        for (let i = 0; i < tabBtns.length; i++) {
          tabBtns[i].classList.toggle("active", tabBtns[i].getAttribute("data-tab") === tab);
        }
        document.getElementById("sectionCommandcode").classList.toggle("active", tab === "commandcode");
        document.getElementById("sectionOpencode").classList.toggle("active", tab === "opencode");
        document.getElementById("sectionAll").classList.toggle("active", tab === "all");
        if (updateHash) {
          if (tab === "all") {
            // All 为默认 tab：激活 All 时清掉 hash（刷新回到默认）
            history.replaceState(null, "", location.pathname + location.search);
          } else {
            location.hash = tab;
          }
        }
      }

      async function fetchUsage() {
        refreshIcon.classList.add("spin");
        btnRefresh.disabled = true;

        const mgmtKey = getStoredManagementKey();
        const overrideToken = inputSessionToken.value.trim();
        const overrideOpenCodeKeys = inputOpenCodeKeys.value.split("\n").map(function (s) { return s.trim(); }).filter(Boolean);

        const headers = {
          "Accept": "application/json",
          "Content-Type": "application/json"
        };
        if (mgmtKey) {
          headers["Authorization"] = "Bearer " + mgmtKey;
          headers["X-Management-Key"] = mgmtKey;
        }

        // 整页只发一次 /all 请求，一次拿两个 provider；
        // 覆盖凭据仅在填写时才进 body，不持久化
        const bodyObj = {};
        if (overrideToken) {
          bodyObj.session_token = overrideToken;
        }
        // 非空才进 body：list 优先（后端仍接受旧 scalar 字段，前端不再发送）
        if (overrideOpenCodeKeys.length > 0) {
          bodyObj.opencode_api_keys = overrideOpenCodeKeys;
        }

        try {
          const res = await fetch(ALL_ENDPOINT, {
            method: "POST",
            headers: headers,
            body: JSON.stringify(bodyObj)
          });

          if (res.status === 401 || res.status === 403) {
            settingsDrawer.classList.add("open");
            showAlert("需要 CLIProxyAPI 管理密钥 (401/403)。请在上方输入框填入 Management Key 并保存。", true);
            return;
          }

          const json = await res.json();
          if (!res.ok) {
            const msg = (json && (json.error || json.message)) || "获取配额失败 (HTTP " + res.status + ")";
            showProviderError("commandcode", msg);
            showProviderError("opencode", msg);
            showAlert(msg);
            return;
          }

          // 部分失败不阻塞：缺失/失败 provider 在各自 tab 内渲染错误卡片
          let anyOk = false;
          if (json.commandcode && json.commandcode.ok !== false) {
            renderUsage(json.commandcode);
            anyOk = true;
          } else {
            showProviderError("commandcode", (json.errors && json.errors.commandcode) || "Command Code 查询失败");
          }
          if (json.opencode && json.opencode.ok !== false) {
            renderOpencode(json.opencode);
            anyOk = true;
          } else {
            showProviderError("opencode", (json.errors && json.errors.opencode) || "OpenCode Go 查询失败");
          }

          if (anyOk) {
            hideAlert();
          } else {
            showAlert("所有数据源查询失败，请检查配置或凭据。");
          }

          const updatedStr = json.updated_at ||
            (json.commandcode && json.commandcode.updated_at) ||
            (json.opencode && json.opencode.updated_at);
          lastUpdated.textContent = updatedStr ? new Date(updatedStr).toLocaleString() : new Date().toLocaleString();

          renderAllTab();
          updateTimers();
        } catch (err) {
          showProviderError("commandcode", "网络或同源请求错误: " + err.message);
          showProviderError("opencode", "网络或同源请求错误: " + err.message);
          showAlert("网络或同源请求错误: " + err.message);
        } finally {
          refreshIcon.classList.remove("spin");
          btnRefresh.disabled = false;
        }
      }

      // Event Listeners
      btnRefresh.addEventListener("click", fetchUsage);

      btnSettings.addEventListener("click", () => {
        settingsDrawer.classList.toggle("open");
      });

      btnCloseSettings.addEventListener("click", () => {
        settingsDrawer.classList.remove("open");
      });

      btnSaveConfig.addEventListener("click", () => {
        const key = inputMgmtKey.value.trim();
        if (key) {
          localStorage.setItem("management_key", key);
        }
        fetchUsage();
      });

      // Tab switching + hash persistence (#opencode / #all)
      document.querySelectorAll(".tab-btn").forEach((btn) => {
        btn.addEventListener("click", () => {
          setActiveTab(btn.getAttribute("data-tab"), true);
        });
      });

      // Init on load
      const initKey = localStorage.getItem("management_key") || localStorage.getItem("cpa_management_key");
      if (initKey) {
        inputMgmtKey.value = initKey;
      }

      function syncTheme() {
        let themeSetting = null;

        // 1. Try reading from parent window (if same-origin iframe)
        try {
          if (window.parent && window.parent !== window && window.parent.document) {
            const parentTheme = window.parent.document.documentElement.getAttribute("data-theme");
            if (parentTheme) {
              themeSetting = parentTheme;
            }
          }
        } catch (e) {}

        // 2. Try reading from localStorage['cli-proxy-theme']
        if (!themeSetting) {
          try {
            const stored = localStorage.getItem("cli-proxy-theme");
            if (stored) {
              const parsed = JSON.parse(stored);
              if (parsed && parsed.state) {
                if (parsed.state.theme) {
                  themeSetting = parsed.state.theme;
                }
                if (parsed.state.resolvedTheme && themeSetting === "auto") {
                  themeSetting = parsed.state.resolvedTheme;
                }
              }
            }
          } catch (e) {}
        }

        // 3. Fallback to own html attribute
        if (!themeSetting) {
          themeSetting = document.documentElement.getAttribute("data-theme");
        }

        // Apply theme to documentElement
        const docEl = document.documentElement;
        if (themeSetting === "dark") {
          docEl.setAttribute("data-theme", "dark");
        } else if (themeSetting === "white" || themeSetting === "light") {
          docEl.setAttribute("data-theme", themeSetting);
        } else {
          // auto or unset: check system preference
          const isDark = window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches;
          if (isDark) {
            docEl.setAttribute("data-theme", "dark");
          } else {
            docEl.setAttribute("data-theme", "light");
          }
        }
      }

      // Initialize theme and sync listeners
      syncTheme();
      window.addEventListener("storage", (e) => {
        if (e.key === "cli-proxy-theme") {
          syncTheme();
        }
      });
      if (window.matchMedia) {
        window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", syncTheme);
      }
      setInterval(syncTheme, 2000);

      if (!timerInterval) {
        timerInterval = setInterval(updateTimers, 1000);
      }

      // Restore tab from location.hash, then initial fetch
      // 无 hash 默认 All：All tab 才能通过刷新后的 #all hash 恢复
      const initHash = location.hash.replace(/^#/, "");
      if (initHash === "opencode" || initHash === "commandcode") {
        setActiveTab(initHash, false);
      } else {
        setActiveTab("all", false);
      }

      // Initial fetch
      fetchUsage();
    })();
  </script>
</body>
</html>
`

// GetQuotaPageHTML returns the raw HTML bytes.
func GetQuotaPageHTML() []byte {
	return []byte(QuotaPageHTML)
}

// GetQuotaPageBase64 returns the base64-encoded Quota page HTML.
func GetQuotaPageBase64() string {
	return base64.StdEncoding.EncodeToString([]byte(QuotaPageHTML))
}
