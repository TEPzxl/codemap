# Stage 03 — Call Extraction

## 目标

从每个函数或方法体中提取调用关系，生成 `from -> to` 的调用边。这个阶段是项目技术核心，需要支持普通函数调用和静态可解析的方法调用。

---

## 本阶段需要完成的内容

### 1. 定义 Call 模型

建议位置：

```text
internal/analyzer/calls.go
```

示例：

```go
type Call struct {
    From       string   `json:"from"`
    To         string   `json:"to"`
    Resolution string   `json:"resolution"`
    Callsite   Callsite `json:"callsite"`
}
```

### 2. 遍历函数体

对每个 `ast.FuncDecl`：

- 找到当前函数对应的 symbol ID。
- 遍历 body。
- 查找所有 `ast.CallExpr`。

### 3. 解析普通函数调用

示例：

```go
validateUser(input)
```

AST 中通常是：

```go
*ast.Ident
```

要求：

- 如果能解析到当前 package 或导入 package 的函数，标记为 `resolved`。
- 如果无法解析，标记为 `unresolved`。

### 4. 解析方法调用

示例：

```go
svc.CreateUser(ctx, input)
```

AST 中通常是：

```go
*ast.SelectorExpr
```

要求：

- 使用 `types.Info` 解析 selector。
- 尽量定位到具体方法 symbol。
- 如果是项目内部方法，标记为 `resolved`。
- 如果是接口方法，标记为 `interface`。
- 如果是标准库或第三方包，标记为 `external`。
- 如果无法解析，标记为 `unresolved`。

### 5. 实现 CLI 命令

```bash
codemap calls <path>
```

输出 calls JSON。

---

## 验收方式

### simple fixture

```bash
go run ./cmd/codemap calls ./examples/simple
```

通过标准：

- 输出合法 JSON。
- 能看到 `main.main -> ...` 的调用边。
- 至少一个 edge 的 `resolution` 为 `resolved`。
- 每个 call 都包含 callsite 文件与行号。

### layered service fixture

```bash
go run ./cmd/codemap calls ./examples/layered-service
```

通过标准：

- 能看到 handler 到 service 的调用。
- 能看到 service 到 repository 的调用。
- 默认不包含标准库调用，或者如果包含，也必须标记为 `external`。

### interface fixture

```bash
go run ./cmd/codemap calls ./examples/interface-call
```

通过标准：

- 接口方法调用不能被错误标记为某个具体实现。
- 如果不能可靠解析到实现，标记为 `interface`。
- 不能 panic。

### 自动化测试

```bash
go test ./internal/analyzer/... -run TestExtractCalls
```

建议断言：

- `from` 是已知 symbol。
- `to` 对 resolved internal call 是已知 symbol。
- interface call 有正确 resolution。
- unresolved call 不会导致整轮分析失败。
- callsite 行号大于 0。

---

## 成功完成标准

本阶段完成时，应满足：

1. `codemap calls <path>` 可用。
2. 能提取普通函数调用。
3. 能提取静态可解析的方法调用。
4. 能区分 `resolved`、`interface`、`external`、`unresolved`。
5. 每个调用边有 callsite。
6. 接口调用处理保守，不做不可靠推断。
7. `go test ./internal/analyzer/... -run TestExtractCalls` 通过。

---

## 常见失败模式

- 只通过文本匹配函数名，导致错误解析。
- 忽略 `go/types`，方法调用无法定位。
- 遇到函数变量调用直接 panic。
- 把接口调用强行解析为第一个实现类。
- 标准库调用混入业务图且未标记 external。
- callsite 使用被调用函数位置，而不是调用发生位置。
