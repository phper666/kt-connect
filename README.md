KT-Connect
===========

![Go](https://github.com/alibaba/kt-connect/workflows/Go/badge.svg)
[![Build Status](https://travis-ci.org/alibaba/kt-connect.svg?branch=master)](https://travis-ci.org/alibaba/kt-connect)
[![Go Report Card](https://goreportcard.com/badge/github.com/alibaba/kt-connect)](https://goreportcard.com/report/github.com/alibaba/kt-connect)
[![Test Coverage](https://api.codeclimate.com/v1/badges/eb13b3946784bd7c67cc/test_coverage)](https://codeclimate.com/github/alibaba/kt-connect/test_coverage)
[![Maintainability](https://api.codeclimate.com/v1/badges/eb13b3946784bd7c67cc/maintainability)](https://codeclimate.com/github/alibaba/kt-connect/maintainability)
[![Release](https://img.shields.io/github/release/alibaba/kt-connect.svg?style=flat-square)](https://img.shields.io/github/release/alibaba/kt-connect.svg?style=flat-square)
![License](https://img.shields.io/github/license/alibaba/kt-connect.svg)

English | [简体中文](./README_CN.md)

---

> **Fork 说明**：本 fork 在官方 v0.3.7 基础上增加[停止信号文件支持](https://github.com/phper666/kt-connect/commit/68c7d02)（供 IDEA 调试插件 ms-localdebug 优雅终止代理），并提供独立的三端打包发布。

## 下载（Fork Release）

- 全部版本：[Releases](https://github.com/phper666/kt-connect/releases)
- 最新 `v0.3.7-phper666.1`（zip：Linux/macOS/Windows × x86_64/i386/arm_64）：

| 平台 | 下载链接 |
|:--|:--|
| macOS (Intel) | `https://github.com/phper666/kt-connect/releases/download/v0.3.7-phper666.1/ktctl_0.3.7-phper666.1_MacOS_x86_64.zip` |
| macOS (Apple Silicon) | `https://github.com/phper666/kt-connect/releases/download/v0.3.7-phper666.1/ktctl_0.3.7-phper666.1_MacOS_arm_64.zip` |
| Linux | `https://github.com/phper666/kt-connect/releases/download/v0.3.7-phper666.1/ktctl_0.3.7-phper666.1_Linux_x86_64.zip` |
| Windows | `https://github.com/phper666/kt-connect/releases/download/v0.3.7-phper666.1/ktctl_0.3.7-phper666.1_Windows_x86_64.zip` |

> IDEA 调试插件 ms-localdebug 的「工具安装」下载地址可直接使用上表链接（安装器按 zip 解压）。

## 打包（维护者）

Actions → `release` → Run workflow：

- **tag 留空**：仅构建验证（snapshot 模式，不发布）
- **填 tag**（如 `v0.3.7-phper666.1`）：正式发布 Release 并上传全部产物

> 注意：二进制内嵌 `main.version` 固定为 `0.3.7`——ktctl 用它拼接 shadow/router 镜像 tag（`v{version}`），需对齐官方镜像仓库，勿改为 release tag。

---


KtConnect ("Kt" is short for "Kubernetes Toolkit") is a utility tool to help you work with Kubernetes dev environment more efficiently.

![Arch](./docs/media/arch.png)

## ✅ Features

* `Connect`: Directly Access a remote Kubernetes cluster. KtConnect use ssh-vpn or socks-proxy to access remote Kubernetes cluster networks.
* `Exchange`: Developer can exchange the workload to redirect the requests to a local app.
* `Mesh`: You can create a mesh version service in local host, and redirect specified workload requests to your local.
* `Preview`: Expose a local running app to Kubernetes cluster as a common service, all requests to that service are redirect to local app.

## 🚀 QuickStart

You can download and install the `ktctl` from [Downloads And Install](docs/en-us/guide/downloads.md)

Read the [Quick Start Guide](docs/en-us/guide/quickstart.md) for more about this tool.

## 💡 Ask For Help

Please feel free to raise an [issue](https://github.com/alibaba/kt-connect/issues) if anything sucks, or go ahead to contact us with DingTalk（钉钉）:

<img src="https://img.alicdn.com/imgextra/i4/O1CN01sTW3D61NzAFgUCNqz_!!6000000001640-0-tps-573-657.jpg" width="50%"></img>

