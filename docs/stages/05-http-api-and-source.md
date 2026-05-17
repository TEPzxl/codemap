# Stage 05 — HTTP API and Source Snippets

## 目标

实现本地 HTTP API，为前端提供 symbols、graph、source snippet 和 warnings。Go 后端仍然是唯一负责扫描本地仓库和读取源码的部分。

---

## 本阶段需要完成的内容

### 1. 实现 serve 命令

```bash
codemap serve <path> --port 8080
```

职责：

- 加载项目。
- 构建 analyzer index。
- 启动 HTTP server。
- 提供 JSON API。
- 开发阶段可只提供 API；发布阶段再嵌入前端静态资源。

### 2. API 设计

建议接口：

```text
GET /api/health
GET /api/symbols
GET /api/graph?entry=<symbol>&depth=5
GET /api/source?node_id=<symbol>
GET /api/warnings
```

### 3. `/api/health`

返回：

```json
{
  "status": "ok"
}
```

### 4. `/api/symbols`

返回所有可作为入口的 symbol：

```json
{
  "symbols": [
    {
      "id": "...",
      "label": "UserService.CreateUser",
      "kind": "method",
      "package": "...",
      "file": "internal/service/user.go",
      "start_line": 10,
      "end_line": 32
    }
  ]
}
```

### 5. `/api/graph`

返回 Graph JSON：

```text
GET /api/graph?entry=github.com/acme/app.main&depth=5
```

要求：

- entry 必填。
- depth 可选，默认 5。
- invalid entry 返回 400。
- 分析错误返回结构化 error。

### 6. `/api/source`

返回当前 node 对应源码：

```json
{
  "node_id": "...",
  "file": "internal/service/user.go",
  "start_line": 10,
  "end_line": 32,
  "source": "func ...",
  "language": "go"
}
```

源码读取要求：

- 只能读取被扫描项目根目录内的文件。
- 防止 path traversal。
- 不返回整个仓库文件。
- 根据 node 的 `file/start_line/end_line` 截取片段。

### 7. `/api/warnings`

返回 package loading warning、unresolved call summary 等。

---

## 验收方式

### 启动服务

```bash
go run ./cmd/codemap serve ./examples/layered-service --port 8080
```

通过标准：

- 服务启动。
- 不 panic。
- 控制台显示本地 URL。
- API 可请求。

### health

```bash
curl -s http://localhost:8080/api/health
```

通过标准：

```json
{"status":"ok"}
```

### symbols

```bash
curl -s http://localhost:8080/api/symbols
```

通过标准：

- 返回合法 JSON。
- 包含 `symbols`。
- 至少包含 `main.main` 或 layered fixture 中的入口 symbol。

### graph

```bash
curl -s "http://localhost:8080/api/graph?entry=main.main&depth=5"
```

通过标准：

- 返回 Graph JSON。
- 包含 nodes 和 edges。
- depth 生效。
- 无破坏性日志混入响应。

### source

```bash
curl -s "http://localhost:8080/api/source?node_id=<known-node-id>"
```

通过标准：

- 返回源码片段。
- 行号范围与 node 一致。
- source 不为空。
- 文件路径在项目根目录内。

### 自动化测试

```bash
go test ./internal/server/... ./internal/source/...
```

建议断言：

- `/api/health` 正常。
- `/api/symbols` 返回 expected content。
- `/api/source` 拒绝未知 node。
- `/api/source` 不允许 path traversal。
- invalid graph entry 返回 400。

---

## 成功完成标准

本阶段完成时，应满足：

1. `codemap serve <path>` 可用。
2. API 能返回 symbols。
3. API 能返回 graph。
4. API 能返回 source snippet。
5. API 能返回 warnings。
6. source 读取具备基本路径安全保护。
7. API 错误返回结构化 JSON。
8. `go test ./internal/server/... ./internal/source/...` 通过。

---

## 常见失败模式

- 在 Next API Route 中读取本地源码，破坏架构边界。
- `/api/source` 允许读取项目根目录外文件。
- 每次请求 `/api/graph` 都重新完整扫描仓库，导致性能差。
- entry 模糊匹配规则不明确。
- API 返回字段与前端类型不一致。
- 错误响应是纯文本，前端难处理。
