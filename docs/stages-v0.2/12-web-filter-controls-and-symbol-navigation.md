# Stage 12 — Web Filter Controls and Symbol Navigation

## 目标

让 Web UI 不再依赖用户手动输入完整 symbol id，而是提供更实用的 symbol 搜索、package 导航和图过滤控件。

本阶段主要修改 `web/`，必要时可小范围扩展 API，但不要修改 analyzer 核心逻辑。

## 需要完成

### Symbol 搜索

1. 从 `/api/symbols` 加载 symbol 列表。
2. 提供搜索输入框。
3. 支持按以下字段搜索：
   - symbol label
   - full id
   - package
   - file
   - receiver
4. 选择结果后自动填充 entry symbol。
5. 提供“复制 symbol id”按钮。

### Package 导航

1. 从 symbols 中归纳 package 列表。
2. 提供 package filter。
3. 选中 package 后，symbol 搜索结果只显示该 package 内 symbol。
4. graph 请求可附带 package filter，若 Stage 11 已支持。

### Graph 控件

增加 UI toggles：

```text
[ ] Show external calls
[ ] Show unresolved calls
[ ] Show interface calls
[ ] Expand interface candidates
```

这些控件映射到 `/api/graph` query 参数：

```text
show_external=true
show_unresolved=true
show_interface=true
expand_interface=true
```

### URL 状态

可选但推荐：把当前 entry、depth、filters 写入 URL query：

```text
/?entry=main.main&depth=5&show_external=true
```

页面刷新后能恢复当前查询。

## UI 要求

- 输入框不能过窄，完整 symbol id 应可横向滚动或以 tooltip 展示。
- 搜索结果要显示短名与 package。
- 选择 entry 后不要立即强制 reload，用户点击 Load graph 或支持自动加载均可，但行为要明确。
- loading/error/empty 状态必须存在。

## 测试要求

如果前端测试框架未建立，本阶段至少保证：

```bash
cd web && corepack pnpm lint
cd web && corepack pnpm typecheck
cd web && corepack pnpm build
```

如果已有前端测试工具，则覆盖：

1. symbol search filter。
2. package filter。
3. graph query 参数拼接。
4. toggle 状态映射。

## 验收命令

```bash
make check
make web-build
make build
./bin/codemap serve ./examples/layered-service --port 8080
```

浏览器打开：

```text
http://localhost:8080
```

人工验收：

1. 不输入完整 id，也能搜索并选择 `main`。
2. 可以搜索 `CreateUser`。
3. 可以按 package 缩小结果。
4. 打开 `Show external calls` 后 graph 变化。
5. 打开 `Expand interface candidates` 后 interface fixture 中出现候选实现。
6. 点击节点后源码仍然显示。

## 成功标准

- Web UI 可用性明显提升。
- 用户无需复制完整 symbol id。
- UI filters 与 API 行为一致。
- 不破坏 v0.1 的源码面板。

## 常见失败模式

- symbol 搜索只搜 label，不搜 full id。
- UI toggle 状态没有传给 API。
- API 错误导致页面白屏。
- graph reload 后 selected node 源码残留，不属于当前 graph。
- package filter 过滤 UI 但没有影响 graph 请求。

## Codex 提示词

```text
现在开始 Stage 12：Web Filter Controls and Symbol Navigation。

主要修改 web/。实现 symbol 搜索、package filter、graph filter toggles。不要修改 analyzer/graph 核心逻辑，不要修改 Graph JSON Schema。完成后运行 make check、web lint/typecheck/build，并人工说明如何验证 UI。
```
