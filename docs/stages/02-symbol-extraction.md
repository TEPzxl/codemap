# Stage 02 — Symbol Extraction

## 目标

从已加载的 Go packages 中提取函数和方法 symbol。每个 symbol 需要包含稳定 ID、展示名、类型、包路径、接收者、文件路径、开始行和结束行。

---

## 本阶段需要完成的内容

### 1. 定义 Symbol 模型

建议位置：

```text
internal/analyzer/symbols.go
```

示例：

```go
type Symbol struct {
    ID        string `json:"id"`
    Label     string `json:"label"`
    Kind      string `json:"kind"` // function | method
    Package   string `json:"package"`
    Receiver  string `json:"receiver,omitempty"`
    File      string `json:"file"`
    StartLine int    `json:"start_line"`
    EndLine   int    `json:"end_line"`
}
```

### 2. 提取普通函数

识别：

```go
func CreateUser() {}
```

生成：

```text
<package-path>.CreateUser
```

### 3. 提取方法

识别：

```go
func (s *UserService) CreateUser() {}
```

生成：

```text
<package-path>.(*UserService).CreateUser
```

UI label 可以是：

```text
UserService.CreateUser
```

### 4. 记录源码位置

使用 `go/token.FileSet` 计算：

- 文件路径
- 起始行
- 结束行

### 5. 实现 CLI 命令

```bash
codemap symbols <path>
```

输出 symbol JSON。

---

## 验收方式

### simple fixture

```bash
go run ./cmd/codemap symbols ./examples/simple
```

通过标准：

- 输出合法 JSON array 或 object。
- 包含 `main.main`。
- 包含至少一个普通函数或方法。
- 每个 symbol 都有：
  - `id`
  - `label`
  - `kind`
  - `package`
  - `file`
  - `start_line`
  - `end_line`

### layered service fixture

```bash
go run ./cmd/codemap symbols ./examples/layered-service
```

通过标准：

- 能找到 handler 层函数或方法。
- 能找到 service 层函数或方法。
- 能找到 repository 层函数或方法。
- 不包含 `_test.go` 中的 symbol。

### 自动化测试

```bash
go test ./internal/analyzer/... -run TestExtractSymbols
```

建议断言：

- `main.main` 存在。
- 方法接收者格式正确。
- `start_line <= end_line`。
- ID 没有重复。
- 同名函数位于不同 package 时 ID 不冲突。

---

## 成功完成标准

本阶段完成时，应满足：

1. `codemap symbols <path>` 可用。
2. 普通函数可以被提取。
3. 方法可以被提取。
4. 每个 symbol 具备源码定位。
5. symbol ID 稳定且唯一。
6. 默认不包含测试代码。
7. `go test ./internal/analyzer/... -run TestExtractSymbols` 通过。

---

## 常见失败模式

- 方法 ID 不包含 receiver，导致同名方法冲突。
- 只记录开始行，不记录结束行，导致源码面板无法准确截取。
- 使用短包名作为 ID 的一部分，导致跨 package 冲突。
- 忽略指针接收者和非指针接收者的差异。
- UI label 和内部 ID 混用。
