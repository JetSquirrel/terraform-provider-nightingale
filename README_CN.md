# Nightingale Terraform Provider

[![Go Report Card](https://goreportcard.com/badge/github.com/JetSquirrel/terraform-provider-nightingale)](https://goreportcard.com/report/github.com/JetSquirrel/terraform-provider-nightingale)
[![License](https://img.shields.io/badge/License-MPL%202.0-blue.svg)](LICENSE)

[English](README.md) | [中文](README_CN.md)

用于管理 [夜莺监控](https://github.com/ccfos/nightingale) (Nightingale/n9e) 资源的 Terraform Provider。

## 功能特性

- 管理告警规则（支持 PromQL 查询）
- 配置通知规则（多通道支持）
- 设置告警订阅
- 完整的 CRUD 支持及资源导入
- 兼容夜莺 v9.x API

## 环境要求

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21（从源码构建时需要）
- 夜莺 v8.0+（已在 v9.x 测试通过）

## 安装

### 从 Terraform Registry 安装

```terraform
terraform {
  required_providers {
    nightingale = {
      source  = "JetSquirrel/nightingale"
      version = "~> 0.1"
    }
  }
}
```

### 从源码安装

```shell
git clone https://github.com/JetSquirrel/terraform-provider-nightingale.git
cd terraform-provider-nightingale
go install
```

## 快速开始

### 1. 配置 Provider

```terraform
provider "nightingale" {
  endpoint = "https://n9e.example.com"
  token    = var.nightingale_token
}
```

或使用环境变量：

```shell
export NIGHTINGALE_ENDPOINT="https://n9e.example.com"
export NIGHTINGALE_TOKEN="your-api-token"
```

### 2. 获取 API Token

1. 登录夜莺 Web 界面
2. 点击右上角头像，进入 **个人中心** > **Token 管理**
3. 点击 **创建 Token**
4. 复制生成的 Token

> **注意：** 确保夜莺配置文件 `config.toml` 中设置了 `[HTTP.TokenAuth] Enable = true`

### 3. 创建资源

```terraform
# CPU 使用率告警规则
resource "nightingale_alert_rule" "high_cpu" {
  busi_group_id   = 1
  name            = "CPU 使用率过高"
  datasource_type = "prometheus"
  datasource_ids  = [1]
  severity        = 2

  queries = [{
    ref              = "A"
    promql           = "100 - avg by (ident) (rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100 > 80"
    duration_seconds = 300
  }]

  annotations = {
    summary     = "{{ $labels.ident }} CPU 使用率过高"
    description = "CPU 使用率超过 80% 持续 5 分钟"
  }

  append_tags = ["managed_by=terraform"]
}
```

## 支持的资源

| 资源 | 说明 | 导入格式 |
|------|------|----------|
| [`nightingale_alert_rule`](docs/resources/alert_rule.md) | 告警规则 | `业务组ID:规则ID` |
| [`nightingale_notify_rule`](docs/resources/notify_rule.md) | 通知规则 | `规则ID` |
| [`nightingale_alert_subscribe`](docs/resources/alert_subscribe.md) | 告警订阅 | `业务组ID:订阅ID` |

## Provider 配置

| 属性 | 环境变量 | 必填 | 默认值 | 说明 |
|------|----------|------|--------|------|
| `endpoint` | `NIGHTINGALE_ENDPOINT` | 是 | - | 夜莺 API 地址 |
| `token` | `NIGHTINGALE_TOKEN` | 是 | - | API Token |
| `timeout_seconds` | `NIGHTINGALE_TIMEOUT_SECONDS` | 否 | 30 | HTTP 超时时间（秒） |
| `insecure_skip_tls_verify` | `NIGHTINGALE_INSECURE_SKIP_TLS_VERIFY` | 否 | false | 跳过 TLS 证书验证 |

## 示例

查看 [examples](examples/) 目录获取完整配置示例：

- [Provider 配置](examples/provider/)
- [完整示例](examples/complete/) - 多资源协同工作
- [单独资源示例](examples/resources/)

## 导入现有资源

```shell
# 告警规则
terraform import nightingale_alert_rule.example 1:123

# 通知规则
terraform import nightingale_notify_rule.example 456

# 告警订阅
terraform import nightingale_alert_subscribe.example 1:789
```

## 开发

### 构建

```shell
make build
```

### 测试

```shell
# 单元测试
make test

# 验收测试（需要运行中的夜莺实例）
export TF_ACC=1
export NIGHTINGALE_ENDPOINT="http://localhost:17000"
export NIGHTINGALE_TOKEN="your-token"
make testacc
```

### 生成文档

```shell
make generate
```

## 贡献

欢迎提交 Pull Request！

## 许可证

[MPL-2.0](LICENSE)

## 相关链接

- [夜莺项目](https://github.com/ccfos/nightingale)
- [夜莺文档](https://flashcat.cloud/docs/)
- [Terraform Registry](https://registry.terraform.io/providers/JetSquirrel/nightingale)
