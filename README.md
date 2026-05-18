# codemap

`codemap` 是一个本地优先的 Go 静态分析与调用图浏览工具。它扫描本地 Go module 或 workspace，提取静态可解析的函数与方法调用关系，并提供一个基于 React Flow 的交互式 Web UI，用于查看调用链、源码片段和调用点。

## v0.2 功能

- 使用 `go/packages` 扫描本地 Go module。
- 提取函数和方法 symbol，并生成稳定 ID。
- 提取 resolved、interface、external 和 unresolved 调用。
- 从入口 symbol 构建带 depth 限制的调用图。
- 通过 `--expand-interface` / `expand_interface=true` 展开接口实现候选。
- CLI、HTTP API 和本地 Web UI 使用一致的图过滤选项。
- Web 支持 symbol 搜索、package filter、depth 控制和 graph filter toggles。
- 通过 `/api/source` 查看节点源码。
- 通过 `/api/callsite` 查看边对应的调用点源码。
- 通过 `/api/meta` 查看项目元信息。
- 通过 `/api/rescan` 手动刷新缓存索引。
- Go server binary 内嵌静态 Web 资源。

默认行为保持保守：默认隐藏标准库和第三方调用，默认隐藏 unresolved 调用，并且只有显式启用时才展示接口实现候选。

## 技术栈

- Go、`go/packages`、`go/ast`、`go/types`、`net/http`、`embed`
- Next.js、React、TypeScript、Tailwind CSS、React Flow
- 通过 Corepack 使用 pnpm 管理前端依赖

## 安装

前置要求：

- Go 1.25 或更新版本
- 带 Corepack 的 Node.js

安装前端依赖：

```bash
make install
```

## 构建与测试

运行完整质量门禁：

```bash
make check
```

`make check` 会运行 Go 测试、fixture 项目的 CLI smoke check、v0.1 golden output 校验、前端 lint、TypeScript 检查和 Next 生产构建。

构建静态 Web UI 并写入 Go embed 使用的目录：

```bash
make web-build
```

构建本地 Go binary：

```bash
make build
```

binary 会输出到：

```text
bin/codemap
```

`go build` 不会运行 pnpm。如果希望 binary 包含当前前端产物，请先运行 `make web-build`，再运行 `make build`。

构建 release binaries：

```bash
make release
```

release 产物会输出到：

```text
dist/codemap-linux-amd64
dist/codemap-darwin-arm64
dist/codemap-darwin-amd64
```

## 快速 Demo

运行 layered-service demo：

```bash
make web-build
make build
./bin/codemap serve ./examples/layered-service --port 8080
```

打开：

```text
http://localhost:8080
```

在 UI 中：

1. 搜索或选择 `main.main`。
2. 点击 `Load graph`。
3. 查看 `main -> handler -> service -> repository` 调用链。
4. 点击 `UserService.CreateUser` 查看节点源码。
5. 点击一条边查看调用点所在行。
6. 查看 `Current graph` 摘要，确认 nodes、edges、packages 和 resolution 分布。
7. 选中任意节点后使用 `Focus downstream`、`Focus upstream` 或 `Focus neighborhood` 聚焦子图。
8. 使用 `Reset to entry` 回到原始入口图。
9. 使用 `Path search` 查询两个 symbols 之间的调用路径，结果会切换为 path graph。
10. 切换到 `Package graph` 查看 package-level call overview，并点击 package 节点设置 package filter。
11. 双击 package 节点切回 function graph 并按该 package 过滤。
12. 切换 filter 或 depth 来刷新图，当前 entry、depth、direction、filter 和 package 会同步到 URL。
13. 点击 `Copy view URL` 复制当前视图链接。
14. 修改本地源码后点击 `Rescan` 刷新索引。

更多说明见：[docs/demo/README.md](docs/demo/README.md)。

![codemap UI screenshot](docs/demo/codemap-ui.png)

## Real Project Demo

v0.3 的主题是 **Focused Graph Exploration**。`codemap` 已用两个真实 Go 项目建立 baseline：

- `contentflow`：28 packages、345 symbols、1572 calls、0 warnings。
- `raft-kv-extended`：7 packages、161 symbols、753 calls、0 warnings。

真实项目从 `main` 入口展开时会快速形成更密的调用图，因此 v0.3 会围绕聚焦入口、路径和 package 的大图探索体验继续推进。详细记录见：[docs/demo/real-projects.md](docs/demo/real-projects.md)。

## CLI 示例

扫描 packages：

```bash
go run ./cmd/codemap scan ./examples/simple
```

列出 symbols：

```bash
go run ./cmd/codemap symbols ./examples/layered-service
```

列出 calls：

```bash
go run ./cmd/codemap calls ./examples/layered-service
```

构建调用图：

```bash
go run ./cmd/codemap graph ./examples/layered-service --entry main.main --depth 5
```

从选中节点查看上游调用：

```bash
go run ./cmd/codemap graph ./examples/layered-service --entry 'github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser' --depth 2 --direction upstream
```

查看上下游邻域：

```bash
go run ./cmd/codemap graph ./examples/layered-service --entry 'github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser' --depth 1 --direction both
```

查找两个 symbols 之间的调用路径：

```bash
go run ./cmd/codemap path ./examples/layered-service --from main.main --to UserRepository.Save --max-depth 8 --limit 5
```

查看 package-level call overview：

```bash
go run ./cmd/codemap packages ./examples/layered-service
go run ./cmd/codemap packages ./examples/layered-service --entry main.main --depth 5
```

显示 external calls：

```bash
go run ./cmd/codemap graph ./examples/layered-service --entry main.main --depth 5 --show-external
```

显示 unresolved calls：

```bash
go run ./cmd/codemap graph ./examples/layered-service --entry main.main --depth 5 --show-unresolved
```

显示 interface calls，但不展开候选实现：

```bash
go run ./cmd/codemap graph ./examples/interface-call --entry main.main --depth 5 --show-interface
```

展开接口实现候选：

```bash
go run ./cmd/codemap graph ./examples/interface-call --entry main.main --depth 5 --expand-interface
```

按 package 或 node limit 过滤：

```bash
go run ./cmd/codemap graph ./examples/layered-service --entry main.main --depth 5 --package github.com/tepzxl/codemap/examples/layered-service/internal/service
go run ./cmd/codemap graph ./examples/layered-service --entry main.main --depth 5 --node-limit 100
```

启动 API 和 UI：

```bash
go run ./cmd/codemap serve ./examples/layered-service --port 8080
```

## HTTP API

Health：

```bash
curl -s http://localhost:8080/api/health
```

Metadata：

```bash
curl -s http://localhost:8080/api/meta | python -m json.tool
```

手动 rescan：

```bash
curl -s -X POST http://localhost:8080/api/rescan | python -m json.tool
```

Symbols：

```bash
curl -s http://localhost:8080/api/symbols
```

Graph：

```bash
curl -s "http://localhost:8080/api/graph?entry=main.main&depth=5"
```

Focus graph：

```bash
curl -s "http://localhost:8080/api/graph?entry=github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser&depth=2&direction=upstream"
curl -s "http://localhost:8080/api/graph?entry=github.com/tepzxl/codemap/examples/layered-service/internal/service.(*UserService).CreateUser&depth=1&direction=both"
```

Path search：

```bash
curl -s "http://localhost:8080/api/path?from=main.main&to=UserRepository.Save&max_depth=8&limit=5"
```

Package graph：

```bash
curl -s "http://localhost:8080/api/package-graph"
curl -s "http://localhost:8080/api/package-graph?entry=main.main&depth=5"
```

带过滤条件的 Graph：

```bash
curl -s "http://localhost:8080/api/graph?entry=main.main&depth=5&show_external=true"
curl -s "http://localhost:8080/api/graph?entry=main.main&depth=5&show_unresolved=true"
curl -s "http://localhost:8080/api/graph?entry=main.main&depth=5&show_interface=true"
curl -s "http://localhost:8080/api/graph?entry=main.main&depth=5&expand_interface=true"
curl -s "http://localhost:8080/api/graph?entry=main.main&depth=5&package=github.com/tepzxl/codemap/examples/layered-service/internal/service"
curl -s "http://localhost:8080/api/graph?entry=main.main&depth=5&node_limit=100"
```

节点源码：

```bash
curl -s "http://localhost:8080/api/source?node_id=<symbol-id>"
```

边调用点：

```bash
curl -s "http://localhost:8080/api/callsite?edge_id=<edge-id>&entry=main.main&depth=5"
```

Warnings：

```bash
curl -s http://localhost:8080/api/warnings
```

## CI

GitHub Actions 会运行：

```bash
make check
make web-build
make build
```

workflow 会启用 Corepack，并激活 `web/package.json` 中声明的 pnpm 版本。

## 开发

运行 Go 测试：

```bash
make test-go
```

运行前端检查：

```bash
make test-web
```

运行单独的前端门禁：

```bash
make web-lint
make web-typecheck
make build-web
```

分别启动 Go API 和 Next dev server：

```bash
make dev-api
make dev-web
```

或者同时启动：

```bash
make dev
```

默认开发地址：

```text
Go API: http://localhost:18080
Web UI: http://127.0.0.1:3000
```

按需覆盖端口：

```bash
make dev-web WEB_PORT=3001
make dev-api API_PORT=8080
```

## 架构

```text
Local Go repo
  -> go/packages loader
  -> AST + type info
  -> symbols
  -> calls
  -> graph builder
  -> CLI / HTTP API
  -> Next + React Flow UI
  -> source and callsite APIs
```

前端不会扫描本地文件，也不会使用 Next API routes 承担核心分析职责。Go server 负责 package loading、analysis、graph building、source reading、cached index metadata 和 API responses。

## 当前限制

- 只支持本地 Go 项目。
- 不支持远程 GitHub URL 扫描。
- 默认忽略 `_test.go` 文件。
- interface candidate expansion 是静态保守候选，不等同于运行时真实分派。
- 通过函数变量触发的动态调用可能会标记为 `unresolved`。
- 调用图是静态近似结果，不是运行时精确调用链。
- 标准库和第三方调用默认隐藏，除非显式启用。
- 不包含数据库、编辑器插件或 LLM 解释层。
