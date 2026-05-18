# Stage 16 — CI, Release, and v0.2 Demo Docs

## 目标

让 `codemap` 更接近可发布项目：CI 可验证、release 构建清晰、demo 文档可复现。

本阶段主要修改：

- `.github/workflows/`
- `Makefile`
- `README.md`
- `docs/`
- `scripts/`

不修改 analyzer 核心逻辑。

## 需要完成

### CI

新增 GitHub Actions workflow，例如：

```text
.github/workflows/ci.yml
```

CI 至少运行：

```bash
make check
make web-build
make build
```

如果 CI 环境中 `corepack` 不稳定，应在 workflow 中显式启用：

```bash
corepack enable
corepack prepare pnpm@10.20.0 --activate
```

版本可与 `web/packageManager` 保持一致。

### Release build

增加：

```bash
make release
```

建议输出：

```text
dist/codemap-linux-amd64
dist/codemap-darwin-arm64
dist/codemap-darwin-amd64
```

Windows 可选。

### Demo docs

补充：

```text
docs/demo/README.md
docs/demo/codemap-ui.png
```

Demo README 应包含：

1. 如何构建。
2. 如何运行 demo。
3. 打开哪个 URL。
4. 如何选择 entry symbol。
5. 如何点击节点查看源码。
6. 如何打开 interface candidate expansion。
7. 当前限制。

### README

README 应更新到 v0.2：

- 新增 v0.2 功能列表。
- 保留 v0.1 命令。
- 加入 interface candidate 说明。
- 加入 filters 说明。
- 加入 rescan/meta 说明，如果 Stage 15 已完成。
- 加入 CI / release 说明。

## 验收命令

```bash
make check
make web-build
make build
make release
./bin/codemap serve ./examples/layered-service --port 8080
curl -s http://localhost:8080/api/health
```

如果已有 GitHub CLI 或 CI 环境，可额外检查 workflow yaml 格式；本地不强制。

## 成功标准

- CI workflow 与本地 `make check` 一致。
- release 构建不依赖手工步骤。
- README 命令可复制执行。
- demo 文档能让陌生用户运行项目。
- `bin/`、`dist/`、`node_modules/` 不被提交。

## 常见失败模式

- CI 中 pnpm 版本与本地不一致。
- release build 忘记先 `make web-build`。
- workflow 里路径错，找不到 `web/`。
- README 描述了未实现功能。
- 把 release binary 提交进 Git。

## Codex 提示词

```text
现在开始 Stage 16：CI, Release, and v0.2 Demo Docs。

只做 CI、release build、README 和 demo 文档。不要修改 analyzer/graph/server 核心逻辑，不要修改 Graph JSON Schema。完成后运行 make check、make web-build、make build、make release，并汇报结果。
```
