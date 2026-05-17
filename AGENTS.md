# AGENTS.md — codemap AI Agent Project Guide

本文件是 `codemap` 项目的 AI 编码代理协作说明。任何 AI agent、IDE coding agent、命令行 coding agent 在修改本仓库前，必须先读取本文件，并遵守其中的项目边界、技术栈、目录约定、阶段目标和验收方式。

---

## 1. 项目一句话

`codemap` 是一个本地优先的 Go 静态分析与调用关系可视化工具。它扫描本地 Go 仓库，提取静态可解析的函数调用关系与方法调用关系，并通过 React Flow 交互式展示业务调用链。用户点击图中节点时，底部源码面板展示该函数或方法的源码片段。

---

## 2. 核心目标

v1 只聚焦以下目标：

1. 扫描本地 Go module / Go workspace。
2. 默认忽略 `_test.go` 测试代码。
3. 提取函数与方法 symbol。
4. 提取静态可解析的函数调用与方法调用关系。
5. 构建中粒度业务调用链图。
6. 默认隐藏标准库调用，默认隐藏第三方依赖调用。
7. 提供 CLI 查询能力。
8. 提供本地 HTTP API。
9. 使用 Next.js + React + TypeScript + Tailwind + React Flow 展示调用图。
10. 点击图节点后展示对应源码片段。
11. 当部分 package 加载失败时生成 partial graph 和 warnings，而不是直接崩溃。

---

## 3. 非目标

v1 不做以下内容：

- 不引入 LLM 解释代码。
- 不支持远程 GitHub URL 扫描。
- 不做变量级引用分析。
- 不做字段读写追踪。
- 不做完整运行时动态调用推断。
- 不保证接口调用一定解析到具体实现类。
- 不分析测试文件，除非未来显式增加参数。
- 不做 VSCode / JetBrains / Zed 插件。
- 不依赖数据库。
- 不使用 Next API Routes 作为核心后端。
- 不把源码全文塞进 graph JSON；源码通过 API 懒加载。

---

## 4. 技术栈

### 后端 / CLI

- Go
- `go/packages`
- `go/ast`
- `go/types`
- `go/token`
- `net/http`
- `encoding/json`
- `embed`
- Cobra，可用于正式 CLI 命令组织

### 前端

- Next.js
- React
- TypeScript
- Tailwind CSS
- React Flow
- CodeMirror，v1 推荐用于源码展示
- Monaco Editor，可作为后续增强

---

## 5. 推荐目录结构

```text
codemap/
  AGENTS.md

  cmd/
    codemap/
      main.go

  internal/
    cli/
      root.go
      scan.go
      symbols.go
      graph.go
      serve.go

    analyzer/
      loader.go
      symbols.go
      calls.go
      resolver.go
      warning.go

    graph/
      model.go
      builder.go
      traversal.go
      filter.go
      schema.go

    source/
      reader.go
      snippet.go

    server/
      server.go
      handlers.go
      static.go

  web/
    app/
      page.tsx
      layout.tsx
      globals.css
    components/
      GraphView.tsx
      SourcePanel.tsx
      SymbolSearch.tsx
      Toolbar.tsx
      WarningPanel.tsx
    lib/
      api.ts
      layoutGraph.ts
    types/
      graph.ts

  examples/
    simple/
    layered-service/
    interface-call/

  docs/
    stages/
      00-contract-and-fixtures.md
      01-cli-and-loader.md
      02-symbol-extraction.md
      03-call-extraction.md
      04-graph-query-and-filtering.md
      05-http-api-and-source.md
      06-web-ui.md
      07-packaging-release-and-demo.md
      08-quality-gates.md

  go.mod
  README.md
```

---

## 6. 数据流

```text
Local Go Repo
   |
   v
go/packages load
   |
   v
AST + type information
   |
   v
Function / method symbols
   |
   v
Call expressions
   |
   v
Call graph
   |
   v
CLI output / HTTP API
   |
   v
Next + React Flow UI
   |
   v
Source snippet panel
```

---

## 7. 核心数据模型原则

### 7.1 Symbol ID

内部 symbol ID 必须稳定、可唯一定位，建议使用完整限定名：

```text
<module-or-import-path>.<function>
<module-or-import-path>.<receiver>.<method>
```

示例：

```text
github.com/acme/app/cmd/api.main
github.com/acme/app/internal/service.(*UserService).CreateUser
github.com/acme/app/internal/repository.(*UserRepo).Save
```

UI 可以展示短名，但后端协议必须使用稳定 ID。

### 7.2 Node

Graph node 表示函数或方法，不表示每一行代码。

```json
{
  "id": "github.com/acme/app/internal/service.(*UserService).CreateUser",
  "label": "UserService.CreateUser",
  "kind": "method",
  "package": "github.com/acme/app/internal/service",
  "receiver": "*UserService",
  "file": "internal/service/user.go",
  "start_line": 24,
  "end_line": 61,
  "is_external": false
}
```

### 7.3 Edge

Graph edge 表示一次或多次调用关系。

```json
{
  "id": "edge-001",
  "from": "github.com/acme/app/cmd/api.main",
  "to": "github.com/acme/app/internal/server.(*Server).Run",
  "kind": "call",
  "resolution": "resolved",
  "callsite": {
    "file": "cmd/api/main.go",
    "line": 18,
    "column": 9
  }
}
```

`resolution` 只能使用以下值：

```text
resolved
interface
external
unresolved
```

含义：

- `resolved`：确定解析到项目内函数或方法。
- `interface`：解析到接口方法，但暂未展开到具体实现。
- `external`：标准库或第三方依赖调用。
- `unresolved`：无法解析的调用，例如函数变量调用、复杂动态分派。

### 7.4 Source Snippet

graph JSON 不应包含完整源码。源码通过以下 API 懒加载：

```text
GET /api/source?node_id=<symbol-id>
```

---

## 8. CLI 命令目标

v1 推荐命令：

```bash
codemap scan <path>
codemap symbols <path>
codemap calls <path>
codemap graph <path> --entry <symbol-id-or-query> --depth 5
codemap serve <path> --port 8080
```

命令行为：

- `scan`：验证 loader，可输出 packages 和 warnings。
- `symbols`：输出函数/方法 symbol。
- `calls`：输出调用边。
- `graph`：输出可视化 graph JSON。
- `serve`：启动本地 HTTP API 和前端页面。

---

## 9. 后端实现约束

### 9.1 Loader

- 使用 `go/packages` 加载本地 Go 项目。
- 默认不加载 `_test.go`。
- package 加载失败时生成 warning，不应直接终止整个扫描。
- 路径必须支持相对路径和绝对路径。
- 建议优先支持当前 module 内 package。

### 9.2 Analyzer

- 使用 `go/ast` 遍历语法树。
- 使用 `go/types` 辅助解析 selector 和方法调用。
- 记录每个 symbol 的文件路径、开始行、结束行。
- 对无法解析的调用保留 unresolved edge，而不是丢弃所有信息。

### 9.3 Graph

- 使用 adjacency list。
- 支持 depth-limited traversal。
- 支持 visited 去重。
- 支持 cycle 标记或至少防止无限递归。
- 默认过滤标准库和第三方依赖。
- 允许通过参数显示 external / unresolved。

### 9.4 Server

- 使用 Go `net/http` 即可。
- API 返回 JSON。
- 错误响应必须包含 `error` 字段。
- warnings 应可从 API 获取。
- 发布版本中 Go 可以通过 `embed` 服务 Next 静态产物。

---

## 10. 前端实现约束

### 10.1 Next.js 定位

Next.js 只负责前端页面，不承担核心后端职责。不要把本地仓库扫描、AST 分析、文件系统读取放到 Next API Routes 中。

### 10.2 React Flow

- graph nodes / edges 必须由后端 JSON 转换而来。
- 节点点击后请求 `/api/source`。
- 默认显示中粒度业务调用链。
- UI 必须支持入口 symbol 搜索和 depth 控制。
- UI 必须显示 partial graph warnings。

### 10.3 源码面板

- v1 使用 CodeMirror 或普通 `<pre>` 都可。
- 源码面板应显示：
  - symbol 名称
  - 文件路径
  - 行号范围
  - 源码片段

---

## 11. 推荐阶段顺序

AI agent 应按以下阶段推进，不要跳阶段：

1. `00-contract-and-fixtures.md`
2. `01-cli-and-loader.md`
3. `02-symbol-extraction.md`
4. `03-call-extraction.md`
5. `04-graph-query-and-filtering.md`
6. `05-http-api-and-source.md`
7. `06-web-ui.md`
8. `07-packaging-release-and-demo.md`
9. `08-quality-gates.md`

每个阶段都必须通过对应文档中的验收命令和成功标准后，才进入下一阶段。

---

## 12. 测试策略

每个阶段至少包含：

- Go 单元测试
- fixture 项目测试
- CLI smoke test
- JSON schema 或结构断言
- 错误路径测试

推荐 fixture：

```text
examples/simple/
  main.go
  service.go

examples/layered-service/
  cmd/api/main.go
  internal/handler/user.go
  internal/service/user.go
  internal/repository/user.go

examples/interface-call/
  main.go
  repo.go
  service.go
```

---

## 13. AI Agent 工作规则

1. 优先保持阶段目标小而完整。
2. 修改代码前先检查对应阶段文档。
3. 不要一次性实现所有阶段。
4. 不要引入不必要依赖。
5. 不要为了解决一个小问题重写整个架构。
6. 每次修改后运行对应阶段的验收命令。
7. 任何加载失败、解析失败、未知调用都应进入 warning 或 unresolved edge。
8. 输出 JSON 时保持字段稳定，不要随意重命名。
9. 前端类型必须与后端 graph schema 对齐。
10. 新增 API 时必须更新文档和前端类型。
11. 不要把测试文件纳入默认分析结果。
12. 不要把标准库调用默认显示到业务图里。
13. 不要承诺运行时精确调用链；本项目提供静态可解析调用关系。
14. 如果需要改变 schema，必须同时更新：
    - Go model
    - TypeScript type
    - 示例 JSON
    - 阶段文档
    - 测试

---

## 14. Definition of Done

一个阶段完成必须同时满足：

1. 代码实现完成。
2. 自动化测试通过。
3. CLI smoke test 通过。
4. 示例 fixture 能产生预期输出。
5. 文档中的验收命令能够运行。
6. 没有把 v1 非目标功能强行加入。
7. 用户可通过 README 或阶段文档复现结果。

---

## 15. 推荐最终演示路径

最终 demo 应展示：

```bash
codemap serve ./examples/layered-service
```

浏览器中完成：

1. 搜索 `main.main`。
2. 展示 `main -> handler -> service -> repository` 调用链。
3. 调整 depth。
4. 点击 `UserService.CreateUser`。
5. 底部源码面板展示 `CreateUser` 源码。
6. warnings 面板显示 partial graph 或 unresolved call 信息。
