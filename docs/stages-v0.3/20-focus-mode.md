# Stage 20 — Focus Mode

## 目标

让用户从任意节点开始聚焦探索，而不是只能从原始 entry 展开。Focus Mode 是 v0.3 的核心功能。

## 功能定义

用户选中一个节点后，可以执行：

- `Focus downstream`：只显示该节点的下游调用链。
- `Focus upstream`：只显示哪些节点调用了该节点。
- `Focus neighborhood`：显示该节点上下游半径 N 内的邻域。
- `Reset to entry`：回到原始 entry graph。

## 允许修改范围

可以修改：

- `internal/graph/*`
- `internal/server/*`
- `internal/cli/*`
- `web/*`
- tests
- README 的使用说明

不要修改：

- analyzer symbol/call extraction 核心逻辑
- Graph JSON Schema 的必填字段

新增 API 和 CLI flag 可以是向后兼容的。

## 推荐 API 设计

可以扩展 `/api/graph` query：

```text
GET /api/graph?entry=<node_id>&depth=5&direction=downstream
GET /api/graph?entry=<node_id>&depth=3&direction=upstream
GET /api/graph?entry=<node_id>&depth=2&direction=both
```

也可以新增 `/api/focus`，但优先复用 `/api/graph`，减少 API 面。

`direction` 合法值：

- `downstream`，默认值，保持 v0.2 行为。
- `upstream`。
- `both`。

## CLI 行为

扩展：

```bash
codemap graph ./repo --entry <symbol> --depth 3 --direction downstream
codemap graph ./repo --entry <symbol> --depth 3 --direction upstream
codemap graph ./repo --entry <symbol> --depth 2 --direction both
```

默认不传 `--direction` 时必须保持 v0.2 行为。

## UI 行为

选中节点后，在详情面板或 toolbar 中显示：

- Focus downstream
- Focus upstream
- Focus neighborhood
- Reset to entry

执行 focus 后：

- 图更新为 focus graph。
- URL query 更新 `entry` 和 `direction`。
- 保留 source panel 或重新加载所选节点 source。
- 顶部或 summary panel 明确显示当前模式：`Focus: upstream/downstream/both`。

## 测试要求

至少覆盖：

- downstream 与 v0.2 默认行为一致。
- upstream 能返回调用当前节点的上游路径。
- both 能返回上下游邻域。
- depth=0 只返回 entry 节点。
- 循环调用不会死循环。
- unknown entry 返回 JSON error。
- CLI `--direction` 校验。

## 验收命令

```bash
make check
make web-build
make build
./scripts/smoke.sh
```

手动验收：

```bash
./bin/codemap serve ./examples/layered-service --port 8080
```

打开 Web UI，选择 `UserService.CreateUser`：

- Focus upstream 应看到 `UserHandler.CreateUser` 等上游。
- Focus downstream 应看到 repository 调用。
- Reset to entry 应恢复原始 graph。

## 成功标准

- 可以从任意节点重新聚焦。
- upstream/downstream/both 都可用。
- v0.2 默认 graph 行为不变。
- UI、API、CLI 三者行为一致。

## 常见失败模式

- upstream 仍然返回 downstream。
- both 模式重复节点或重复边过多。
- focus 后 URL 没更新。
- focus 使用 display label，而不是完整 node id。
- source lookup 使用短名导致失败。
