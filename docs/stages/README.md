# codemap 阶段文档索引

本目录提供适合 AI coding agent 按阶段执行的项目文档。建议严格按编号推进。

| 阶段 | 文档 | 目标 |
|---:|---|---|
| 00 | `00-contract-and-fixtures.md` | 固定 Graph JSON Schema、fixture 项目和基础协议 |
| 01 | `01-cli-and-loader.md` | 建立 Go CLI 和 package loader |
| 02 | `02-symbol-extraction.md` | 提取函数与方法 symbol |
| 03 | `03-call-extraction.md` | 提取函数调用与方法调用关系 |
| 04 | `04-graph-query-and-filtering.md` | 构建调用图、支持入口展开和过滤 |
| 05 | `05-http-api-and-source.md` | 提供本地 HTTP API 和源码片段读取 |
| 06 | `06-web-ui.md` | 使用 Next + React Flow 展示调用图 |
| 07 | `07-packaging-release-and-demo.md` | 打包成单二进制并完成演示材料 |
| 08 | `08-quality-gates.md` | 统一测试、质量门禁和回归检查 |

建议执行方式：

```bash
# 每完成一个阶段后
go test ./...
go run ./cmd/codemap --help
```

前端阶段开始后：

```bash
cd web
pnpm lint
pnpm typecheck
pnpm build
```
