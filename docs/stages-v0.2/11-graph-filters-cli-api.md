# Stage 11 — Graph Filters in CLI and API

## 目标

统一 CLI 和 HTTP API 的 graph filtering 能力，使用户可以明确控制图中是否显示 external、unresolved、interface、interface candidates 等边和节点。

Stage 11 主要做后端过滤接口，不做复杂 UI。

## 背景

v0.1 已有部分 CLI flag：

```text
--show-external
--show-unresolved
--show-interface
```

v0.2 应把这些能力统一到：

- CLI
- HTTP API
- Graph builder options
- tests

## 需要支持的过滤选项

建议定义统一结构：

```go
type GraphOptions struct {
    Entry           string
    Depth           int
    ShowExternal    bool
    ShowUnresolved  bool
    ShowInterface   bool
    ExpandInterface bool
    PackagePrefixes []string
    NodeLimit       int
}
```

具体字段名可以按当前代码风格调整。

## CLI 行为

支持：

```bash
codemap graph ./examples/layered-service --entry main.main --depth 5
codemap graph ./examples/layered-service --entry main.main --depth 5 --show-external
codemap graph ./examples/interface-call --entry main.main --depth 5 --show-interface
codemap graph ./examples/interface-call --entry main.main --depth 5 --expand-interface
codemap graph ./examples/layered-service --entry main.main --depth 5 --package github.com/tepzxl/codemap/examples/layered-service/internal/service
codemap graph ./examples/layered-service --entry main.main --depth 5 --node-limit 100
```

## API 行为

支持 query 参数：

```text
GET /api/graph?entry=main.main&depth=5
GET /api/graph?entry=main.main&depth=5&show_external=true
GET /api/graph?entry=main.main&depth=5&show_unresolved=true
GET /api/graph?entry=main.main&depth=5&show_interface=true
GET /api/graph?entry=main.main&depth=5&expand_interface=true
GET /api/graph?entry=main.main&depth=5&package=<prefix>
GET /api/graph?entry=main.main&depth=5&node_limit=100
```

## 默认行为

默认保持 v0.1：

- `show_external=false`
- `show_unresolved=false`
- `show_interface=false` 或沿用 v0.1 当前默认
- `expand_interface=false`
- `node_limit` 使用合理默认值，例如 500 或 1000

## 错误处理

必须返回 JSON error：

```json
{"error":"invalid depth"}
```

覆盖：

- 非法 boolean。
- 非法 depth。
- 非法 node limit。
- unknown entry。
- package filter 过滤后 graph 为空。

## 测试要求

至少覆盖：

1. 默认隐藏 external。
2. `--show-external` 显示 external。
3. `show_external=true` API 生效。
4. `show_unresolved=true` API 生效。
5. `show_interface=true` API 生效。
6. `expand_interface=true` API 生效。
7. package prefix filter 生效。
8. node limit 触发 warning 或明确错误。
9. CLI 和 API 对同一 options 输出一致。

## 验收命令

```bash
make check
go run ./cmd/codemap graph ./examples/layered-service --entry main.main --depth 5
go run ./cmd/codemap graph ./examples/layered-service --entry main.main --depth 5 --show-external
go run ./cmd/codemap graph ./examples/interface-call --entry main.main --depth 5 --expand-interface
./bin/codemap serve ./examples/interface-call --port 8080
curl -s "http://localhost:8080/api/graph?entry=main.main&depth=5&expand_interface=true" | python -m json.tool
```

## 成功标准

- CLI 和 API 使用同一套 GraphOptions。
- 默认行为不破坏 v0.1。
- 不同过滤开关有明确测试覆盖。
- API 错误为 JSON。
- warnings 能说明过滤导致的信息丢失。

## 常见失败模式

- CLI 和 API 过滤逻辑各写一套，行为不一致。
- bool query 参数只接受 `true`，不处理非法值。
- 过滤 nodes 后留下 dangling edges。
- package filter 只过滤 node，不过滤 edge。
- node limit 截断后没有 warning。
