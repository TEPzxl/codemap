# Stage 22 — Package Overview and Collapse

## 目标

为大项目提供 package 级别的结构视图。用户可以先看 package 之间的调用关系，再进入函数/方法级别。

## 功能定义

新增 package-level graph：

```text
cmd/server -> internal/handler -> internal/service -> internal/repository
```

package edge 表示：

> package A 中的某个函数/方法调用了 package B 中的某个函数/方法。

## 允许修改范围

可以修改：

- `internal/graph/*`
- `internal/server/*`
- `internal/cli/*`
- `web/*`
- tests
- README

不要修改：

- analyzer extraction 核心逻辑
- 原有 function-level Graph JSON 必填字段

## 推荐 CLI

```bash
codemap packages ./repo
codemap packages ./repo --entry main.main --depth 5
```

输出 package graph JSON。可以使用独立 schema，例如：

```json
{
  "nodes": [
    {"id": "internal/service", "package": "internal/service", "symbols": 42}
  ],
  "edges": [
    {"from": "internal/handler", "to": "internal/service", "calls": 7}
  ],
  "warnings": []
}
```

## 推荐 API

```text
GET /api/package-graph?entry=<symbol>&depth=5
```

如果不提供 entry，可以返回全项目 package dependency/call overview。

## UI 行为

新增视图切换：

- Function graph
- Package graph

Package graph 中：

- 节点显示 package 相对路径。
- 节点显示 symbol/call 数量。
- 边显示 call count。
- 点击 package 节点时，左侧 package filter 自动设置为该 package。
- 双击 package 节点可以切回 function graph 并过滤/定位该 package。

## 聚合规则

- 仅聚合当前 graph 中的 edges，或按 query 参数聚合全项目。
- 同 package 内调用可以默认隐藏，也可以显示为 self edge 计数；v0.3 建议默认隐藏 self edge。
- external/unresolved/interface 是否计入由 filters 决定。

## 测试要求

至少覆盖：

- layered-service package graph 包含 handler -> service -> repository。
- self edge 默认隐藏。
- call count 正确累计。
- package filter 不破坏 function graph 的跨 package 调用链。
- `/api/package-graph` JSON error。

## 验收命令

```bash
make check
make web-build
make build
```

手动验收：

```bash
./bin/codemap packages ./examples/layered-service
curl -s "http://localhost:8080/api/package-graph?entry=main.main&depth=5"
```

Web 中切换到 Package graph，确认能看到 package-level 调用结构。

## 成功标准

- 用户可以先理解 package 层结构。
- function graph 与 package graph 可以切换。
- package graph 不破坏现有 function graph。

## 常见失败模式

- package graph 直接使用 import graph，而不是 call graph 聚合。
- package id 使用绝对路径，导致 UI 过长。
- package filter 裁剪掉跨 package 调用链。
- self edge 太多导致图不可读。
