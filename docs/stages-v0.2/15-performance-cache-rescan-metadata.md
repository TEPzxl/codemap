# Stage 15 — Performance, Cache, Rescan, and Metadata

## 目标

提升本地服务对中型 Go 项目的可用性：展示项目元信息、分析耗时、缓存状态，并支持用户手动重新扫描。

v0.1 已在 `serve` 启动时加载并缓存项目。本阶段要把缓存状态和刷新能力显式化。

## 需要完成

### Project metadata

新增 endpoint：

```text
GET /api/meta
```

建议返回：

```json
{
  "root": "/path/to/project",
  "module": "github.com/example/project",
  "packages": 12,
  "symbols": 128,
  "calls": 340,
  "warnings": 0,
  "analyzed_at": "2026-...",
  "analysis_duration_ms": 123,
  "version": "v0.2.0"
}
```

### Rescan

新增 endpoint：

```text
POST /api/rescan
```

行为：

1. 重新加载当前项目。
2. 重建 symbols / calls / graph index。
3. 保证 rescan 过程中 API 不返回部分损坏状态。
4. rescan 失败时保留旧 index，返回 JSON error。

### CLI

可选：

```bash
codemap serve ./project --watch
```

本阶段不强制做 watch。手动 rescan 优先。

### UI

1. 显示项目元信息。
2. 显示分析时间和 symbol/call 数量。
3. 提供 Rescan 按钮。
4. Rescan 时显示 loading。
5. Rescan 成功后刷新 symbols 和 graph。

## 并发要求

如果 server 使用共享 index：

- 读取请求可以并发。
- rescan 时不能造成 data race。
- 建议使用 `sync.RWMutex` 或不可变 snapshot 替换。
- 必须能通过 `go test -race ./...`，如果环境允许。

## 测试要求

至少覆盖：

1. `/api/meta` 返回正确 counts。
2. `/api/rescan` 成功后 metadata 更新。
3. rescan 失败时返回 JSON error。
4. rescan 失败不清空旧 index。
5. 并发请求不 panic。
6. UI build 通过。

## 验收命令

```bash
make check
go test -race ./internal/server/... ./internal/analyzer/... ./internal/graph/...
make web-build
make build
./bin/codemap serve ./examples/layered-service --port 8080
curl -s http://localhost:8080/api/meta | python -m json.tool
curl -s -X POST http://localhost:8080/api/rescan | python -m json.tool
```

人工验收：

1. Web UI 显示 packages / symbols / calls。
2. 点击 Rescan 后 graph 仍然可加载。
3. Rescan 期间 UI 不白屏。

## 成功标准

- API 请求不重复扫描项目。
- 用户能看到项目分析状态。
- 用户能手动刷新项目索引。
- rescan 失败不破坏已有可用 graph。

## 常见失败模式

- 每次 `/api/graph` 都重新扫描项目。
- rescan 过程中 index 被部分替换。
- rescan 失败后 symbols 为空。
- metadata counts 与真实 index 不一致。
- 使用全局变量导致测试互相污染。
