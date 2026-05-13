# GoAI Relay：基于 Go 的 OpenAI 兼容智能中转网关

[English](README.md) | [中文](README_zh.md)

<p align="center">
  <img src="./assets/banner.png" alt="API-V2Ray Banner" width="100%">
</p>

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

> **基于 Go 构建的高性能 OpenAI 兼容 API 中转平台，支持多种 V2Ray 内核订阅格式，提供统一接入、订阅解析、节点管理与稳定转发能力。**

**本项目是一个使用 Go 语言构建的 OpenAI 兼容 API 中转平台，面向高并发、可扩展和多节点接入场景设计。平台兼容多种 V2Ray 内核订阅格式，能够对订阅内容进行统一解析、管理与转发。**

**💡 核心独创特性：一节点一接口 (1:1 Proxy-to-API Binding)**
平台具备极其灵活的出口控制能力，支持**“一个代理节点严格适配一个 API 链接”**。通过底层的精准路由绑定，确保不同的 API 请求能够通过指定的 V2Ray 节点（如 VLESS, Hysteria2 等）稳定、安全地直达上游，为开发者和服务提供者提供稳定、高效、灵活的 AI API 接入层解决方案。

## 🚀 核心特性

*   **智能路由与熔断机制 (Intelligent Routing & Failover)**
    支持动态节点评估与自动熔断，内置多级故障转移策略（自定义冷却机制、上游备用回退）。
*   **去中心化节点管理 (Decentralized Node Management)**
    原生集成 Xray-core 底层，透明管理复杂的代理拓扑。支持零宕机配置热重载以及实时订阅解析。
*   **全协议支持 (Omni-Protocol Support)**
    无缝接管多种传输层协议，涵盖：VLESS (xtls-rprx-vision/reality), Hysteria2, VMess, Trojan，以及标准的 Shadowsocks/SOCKS5。
*   **LLM API 专项优化 (LLM API Optimization)**
    专为 OpenAI 及兼容格式的 API schema 打造。内置请求多路复用（Multiplexing）、智能节点绑定（Endpoint Binding）以及无缝的数据流传输（Payload Streaming）。
*   **零开销启动 (Zero-Overhead Bootstrap)**
    扁平化的运行时生成逻辑，保证极低的 GC 停顿与内存占用，最大化并发吞吐量。

## 🏗️ 架构设计

在核心实现上，`API-V2Ray` 采用了确定性的多阶段处理流水线：
1.  **Ingress (入口网络):** 轻量级、高并发的 HTTP 层，负责承接客户端的 RESTful 呼叫。
2.  **Router/Binder (路由与绑定器):** 基于上游映射、熔断状态机和优先级队列执行动态流量调度。
3.  **Proxy Runtime (代理运行时):** 动态生成与特定节点配置严格绑定的 Xray Outbounds（出站协议）。
4.  **Egress (出口网络):** 借助底层的 Xray 进程建立多路复用的加密安全隧道，完成到达上游 AI 端点的“最后一公里”。

## 🧩 服务分层（当前实现）

当前 HTTP 服务已按职责拆分到独立文件，提升可维护性与可读性：

- 认证相关处理器：[`internal/server/server_handlers_auth.go`](internal/server/server_handlers_auth.go)
- 管理/配置/引导处理器：[`internal/server/server_handlers_admin.go`](internal/server/server_handlers_admin.go)
- 导入处理器：[`internal/server/server_handlers_imports.go`](internal/server/server_handlers_imports.go)
- 可观测性处理器：[`internal/server/metrics.go`](internal/server/metrics.go)
- 通用中间件：[`internal/server/middleware.go`](internal/server/middleware.go)

通过拆分，显著降低了 [`Server`](internal/server/server.go:28) 的体量与耦合度，接口职责也更清晰。

## 🛠️ 快速开始

### 环境依赖
- Go 1.22+
- 预编译的 Xray-core 二进制文件（默认放置于 `./.bin/xray`）

### 构建与运行
```bash
# 克隆仓库
git clone https://github.com/your-org/api-v2ray.git
cd api-v2ray

# 构建二进制文件
go build -o bin/api-v2ray ./cmd/api-v2ray

# 使用本地配置运行
./bin/api-v2ray -config ./configs/config.local.json
```

## 🔐 安全部署建议

- 将 [`server.admin_token`](configs/config.example.json) 设置为高强度随机串，避免默认值。
- 仅在受信网络暴露管理接口，建议前置反向代理并启用 HTTPS。
- 定期轮换上游 [`api_key`](configs/config.example.json) 并最小化权限范围。
- 为运行目录（如 [`runtime/`](runtime/)）配置最小权限账户，避免使用高权限用户运行。

## 📈 可观测性

### 内置观测接口

- `GET /healthz`：基础存活探针。
- `GET /api/health/routes`：路由健康快照（需认证）。
- `GET /api/metrics/upstream`：上游请求计数指标，来自 [`StatsSnapshot()`](internal/upstream/client.go:160)（需认证）。
- `GET /api/metrics/runtime`：返回进程状态文件路径、已跟踪进程列表、路由健康与上游计数（需认证）。
- `GET /api/diagnostics/exit-ip?url=...&fallback_url=...`：出口 IP 诊断，支持主探针失败后回退探针（需认证）。

### 访问日志与请求追踪

全局中间件链已接入 [`Handler()`](internal/server/server.go:52)：
- panic 恢复（含堆栈日志）、
- `X-Request-ID` 透传/自动生成、
- 统一访问日志（method/path/status/bytes/latency）。

## 🧰 常见排障

- `invalid json body`：请求体包含未知字段、多个 JSON 对象，或格式不合法。
- `dry-run bootstrap failed`：配置结构可解析但路由/绑定/节点关系在预检阶段失败。
- `runtime launch failed`：Xray 进程启动异常，请检查 [`runtime/xray/*.json`](runtime/) 与二进制路径。
- `all upstream candidates failed`：上游不可达、认证失败或代理节点异常，建议查看服务日志中的 `upstream_request_*` 记录。

## 🧪 运维手册（精简）

- 配置变更建议优先使用 `POST /api/config/apply`，执行 dry-run 引导与运行时启动校验。
- 若保存阶段失败且运行时已启动，系统会自动尝试回滚（旧配置 + 运行时重启）。
- 运行时进程状态持久化在 `runtime/xray/processes.json`（或自定义 runtime 目录），可通过 `GET /api/metrics/runtime` 检查。
- 上游抖动排查时，建议联合分析路由健康（`/api/health/routes`）、上游计数（`/api/metrics/upstream`）与 `upstream_request_*` 日志。

## 配置映射

`API-V2Ray` 采用高度健壮的 JSON/YAML schema 来映射上游（upstreams）、代理节点（proxy nodes）以及绑定关系（bindings）。核心组件包括：
- `upstreams`: 定义目标 AI 端点（例如 GPT-4 接口）以及鉴权密钥。
- `proxy_nodes`: 定义原始的代理链接（VLESS, Hysteria2 等）。
- `bindings`: 将特定上游与代理节点硬性绑定，以实现确定性的出口路由。

详尽的架构参考请查阅 `configs/config.example.json`。

---
*为构建坚不可摧的分布式网络而生。*
