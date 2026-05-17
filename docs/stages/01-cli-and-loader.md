# Stage 01 — CLI and Go Package Loader

## 目标

实现最小可用 CLI，并能扫描本地 Go 仓库，加载 packages，忽略测试代码，输出 packages 和 warnings。这个阶段不要求提取函数调用，只验证项目输入和加载路径是否可靠。

---

## 本阶段需要完成的内容

### 1. 初始化 CLI

建议命令：

```bash
codemap scan <path>
codemap --help
```

如果使用 Cobra，建议目录：

```text
cmd/codemap/main.go
internal/cli/root.go
internal/cli/scan.go
```

如果暂时不用 Cobra，也可以先用标准库 `flag`，但后续正式化时建议迁移到 Cobra。

### 2. 实现 loader

建议位置：

```text
internal/analyzer/loader.go
```

输入：

```go
type LoadRequest struct {
    RootPath     string
    IncludeTests bool
}
```

输出：

```go
type LoadResult struct {
    Packages []PackageInfo
    Warnings []AnalyzeWarning
}
```

要求：

- 支持相对路径。
- 支持绝对路径。
- 能识别本地 Go module。
- 默认 `IncludeTests=false`。
- package 加载失败时记录 warning。
- 不应因单个 package 加载失败导致整个命令崩溃。

### 3. 使用 `go/packages`

建议加载模式至少包含：

```go
packages.NeedName
packages.NeedFiles
packages.NeedSyntax
packages.NeedTypes
packages.NeedTypesInfo
packages.NeedModule
```

可以根据性能后续调整。

### 4. 输出 scan JSON

`codemap scan ./examples/simple` 输出类似：

```json
{
  "root": "./examples/simple",
  "packages": [
    {
      "id": "github.com/acme/simple",
      "name": "main",
      "pkg_path": "github.com/acme/simple",
      "files": ["main.go"]
    }
  ],
  "warnings": []
}
```

---

## 验收方式

### CLI help

```bash
go run ./cmd/codemap --help
```

通过标准：

- 命令可以运行。
- 输出中能看到 `scan` 命令。

### scan fixture

```bash
go run ./cmd/codemap scan ./examples/simple
```

通过标准：

- 输出合法 JSON。
- JSON 包含 `packages` 字段。
- 至少加载到一个 package。
- 不包含 `_test.go` 文件。

### scan layered service

```bash
go run ./cmd/codemap scan ./examples/layered-service
```

通过标准：

- 能加载多个 package，或至少加载 layered fixture 中的核心 package。
- `warnings` 字段存在，即使为空。

### 自动化测试

```bash
go test ./internal/analyzer/... -run TestLoadPackages
```

通过标准：

- loader 可以加载 `examples/simple`。
- loader 默认不包含测试文件。
- invalid path 返回明确 error 或 warning，不 panic。

---

## 成功完成标准

本阶段完成时，应满足：

1. `codemap scan <path>` 可用。
2. loader 能加载本地 Go module。
3. 默认忽略 `_test.go`。
4. 加载失败能生成 warning。
5. scan 输出 JSON。
6. `go test ./internal/analyzer/...` 通过。
7. `go run ./cmd/codemap scan ./examples/simple` 通过。

---

## 常见失败模式

- 直接递归读 `.go` 文件而不是用 `go/packages`，导致类型信息缺失。
- 遇到 package error 直接退出，无法生成 partial result。
- 路径处理只支持当前目录，不支持相对/绝对输入。
- 测试文件混入默认结果。
- 输出日志和 JSON 混在 stdout，导致前端或脚本无法解析。
