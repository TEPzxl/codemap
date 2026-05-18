# codemap v0.2 阶段文档索引

本文档集用于在 `codemap v0.1.0` 已完成的基础上，指导 Codex / AI agent 按阶段实现 `v0.2`。

## v0.1 当前基线

v0.1 已完成：

- 本地 Go 仓库扫描。
- `scan` / `symbols` / `calls` / `graph` / `serve` CLI。
- 基于 `go/packages`、AST、类型信息的函数与方法调用关系提取。
- Graph JSON Schema。
- 本地 HTTP API：`/api/health`、`/api/symbols`、`/api/graph`、`/api/source`、`/api/warnings`。
- Next + React + TypeScript + Tailwind + React Flow 前端。
- 点击图节点后显示源码片段。
- Go 二进制服务 Web UI 与 `/api/*`。
- `make check` / smoke / demo / README。

## v0.2 总目标

v0.2 不再做基础闭环，而是提高项目的“代码理解能力”和“产品可用性”。主要目标：

1. 改进 interface call 的候选实现解析。
2. 增加 CLI/API/UI 层面的图过滤能力。
3. 优化入口 symbol 搜索和 package 导航。
4. 提升 React Flow 布局和交互体验。
5. 支持 edge callsite 源码查看。
6. 增加性能指标、缓存刷新和项目元信息。
7. 完善 CI / release / demo / v0.2 质量门禁。

## v0.2 非目标

v0.2 仍然不做：

- 远程 GitHub URL 扫描。
- LLM 解释代码。
- VSCode / Zed 插件。
- 数据库持久化。
- 完整运行时动态调用图。
- 测试代码默认纳入分析。
- 变量级引用、字段读写追踪、完整 IDE 级 Find References。

## 阶段顺序

建议严格按顺序推进：

1. [Stage 09 — v0.2 Baseline and Compatibility Contract](./09-v0.2-baseline-and-contract.md)
2. [Stage 10 — Interface Implementation Resolution](./10-interface-implementation-resolution.md)
3. [Stage 11 — Graph Filters in CLI and API](./11-graph-filters-cli-api.md)
4. [Stage 12 — Web Filter Controls and Symbol Navigation](./12-web-filter-controls-and-symbol-navigation.md)
5. [Stage 13 — Graph Layout and Interaction Polish](./13-graph-layout-and-interaction-polish.md)
6. [Stage 14 — Source Viewer and Callsite Inspection](./14-source-viewer-and-callsite-inspection.md)
7. [Stage 15 — Performance, Cache, Rescan, and Metadata](./15-performance-cache-rescan-metadata.md)
8. [Stage 16 — CI, Release, and v0.2 Demo Docs](./16-ci-release-demo-docs.md)
9. [Stage 17 — v0.2 Quality Gate](./17-v0.2-quality-gate.md)

## 执行方式

每个阶段都应遵守：

```bash
git status --short
make check
```

每个阶段完成后：

```bash
git add .
git commit -m "stage XX: <summary>"
```

不要多个 agent 同时修改以下区域：

- `internal/analyzer/`
- `internal/graph/`
- `internal/server/`
- Graph JSON Schema
- `web/types/graph.ts`

可以并行的区域：

- README / docs
- Web UI 纯展示优化
- CI / scripts
- demo 截图和文档

## Schema 原则

v0.2 可以对 Graph JSON Schema 做**向后兼容的可选字段扩展**，但不能破坏 v0.1 已有字段。

允许：

- 增加 optional 字段。
- 增加新的 API query 参数。
- 增加 CLI flag。
- 增加新的 endpoint。

不允许：

- 删除 `nodes` / `edges` / `warnings`。
- 改变既有字段含义。
- 让 v0.1 的 CLI 命令失效。
- 让 `make check` 依赖不可复现的外部状态。
