# Stage 00 — Contract, Schema, Fixtures

## 目标

在开始实现 analyzer 和 UI 之前，先固定 `codemap` 的核心协议、fixture 项目和验收样例。这个阶段的目标不是实现完整功能，而是避免后端、HTTP API 和前端之间反复返工。

---

## 本阶段需要完成的内容

### 1. 建立 graph 数据模型

在 Go 后端中定义基础模型，建议位置：

```text
internal/graph/model.go
```

需要包含：

- `Graph`
- `Node`
- `Edge`
- `Callsite`
- `Warning`
- `NodeKind`
- `EdgeResolution`

推荐字段：

```go
type Graph struct {
    Entry    string    `json:"entry"`
    Nodes    []Node    `json:"nodes"`
    Edges    []Edge    `json:"edges"`
    Warnings []Warning `json:"warnings"`
}

type Node struct {
    ID         string `json:"id"`
    Label      string `json:"label"`
    Kind       string `json:"kind"`
    Package    string `json:"package"`
    Receiver   string `json:"receiver,omitempty"`
    File       string `json:"file"`
    StartLine  int    `json:"start_line"`
    EndLine    int    `json:"end_line"`
    IsExternal bool   `json:"is_external"`
}

type Edge struct {
    ID         string   `json:"id"`
    From       string   `json:"from"`
    To         string   `json:"to"`
    Kind       string   `json:"kind"`
    Resolution string   `json:"resolution"`
    Callsite   Callsite `json:"callsite"`
}
```

### 2. 建立 TypeScript 类型

建议位置：

```text
web/types/graph.ts
```

字段应与 Go JSON schema 对齐。

### 3. 建立 fixture 项目

创建以下 fixture：

```text
examples/simple/
examples/layered-service/
examples/interface-call/
```

#### `examples/simple`

用于验证普通函数和简单方法调用。

期望调用链：

```text
main.main -> app.Run -> service.CreateUser
```

#### `examples/layered-service`

用于展示简历项目核心 demo。

期望调用链：

```text
main.main
  -> handler.CreateUser
  -> service.CreateUser
  -> repository.Save
```

#### `examples/interface-call`

用于验证接口调用被标记为 `interface` 或 conservative resolution。

期望结果：

```text
service.CreateUser -> UserRepository.Save
```

可以暂时不展开到具体实现。

### 4. 建立样例 JSON

建议位置：

```text
docs/examples/graph.simple.json
docs/examples/graph.layered-service.json
```

这些样例用于前端初始开发和后端回归测试。

---

## 验收方式

### 自动化检查

```bash
go test ./internal/graph/...
```

通过标准：

- Graph model 能被 JSON marshal / unmarshal。
- `resolution` 只允许：
  - `resolved`
  - `interface`
  - `external`
  - `unresolved`
- Node kind 只允许：
  - `function`
  - `method`
  - `external`
  - `unresolved`

### JSON 结构检查

可以增加测试：

```bash
go test ./internal/graph/... -run TestGraphJSONRoundTrip
```

通过标准：

- `Graph -> JSON -> Graph` 后数据不丢失。
- `nodes` 和 `edges` 字段存在。
- 每个 edge 的 `from` 和 `to` 能对应到节点，除非该 edge 是显式 external/unresolved 策略的一部分。

### Fixture 检查

```bash
find examples -maxdepth 2 -name "go.mod" -print
```

通过标准：

- `examples/simple/go.mod` 存在。
- `examples/layered-service/go.mod` 存在。
- `examples/interface-call/go.mod` 存在。
- 每个 fixture 可以 `go test ./...`。

---

## 成功完成标准

本阶段完成时，应满足：

1. Go graph model 已定义。
2. TypeScript graph type 已定义。
3. 至少三个 fixture 项目已创建。
4. 至少两个样例 graph JSON 已创建。
5. `go test ./internal/graph/...` 通过。
6. 每个 fixture 的 `go test ./...` 通过。
7. 后续阶段不得随意改 schema；如需修改，必须同步 Go、TypeScript、样例 JSON 和测试。

---

## 常见失败模式

- 一开始把源码全文放进 node，导致 graph JSON 过大。
- node ID 使用短名，导致不同 package 下同名函数冲突。
- TypeScript 类型和 Go JSON 字段不一致。
- fixture 过于复杂，导致早期 analyzer 无法验证。
- 忽略 interface call fixture，后期才发现 schema 无法表达不确定调用。
