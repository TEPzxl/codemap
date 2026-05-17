# Stage 06 — Web UI with Next, React Flow, TypeScript, Tailwind

## 目标

实现可交互的前端页面：用户可以搜索入口 symbol、请求调用图、通过 React Flow 展示业务调用链、点击节点后在底部查看源码片段。

---

## 本阶段需要完成的内容

### 1. 初始化前端

建议：

```bash
cd web
pnpm install
pnpm dev
```

使用：

- Next.js
- React
- TypeScript
- Tailwind CSS
- React Flow
- CodeMirror 或 `<pre>`

### 2. 页面布局

推荐布局：

```text
+------------------------------------------------+
| Header: codemap                                |
+-------------------+----------------------------+
| Symbol Search     | React Flow Graph           |
| Depth Control     |                            |
| Filters           |                            |
| Warnings          |                            |
+-------------------+----------------------------+
| Source Panel                                    |
+------------------------------------------------+
```

### 3. API client

建议位置：

```text
web/lib/api.ts
```

需要实现：

```ts
fetchSymbols()
fetchGraph(entry: string, depth: number)
fetchSource(nodeId: string)
fetchWarnings()
```

### 4. 类型定义

建议位置：

```text
web/types/graph.ts
```

必须与 Go schema 对齐。

### 5. React Flow

建议组件：

```text
web/components/GraphView.tsx
```

要求：

- 将后端 nodes/edges 转为 React Flow nodes/edges。
- 节点 label 显示短名。
- 点击节点触发 source 请求。
- 支持 fit view。
- 空图时显示 empty state。
- 错误时显示 error state。

### 6. Symbol Search

建议组件：

```text
web/components/SymbolSearch.tsx
```

要求：

- 从 `/api/symbols` 获取候选。
- 支持按 label / package / file 搜索。
- 选择后请求 graph。

### 7. Source Panel

建议组件：

```text
web/components/SourcePanel.tsx
```

要求：

- 显示当前选中 symbol。
- 显示 file、start_line、end_line。
- 显示 Go 源码。
- 未选中节点时显示提示。

### 8. Warnings

建议组件：

```text
web/components/WarningPanel.tsx
```

要求：

- 显示 package loading warning。
- 显示 unresolved call summary。
- 不阻断主图使用。

---

## 验收方式

### 前端开发启动

```bash
cd web
pnpm dev
```

通过标准：

- 页面可以打开。
- 无 TypeScript 编译错误。
- Tailwind 样式生效。

### 类型检查

```bash
cd web
pnpm typecheck
```

通过标准：

- TypeScript 无错误。
- Graph schema 类型与 API client 使用一致。

### lint

```bash
cd web
pnpm lint
```

通过标准：

- 无严重 lint error。

### build

```bash
cd web
pnpm build
```

通过标准：

- Next build 成功。
- 无阻断性错误。

### 手动集成测试

先启动 Go API：

```bash
go run ./cmd/codemap serve ./examples/layered-service --port 8080
```

再启动前端：

```bash
cd web
pnpm dev
```

浏览器验证：

1. 页面能加载 symbols。
2. 搜索 `main.main` 或目标入口。
3. 点击入口后能看到 React Flow 图。
4. 调整 depth 后图发生变化。
5. 点击某个节点后，底部源码面板显示源码。
6. warnings 区域能显示 warning，或者显示空状态。

---

## 成功完成标准

本阶段完成时，应满足：

1. Next 页面可运行。
2. React Flow 能展示后端 graph。
3. Symbol 搜索可用。
4. Depth 控制可用。
5. 节点点击后源码面板可用。
6. Warnings 可展示。
7. `pnpm typecheck` 通过。
8. `pnpm build` 通过。
9. 前端不依赖 Next API Routes 扫描本地仓库。

---

## 常见失败模式

- 把分析逻辑放进前端或 Next API Routes。
- React Flow 节点 ID 与后端 node ID 不一致，导致点击查源码失败。
- 图没有布局，所有节点重叠。
- 每次输入搜索字符都请求 graph，导致 API 频繁调用。
- 忽略 loading / error / empty 状态。
- TypeScript 类型和后端 JSON 不一致。
