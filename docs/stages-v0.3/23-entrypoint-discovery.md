# Stage 23 — Entrypoint Discovery

## 目标

降低用户寻找 entry symbol 的成本。真实项目中，用户往往不知道应该从哪个函数开始看。

## 功能定义

自动发现候选入口点：

- `main` functions。
- exported functions。
- exported methods。
- 常见启动方法：`Run`、`Start`、`Serve`、`Listen`、`Execute`。
- 常见 handler 方法：名称包含 `Handler`、`Handle`，或 receiver 名称包含 `Handler`。
- goroutine starter：函数体中包含 `go <call>` 的函数。

这些都是启发式，不应宣称绝对准确。

## 允许修改范围

可以修改：

- `internal/analyzer` 中新增 entrypoint discovery，不改已有 symbol/call extraction 行为
- `internal/server/*`
- `internal/cli/*`
- `web/*`
- tests
- README

不要修改：

- Graph JSON Schema 必填字段
- 原有 `/api/symbols` 行为，除非做向后兼容扩展

## 推荐模型

```go
type Entrypoint struct {
    ID      string   `json:"id"`
    Label   string   `json:"label"`
    Package string   `json:"package"`
    File    string   `json:"file"`
    Kind    string   `json:"kind"`
    Reasons []string `json:"reasons"`
}
```

`reasons` 示例：

- `main-function`
- `exported-function`
- `exported-method`
- `name:Run`
- `receiver:Handler`
- `contains-goroutine`

## 推荐 CLI

```bash
codemap entrypoints ./repo
```

## 推荐 API

```text
GET /api/entrypoints
```

## UI 行为

新增 `Entrypoints` 列表或 tab：

- 按 reason 分组。
- 优先显示 main functions。
- 点击 entrypoint 自动设置 entry 并加载 graph。
- 保留 symbol search。

## 测试要求

至少覆盖：

- examples 中识别 main。
- layered-service 中识别 handler/service exported methods。
- pure Go 项目中识别 `Run`/`Start`/`Serve` 等方法。
- reasons 正确。
- 不包含 `_test.go`。

## 验收命令

```bash
make check
make web-build
make build
```

手动验收：

```bash
./bin/codemap entrypoints ./examples/layered-service
curl -s http://localhost:8080/api/entrypoints
```

Web 中点击 entrypoint 后 graph 自动加载。

## 成功标准

- 用户不需要手动知道完整 symbol id 也能开始探索。
- entrypoint discovery 是启发式且有 reasons。
- 不破坏 existing symbols API。

## 常见失败模式

- 把所有 exported symbols 都堆在最前面，main 不突出。
- reasons 不清晰，用户不知道为什么推荐该入口。
- 使用 display label 作为 id，导致点击后 graph 找不到 entry。
