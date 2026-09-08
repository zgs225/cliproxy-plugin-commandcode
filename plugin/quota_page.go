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
  <title>Command Code 配额与用量 - CLIProxyAPI</title>
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

    .brand-subtitle {
      font-size: 13px;
      color: var(--text-muted);
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
            Command Code 配额
            <span class="version-tag">v0.2.1</span>
            <span id="planBadge" class="plan-tag" style="display:none;">Plan: -</span>
          </div>
          <div class="brand-subtitle">CLIProxyAPI 实时限额与 Credits 用量监控</div>
        </div>
      </div>

      <div class="action-group">
        <div id="statusBadge" class="status-badge">
          <span class="status-dot"></span>
          <span id="statusText">正在检查...</span>
        </div>

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
      </div>
      <div style="margin-top: 14px; display: flex; justify-content: flex-end; gap: 10px;">
        <button id="btnSaveConfig" class="btn btn-primary">保存并重新获取用量</button>
      </div>
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

    <!-- Footer Status -->
    <div class="footer-bar">
      <div>最后同步时间: <span id="lastUpdated">-</span></div>
      <div class="footer-links">
        <span id="authInfo">Provider: commandcode</span>
      </div>
    </div>
  </div>

  <script>
    (function () {
      const USAGE_ENDPOINT = "/v0/management/plugins/commandcode/usage";

      // Elements
      const btnRefresh = document.getElementById("btnRefresh");
      const refreshIcon = document.getElementById("refreshIcon");
      const btnSettings = document.getElementById("btnSettings");
      const btnCloseSettings = document.getElementById("btnCloseSettings");
      const btnSaveConfig = document.getElementById("btnSaveConfig");
      const settingsDrawer = document.getElementById("settingsDrawer");
      const inputMgmtKey = document.getElementById("inputMgmtKey");
      const inputSessionToken = document.getElementById("inputSessionToken");
      const alertBox = document.getElementById("alertBox");
      const alertMsg = document.getElementById("alertMsg");
      const statusBadge = document.getElementById("statusBadge");
      const statusText = document.getElementById("statusText");
      const planBadge = document.getElementById("planBadge");

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

      let monthlyTargetTime = null;
      let fiveHourTargetTime = null;
      let weeklyTargetTime = null;
      let timerInterval = null;

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

      function formatNumber(num) {
        if (num === null || num === undefined || isNaN(num)) return "0";
        return Number(num).toLocaleString(undefined, { maximumFractionDigits: 2 });
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
      }

      function renderUsage(data) {
        hideAlert();
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
        valMonthlyCredits.textContent = formatNumber(credits.monthly_credits);
        valOpensourceCredits.textContent = formatNumber(credits.opensource_monthly_credits);
        valTotalCredits.textContent = formatNumber(credits.total_credits);

        // Monthly (billing period)
        const monthly = limits.monthly || {};
        const pMonth = Math.min(100, Math.max(0, monthly.percentage || 0));
        badgeMonthly.textContent = pMonth.toFixed(1) + "%";
        usedMonthly.textContent = formatNumber(monthly.used);
        capMonthly.textContent = "/ " + formatNumber(monthly.cap);
        remainMonthly.textContent = formatNumber(monthly.remaining);
        barMonthly.style.width = pMonth + "%";

        barMonthly.className = "progress-bar" + (pMonth >= 90 || monthly.exceeded ? " danger" : pMonth >= 70 ? " warning" : "");
        cardMonthly.className = "quota-card" + (monthly.exceeded ? " is-exceeded" : "");
        monthlyTargetTime = null;
        timerMonthly.textContent = "账单周期";

        // Five Hour
        const fiveHour = limits.five_hour || {};
        const pFive = Math.min(100, Math.max(0, fiveHour.percentage || 0));
        badgeFiveHour.textContent = pFive.toFixed(1) + "%";
        usedFiveHour.textContent = formatNumber(fiveHour.used);
        capFiveHour.textContent = "/ " + formatNumber(fiveHour.cap);
        remainFiveHour.textContent = formatNumber(fiveHour.remaining);
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
        usedWeekly.textContent = formatNumber(weekly.used);
        capWeekly.textContent = "/ " + formatNumber(weekly.cap);
        remainWeekly.textContent = formatNumber(weekly.remaining);
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

        // Overall Status
        if (fiveHour.exceeded || weekly.exceeded || monthly.exceeded) {
          statusBadge.className = "status-badge exceeded";
          statusText.textContent = "已达限额 (Exceeded)";
        } else if (pFive >= 80 || pWeek >= 80 || pMonth >= 80) {
          statusBadge.className = "status-badge warning";
          statusText.textContent = "配额紧张 (Warning)";
        } else {
          statusBadge.className = "status-badge online";
          statusText.textContent = "正常运行 (Normal)";
        }

        const updatedAtStr = data.updated_at || (data.data && data.data.updated_at);
        if (updatedAtStr) {
          lastUpdated.textContent = new Date(updatedAtStr).toLocaleString();
        } else {
          lastUpdated.textContent = new Date().toLocaleString();
        }

        updateTimers();
      }

      async function fetchUsage() {
        refreshIcon.classList.add("spin");
        btnRefresh.disabled = true;

        const mgmtKey = getStoredManagementKey();
        const overrideToken = inputSessionToken.value.trim();

        const headers = {
          "Accept": "application/json"
        };
        if (mgmtKey) {
          headers["Authorization"] = "Bearer " + mgmtKey;
          headers["X-Management-Key"] = mgmtKey;
        }

        let method = "GET";
        let body = null;
        if (overrideToken) {
          method = "POST";
          headers["Content-Type"] = "application/json";
          body = JSON.stringify({ session_token: overrideToken });
        }

        try {
          const res = await fetch(USAGE_ENDPOINT, {
            method: method,
            headers: headers,
            body: body
          });

          if (res.status === 401 || res.status === 403) {
            settingsDrawer.classList.add("open");
            showAlert("需要 CLIProxyAPI 管理密钥 (401/403)。请在上方输入框填入 Management Key 并保存。", true);
            statusBadge.className = "status-badge warning";
            statusText.textContent = "未授权";
            return;
          }

          const json = await res.json();
          if (!res.ok || json.ok === false) {
            const msg = json.error || (json.message ? json.message : "获取配额失败 (HTTP " + res.status + ")");
            showAlert(msg);
            statusBadge.className = "status-badge exceeded";
            statusText.textContent = "查询错误";
            return;
          }

          renderUsage(json);
        } catch (err) {
          showAlert("网络或同源请求错误: " + err.message);
          statusBadge.className = "status-badge exceeded";
          statusText.textContent = "连接失败";
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
