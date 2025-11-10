# NFPCF 使用说明

## 概述

NFPCF (NF Profile Cache Function) 是一个用于缓存 NF Profile 的代理服务，位于 NF 实例和 NRF 之间。它通过缓存 NF Discovery 的结果来减少对 NRF 的请求压力，提高发现性能。

## 核心功能

### 1. NF Discovery 缓存
- 收到 NF Discovery 请求时，首先从缓存查找
- 缓存命中则直接返回，缓存未命中则向 NRF 查询
- 查询结果会被缓存，TTL 默认 5 分钟

### 2. NF Management 透传
- NF 注册/注销/更新请求透传到后端 NRF
- 注销和更新操作会使缓存失效

### 3. 自动缓存清理
- 定期清理过期的缓存条目
- 避免内存泄漏

## 快速开始

### 构建

```bash
cd /home/mazhuoran/src/free5gc-compose/base/free5gc/NFs/nfpcf
make build
```

### 配置

编辑 `config/nfpcfcfg.yaml`:

```yaml
info:
  version: 1.0.0
  description: NF Profile Cache Function

server:
  bindAddr: 0.0.0.0:8000  # NFPCF 监听地址

nrf:
  url: http://nrf:8000  # 后端 NRF 地址

cache:
  ttl: 300000000000  # 缓存 TTL (纳秒，这里是 5 分钟)

logger:
  level: info
```

### 运行

#### 方式 1: 使用 Makefile

```bash
make run
```

#### 方式 2: 直接运行

```bash
./bin/nfpcf -c ./config/nfpcfcfg.yaml
```

#### 方式 3: 自定义配置

```bash
./bin/nfpcf --config /path/to/your/config.yaml
```

## 与 Free5GC 集成

### 1. 修改 NF 配置

将各 NF 的 `nrfUri` 指向 NFPCF 而不是直接指向 NRF。

例如，修改 AMF 配置 (`config/amfcfg.yaml`):

```yaml
configuration:
  nrfUri: http://nfpcf:8000  # 原来是 http://nrf:8000
```

类似地修改其他 NF 配置 (SMF, UPF, AUSF, UDM, PCF, etc.)

### 2. Docker Compose 配置

在 `docker-compose.yaml` 中添加 NFPCF 服务:

```yaml
services:
  nfpcf:
    container_name: nfpcf
    build:
      context: ./base/free5gc/NFs/nfpcf
      dockerfile: Dockerfile
    environment:
      - NRF_URL=http://nrf:8000
    networks:
      - free5gc
    ports:
      - "8000:8000"
    depends_on:
      - nrf

  # 其他 NF 配置...
  amf:
    depends_on:
      - nfpcf  # 依赖 NFPCF
    # ...
```

### 3. 启动顺序

1. 启动 NRF
2. 启动 NFPCF
3. 启动其他 NF

```bash
docker-compose up -d nrf
docker-compose up -d nfpcf
docker-compose up -d
```

## API 端点

### NF Management API

```
PUT    /nnrf-nfm/v1/nf-instances/:nfInstanceID     # 注册 NF
GET    /nnrf-nfm/v1/nf-instances/:nfInstanceID     # 获取 NF Profile
DELETE /nnrf-nfm/v1/nf-instances/:nfInstanceID     # 注销 NF
PATCH  /nnrf-nfm/v1/nf-instances/:nfInstanceID     # 更新 NF
```

### NF Discovery API

```
GET    /nnrf-disc/v1/nf-instances?target-nf-type=AMF&requester-nf-type=SMF  # 发现 NF
```

## 测试

### 1. 测试 NF Discovery (使用 curl)

```bash
# 发现 AMF
curl -X GET "http://localhost:8000/nnrf-disc/v1/nf-instances?target-nf-type=AMF&requester-nf-type=SMF"

# 第一次请求会向 NRF 查询，第二次请求会从缓存返回
curl -X GET "http://localhost:8000/nnrf-disc/v1/nf-instances?target-nf-type=AMF&requester-nf-type=SMF"
```

### 2. 测试 NF Registration

```bash
curl -X PUT http://localhost:8000/nnrf-nfm/v1/nf-instances/test-nf-001 \
  -H "Content-Type: application/json" \
  -d '{
    "nfInstanceId": "test-nf-001",
    "nfType": "AMF",
    "nfStatus": "REGISTERED",
    "ipv4Addresses": ["10.0.0.1"]
  }'
```

## 性能优化建议

### 1. 调整缓存 TTL

根据网络环境调整 TTL:

- 测试环境: 1-5 分钟
- 生产环境: 5-15 分钟
- 高频更新环境: 30 秒 - 1 分钟

```yaml
cache:
  ttl: 60000000000  # 1 分钟 (纳秒)
```

### 2. 多实例部署

对于高可用部署，可以运行多个 NFPCF 实例，使用负载均衡:

```yaml
services:
  nfpcf1:
    # ...
  nfpcf2:
    # ...

  nginx:
    image: nginx
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    ports:
      - "8000:8000"
```

## 监控

### 日志

NFPCF 日志输出到标准输出，可以使用 Docker logs 查看:

```bash
docker logs -f nfpcf
```

### 指标

未来可以添加 Prometheus 指标:
- 缓存命中率
- 请求延迟
- 缓存大小
- NRF 请求数

## 故障排查

### 1. NFPCF 无法连接 NRF

检查配置中的 NRF URL 是否正确:

```yaml
nrf:
  url: http://nrf:8000
```

### 2. 缓存未生效

- 检查 TTL 配置
- 确认查询参数一致
- 查看日志确认缓存操作

### 3. NF 无法注册

- 确认 NFPCF 可以访问 NRF
- 检查网络连接
- 查看 NFPCF 和 NRF 日志

## 架构图

```
┌────────────────┐
│  NF Instances  │
│ (AMF/SMF/UPF)  │
└────────┬───────┘
         │
         │ Discovery/Registration
         │
         ▼
┌─────────────────────────┐
│       NFPCF             │
│                         │
│  ┌──────────────────┐   │
│  │  Cache Layer     │   │
│  │  - TTL Based     │   │
│  │  - Type Indexed  │   │
│  └──────────────────┘   │
│                         │
│  [Cache Hit] ───────►   │
│       │                 │
│  [Cache Miss]           │
│       │                 │
│       ▼                 │
│  Forward to NRF         │
└────────┬────────────────┘
         │
         │ Fallback
         │
         ▼
┌────────────────────┐
│       NRF          │
│   (Backend)        │
└────────────────────┘
```

## 与 NRF 的区别

| 特性 | NRF | NFPCF |
|------|-----|-------|
| 数据持久化 | MongoDB | 内存缓存 |
| 权威数据源 | 是 | 否 |
| 查询延迟 | 较高 | 很低 (缓存命中时) |
| 数据一致性 | 强一致 | 最终一致 (TTL) |
| 适用场景 | 注册/管理 | 频繁发现 |

## 限制

1. **只缓存 Discovery 结果**: NF Management 操作直接透传，不缓存
2. **内存存储**: 缓存只在内存中，重启会丢失
3. **单机部署**: 多实例之间缓存不共享
4. **最终一致性**: 缓存可能与 NRF 有延迟

## 未来改进

1. **Redis 支持**: 使用 Redis 作为共享缓存
2. **Metrics**: 添加 Prometheus 指标
3. **管理 API**: 添加缓存管理接口
4. **事件通知**: 支持 NRF 事件通知以主动失效缓存
5. **更智能的匹配**: 改进查询参数匹配逻辑

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

与 Free5GC 项目相同
