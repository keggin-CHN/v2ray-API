# GoAI Relay: A Go-Based OpenAI-Compatible Intelligent API Gateway

[English](README.md) | [中文](README_zh.md)

<p align="center">
  <img src="./assets/banner.png" alt="API-V2Ray Banner" width="100%">
</p>

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

> **A high-performance OpenAI-compatible API relay platform built with Go, supporting multiple V2Ray-core subscription formats with unified access, subscription parsing, node management, and reliable forwarding.**

**This project is an OpenAI-compatible API relay platform built with Go, designed for high-concurrency, scalable, and multi-node access scenarios. It supports multiple V2Ray-core subscription formats and provides unified subscription parsing, management, and forwarding capabilities.**

**💡 Core Innovation: 1:1 Proxy-to-API Binding**
The platform features extremely flexible egress control, supporting **strict one-to-one mapping between a specific proxy node and an API endpoint**. Through precise underlying routing, it ensures that different API requests securely and stably reach their upstream destinations via designated V2Ray nodes (e.g., VLESS, Hysteria2), offering developers and service providers a stable, efficient, and flexible AI API gateway solution.

## 🚀 Core Features

*   **Intelligent Routing & Failover**
    Dynamic endpoint evaluation with automatic circuit breaking and multi-tier failover strategies (customizable cooldown mechanisms, fallback upstreaming).
*   **Decentralized Node Management**
    Native integration with Xray-core for transparently managing proxy topologies. Supports zero-downtime hot-reloading of configurations and real-time subscription parsing.
*   **Omni-Protocol Support**
    Seamlessly handles multiple transport layers: VLESS (xtls-rprx-vision/reality), Hysteria2, VMess, Trojan, and standard Shadowsocks/SOCKS5.
*   **LLM API Optimization**
    Built specifically for OpenAI and compatible API schemas. Includes request multiplexing, intelligent endpoint binding, and transparent payload streaming.
*   **Zero-Overhead Bootstrap**
    Flat runtime generation logic ensuring minimal GC pauses, minimal memory footprint, and maximum parallel throughput.

## 🏗️ Architecture Design

At its core, `API-V2Ray` implements a multi-stage deterministic pipeline:
1.  **Ingress:** Lightweight, high-concurrency HTTP layer handling incoming client RESTful calls.
2.  **Router/Binder:** Dynamic routing based on upstream mappings, failover states, and priority queues.
3.  **Proxy Runtime:** On-the-fly generation of Xray outbounds strictly bound to specific node configurations.
4.  **Egress:** Establishing multiplexed secure tunnels via underlying Xray processes for the final mile to the upstream AI endpoint.

## 🛠️ Quick Start

### Prerequisites
- Go 1.22+
- Pre-compiled Xray-core binary (placed in `./.bin/xray` by default)

### Build & Run
```bash
# Clone the repository
git clone https://github.com/your-org/api-v2ray.git
cd api-v2ray

# Build the binary
go build -o bin/api-v2ray ./cmd/api-v2ray

# Run with local configuration
./bin/api-v2ray -config ./configs/config.local.json
```

## 📜 Configuration Mapping

`API-V2Ray` employs a robust JSON/YAML schema for mapping upstreams, proxy nodes, and bindings. Key components include:
- `upstreams`: Defines target AI endpoints (e.g., GPT-4 endpoints) and authentication keys.
- `proxy_nodes`: Defines raw proxy URIs (VLESS, Hysteria2, etc.).
- `bindings`: Glues upstreams to specific proxy nodes for deterministic egress.

See `configs/config.example.json` for a detailed reference architecture.

---
*Crafted for robust distributed networking.*
