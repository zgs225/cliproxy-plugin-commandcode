# CLIProxyAPI Command Code Plugin (`commandcode`)

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-blue.svg)](https://golang.org)
[![CLIProxyAPI Plugin ABI](https://img.shields.io/badge/C%20ABI-v1-emerald.svg)](https://help.router-for.me/plugin/development.html)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

[CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) 动态 C ABI 插件，用于提供 **Command Code** 上游配额与窗口限额查询、以及嵌入式配额监控仪表盘卡片（QuotaCard）。

---

## 目录

- [功能特性](#功能特性)
- [系统架构](#系统架构)
- [快速开始](#快速开始)
  - [构建插件](#构建插件)
  - [安装与目录结构](#安装与目录结构)
  - [宿主配置 (`config.yaml`)](#宿主配置-configyaml)
- [管理端点与资源页](#管理端点与资源页)
  - [1. 浏览器资源页 (`QuotaCard`)](#1-浏览器资源页-quotacard)
  - [2. 管理 API: 查询用量 (`GET`)](#2-管理-api-查询用量-get)
  - [3. 管理 API: 测试用量 (`POST`)](#3-管理-api-测试用量-post)
- [用量数据结构说明](#用量数据结构说明)
- [开发与测试](#开发与测试)
- [许可证](#许可证)

---

## 功能特性

1. **标准 C ABI 兼容**：
   - 导出 `cliproxy_plugin_init`、`cliproxyPluginCall`、`cliproxyPluginFree`、`cliproxyPluginShutdown`。
   - 遵照 CLIProxyAPI 官方 JSON Envelope 规范（`ok`, `result`, `error`）。
2. **纯粹的管理监控能力 (`management_api`)**：
   - 注册插件自有的用量管理端点与浏览器嵌入式仪表盘资源页面。
   - 无多余的 OAuth 提供商注册，不污染 CLIProxyAPI 后台的 OAuth 授权列表。
3. **Session Token 灵活提取与支持**：
   - 支持在 `config.yaml` 配置或在配额页面上直接输入。
   - 支持纯 token 或完整 Cookie 字符串（自动提取 `__Secure-commandcode_prod_.session_token`）。
4. **精确用量与双滑动窗口限额解析**：
   - 上游接口：`GET https://api.commandcode.ai/internal/billing/credits`。
   - 请求优先走宿主提供的 `host.http.do` 回调（复用宿主代理、日志与鉴权管道），离线或未注入宿主时自动无缝降级至 Go 标准 `net/http`。
   - 全面解析 `credits`（月度基础额度、开源奖励额度、总可用额度）与 `windowLimits`（5小时短期滑动窗口、周度窗口限额，计算已用量、上限、剩余量、使用百分比及重置时间）。
5. **嵌入式纯单文件 QuotaCard 资源页**：
   - 页面挂载于 `/v0/resource/plugins/commandcode/quota`。
   - 零外部 CDN 依赖，纯内置 HTML + CSS + JS，深色/浅色模式自适应。
   - 具有进度条颜色变化、5小时/周限额卡片、秒级动态重置倒计时、同源 `localStorage` 鉴权与一键刷新。

---

## 系统架构

```
┌────────────────────────────────────────────────────────┐
│                      CLIProxyAPI                       │
│                                                        │
│                       ┌─────────────────────┐          │
│                       │  Management Center  │          │
│                       │   (/v0/management)  │          │
│                       └──────────┬──────────┘          │
│                                  │ C ABI               │
│                                  ▼                     │
│  ┌──────────────────────────────────────────────────┐  │
│  │       cliproxy-plugin-commandcode.dylib/.so      │  │
│  │                                                  │  │
│  │  • management.register / management.handle       │  │
│  │  • Usage Parser & Window Limits Formatter        │  │
│  │  • Embedded Single-file HTML/CSS/JS QuotaCard    │  │
│  └───────────────────────────┬──────────────────────┘  │
│                              │                         │
│                              │ host.http.do            │
│                              ▼                         │
│  ┌──────────────────────────────────────────────────┐  │
│  │          Host Transport / Proxy Pipeline         │  │
│  └───────────────────────────┬──────────────────────┘  │
└──────────────────────────────┼─────────────────────────┘
                               │ Upstream HTTPS
                               ▼
        https://api.commandcode.ai/internal/billing/credits
```

---

## 快速开始

### 构建插件

项目提供标准的 `Makefile`，可直接编译与操作系统相对应的 C 共享动态库：

```bash
# 自动编译出 commandcode.dylib (macOS) 或 commandcode.so (Linux)
make build

# 运行完整单元测试与竞态检测
make test

# 清理构建产物
make clean
```

### 安装与目录结构

将编译出的动态库放入 CLIProxyAPI 的插件目录中：

```bash
# macOS
mkdir -p plugins/darwin/arm64
cp commandcode.dylib plugins/darwin/arm64/commandcode.dylib

# Linux
mkdir -p plugins/linux/amd64
cp commandcode.so plugins/linux/amd64/commandcode.so
```

### 宿主配置 (`config.yaml`)

在 CLIProxyAPI 的 `config.yaml` 中启用插件并配置默认参数：

```yaml
plugins:
  enabled: true
  dir: "plugins"
  configs:
    commandcode:
      enabled: true
      priority: 1
      session_token: "YOUR_COMMANDCODE_SESSION_TOKEN" # 支持纯 token 或完整 Cookie 字符串
      api_base: "https://api.commandcode.ai" # 可选，默认为官方接口
```

---

## 管理端点与资源页

### 1. 浏览器资源页 (`QuotaCard`)

- **访问路径**：`GET http://<cpa-host>:8317/v0/resource/plugins/commandcode/quota`
- **菜单名**：`Command Code 配额`
- **说明**：
  - 资源请求本身无需经过管理认证，可在浏览器中直接打开或嵌入仪表盘。
  - 在同源模式下，页面 JavaScript 会自动读取 `localStorage` 中的管理密钥向 `/v0/management/plugins/commandcode/usage` 请求数据。
  - 若在独立或跨域测试环境下打开，页面提供内置的诊断面板，可手动输入 Management Key 或测试 Session Token。

### 2. 管理 API: 查询用量 (`GET`)

- **端点**：`GET /v0/management/plugins/commandcode/usage`
- **认证**：需要管理密钥 (`Authorization: Bearer <MANAGEMENT_KEY>` 或 `X-Management-Key: <MANAGEMENT_KEY>`)
- **可选查询参数**：
  - `session_token`: 临时覆盖查询的 token
  - `api_base`: 临时覆盖的上游基础 URL
- **响应示例**：

```json
{
  "ok": true,
  "credits": {
    "monthly_credits": 1000.0,
    "opensource_monthly_credits": 500.0,
    "total_credits": 1500.0,
    "details": {
      "monthlyCredits": 1000,
      "opensourceMonthlyCredits": 500
    }
  },
  "window_limits": {
    "five_hour": {
      "used": 12.5,
      "cap": 100.0,
      "remaining": 87.5,
      "percentage": 12.5,
      "exceeded": false,
      "reset_at": "2025-03-04T16:30:00Z",
      "reset_in_seconds": 7200
    },
    "weekly": {
      "used": 150.0,
      "cap": 1000.0,
      "remaining": 850.0,
      "percentage": 15.0,
      "exceeded": false,
      "reset_at": "2025-03-10T00:00:00Z",
      "reset_in_seconds": 475200
    }
  },
  "updated_at": "2025-03-04T14:30:00Z"
}
```

### 3. 管理 API: 测试用量 (`POST`)

- **端点**：`POST /v0/management/plugins/commandcode/usage`
- **请求体**：

```json
{
  "session_token": "YOUR_TEMPORARY_TOKEN",
  "api_base": "https://api.commandcode.ai"
}
```

---

## 用量数据结构说明

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `credits.monthly_credits` | `float64` | 当前账单周期的月度基础 Credits 额度 |
| `credits.opensource_monthly_credits` | `float64` | 开源项目贡献者获得的奖励额度 |
| `credits.total_credits` | `float64` | 可用 Credits 总计 (`monthly + opensource`) |
| `window_limits.five_hour.used` | `float64` | 5小时滑动窗口内已消耗的量 |
| `window_limits.five_hour.cap` | `float64` | 5小时滑动窗口上限 |
| `window_limits.five_hour.remaining` | `float64` | 5小时滑动窗口剩余可用量 |
| `window_limits.five_hour.percentage` | `float64` | 5小时窗口使用百分比（0-100%） |
| `window_limits.five_hour.exceeded` | `bool` | 是否已触发 5 小时限额熔断 |
| `window_limits.five_hour.reset_at` | `string` | 5小时窗口重置时间的 RFC3339 字符串 |
| `window_limits.five_hour.reset_in_seconds`| `int64` | 距离 5 小时窗口重置的剩余秒数 |
| `window_limits.weekly.*` | - | 每周限额对应指标（结构同 5 小时窗口） |

---

## 开发与测试

```bash
# 运行单元测试
go test -v ./...

# 运行代码规范检查
go vet ./...

# 运行竞态检查测试
go test -race -v ./...
```

---

## 许可证

本项目基于 [MIT License](LICENSE) 开源。
