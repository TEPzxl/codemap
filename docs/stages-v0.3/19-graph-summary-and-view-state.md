# Stage 19 — Graph Summary and View State

## 目标

让用户理解“当前正在看的图有多大、由哪些 resolution 类型组成、当前视图参数是什么”。同时让 URL 保存当前视图状态，方便刷新、复制和分享。

## 允许修改范围

可以修改：

- `web/*`
- `README.md` 中与 UI 使用相关的小节
- 必要的前端测试或类型文件

除非文档明确要求，不要修改：

- Go analyzer
- graph traversal
- server API
- Graph JSON Schema

本阶段优先在前端从现有 `/api/graph` 返回值计算 summary，不要求新增后端 API。

## 需要完成

### 1. 当前图摘要

在 Web UI 中增加 `Current graph` 或 `Graph summary` 面板，显示：

- 当前 entry。
- depth。
- nodes count。
- edges count。
- external edges count。
- unresolved edges count。
- interface edges count。
- resolved edges count。
- package count。
- 当前开启的 filters。

### 2. URL view state

将以下状态同步到 URL query：

- `entry`
- `depth`
- `show_external`
- `show_unresolved`
- `show_interface`
- `expand_interface`
- `package`

刷新页面后应恢复这些状态并自动加载 graph。

### 3. Copy current view

增加 `Copy view URL` 按钮，复制当前完整 URL。

### 4. 自动刷新约束

- filter 改变后可以自动刷新。
- depth 改变后应 debounce。
- entry 手动输入不应每输入一个字符就请求；点击候选或按 Enter 时加载。

## 测试要求

- 前端类型检查通过。
- URL 中存在 entry/depth/filter 时，页面初始化能恢复状态。
- filter 改变后 URL 随之更新。
- graph summary 统计值与当前 graph JSON 一致。

## 验收命令

```bash
make check
make web-build
make build
```

手动验收：

```bash
./bin/codemap serve ./examples/layered-service --port 8080
```

打开：

```text
http://localhost:8080?entry=main.main&depth=5&show_external=true
```

确认 UI 恢复状态并显示 graph summary。

## 成功标准

- 当前 graph 的规模和 resolution 分布清晰可见。
- URL 能恢复当前视图。
- Copy view URL 可用。
- 没有修改 Go 核心逻辑。

## 常见失败模式

- graph summary 使用过期数据。
- 修改 filter 后 URL 没变。
- 刷新页面后 entry 丢失。
- entry 输入每个字符都请求 API，造成大量 invalid entry 请求。
