# Stage 14 — Source Viewer and Callsite Inspection

## 目标

v0.1 只能点击节点查看函数/方法源码。v0.2 应支持点击边查看 callsite，即“这条调用发生在调用方源码的哪一行”。

本阶段提升代码阅读体验。

## 需要完成

### 后端 API

新增或扩展 source API。

推荐新增：

```text
GET /api/callsite?edge_id=<edge_id>
```

返回：

```json
{
  "edge_id": "edge-000001",
  "file": "internal/service/user.go",
  "line": 16,
  "column": 20,
  "start_line": 13,
  "end_line": 19,
  "source": "...",
  "highlight_line": 16
}
```

也可以扩展 `/api/source` 支持 `edge_id`，但推荐独立 endpoint，语义更清楚。

### 安全要求

- 文件读取必须限制在项目 root 内。
- path traversal 必须拒绝。
- unknown edge id 返回 JSON 404。
- invalid request 返回 JSON 400。

### Web UI

1. 点击 node：继续显示 node source。
2. 点击 edge：显示 callsite source。
3. source panel 标识当前模式：
   - `Node source`
   - `Callsite`
4. callsite 行应高亮或至少显示 `line:column`。
5. source panel 应保持换行、缩进和 monospace。
6. 可增加 copy file path / copy symbol id。

### 可选增强

- 用 CodeMirror 或轻量 syntax highlight 展示 Go 代码。
- 显示 line numbers。
- 显示调用方和被调用方 symbol。

## 测试要求

后端测试至少覆盖：

1. valid edge id 返回 callsite snippet。
2. unknown edge id 返回 JSON 404。
3. source line range 包含 callsite line。
4. path traversal 被拒绝。
5. API 不返回项目 root 外文件。

前端至少通过：

```bash
cd web && corepack pnpm lint
cd web && corepack pnpm typecheck
cd web && corepack pnpm build
```

## 验收命令

```bash
make check
make web-build
make build
./bin/codemap serve ./examples/layered-service --port 8080
curl -s "http://localhost:8080/api/graph?entry=main.main&depth=5" | python -m json.tool
curl -s "http://localhost:8080/api/callsite?edge_id=<real_edge_id>" | python -m json.tool
```

人工验收：

1. 点击 node 显示函数源码。
2. 点击 edge 显示调用发生位置。
3. callsite line 能定位到源码里的调用表达式。
4. unknown edge id 不导致 UI 白屏。

## 成功标准

- 节点源码和边调用点源码都可查看。
- 用户能知道某条边来自源码哪一行。
- 安全限制仍然有效。
- 不破坏 v0.1 `/api/source?node_id=...`。

## 常见失败模式

- edge id 不稳定，前端点击后找不到。
- edge click 被 node click 覆盖。
- source panel 残留旧 node source。
- callsite snippet 只返回一行，缺少上下文。
- 后端直接使用 query file path 读取文件，造成安全风险。

## Codex 提示词

```text
现在开始 Stage 14：Source Viewer and Callsite Inspection。

新增 edge callsite API 和前端 edge 点击源码展示。保留 v0.1 node source 行为。文件读取必须限制在项目 root 内。不要修改 Graph JSON Schema，除非只添加可选字段且说明原因。完成后运行 make check、web build 和 callsite API 验收命令。
```
