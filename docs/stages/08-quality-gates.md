# Stage 08 — Quality Gates and Regression Checks

## 目标

建立统一质量门禁，防止后续迭代破坏 analyzer、graph schema、HTTP API 或前端 UI。这个阶段不是新增功能，而是保证项目稳定可维护。

---

## 本阶段需要完成的内容

### 1. Go 测试门禁

必须通过：

```bash
go test ./...
```

建议额外增加：

```bash
go test ./internal/analyzer/... -run TestExtractSymbols
go test ./internal/analyzer/... -run TestExtractCalls
go test ./internal/graph/... -run TestGraphTraversal
go test ./internal/server/... ./internal/source/...
```

### 2. CLI smoke test

建议脚本：

```bash
scripts/smoke.sh
```

包含：

```bash
go run ./cmd/codemap scan ./examples/simple
go run ./cmd/codemap symbols ./examples/simple
go run ./cmd/codemap calls ./examples/simple
go run ./cmd/codemap graph ./examples/layered-service --entry main.main --depth 5
```

通过标准：

- 所有命令 exit code 为 0。
- JSON 可解析。
- graph nodes / edges 非空。

### 3. Frontend quality gate

必须通过：

```bash
cd web
pnpm lint
pnpm typecheck
pnpm build
```

### 4. Schema regression

建议维护固定样例：

```text
docs/examples/graph.layered-service.json
```

新增测试：

- graph 输出字段不能缺失。
- node 字段不能随意改名。
- edge resolution 不能出现未知值。
- TypeScript schema 与 Go schema 保持一致。

### 5. Fixture regression

每个 fixture 都应有明确用途：

| fixture | 用途 |
|---|---|
| `examples/simple` | 普通函数和简单方法调用 |
| `examples/layered-service` | 业务调用链 demo |
| `examples/interface-call` | 接口调用保守解析 |
| `examples/cycle`，可选 | 循环调用保护 |
| `examples/unresolved`，可选 | 函数变量等 unresolved call |

### 6. 性能基线

v1 不需要极致性能，但应避免明显问题：

- 不要每次 source 请求都重新扫描项目。
- 不要每次 graph 请求都完整重新 load packages。
- analyzer 结果应在 serve 生命周期内缓存。
- 大图必须受 depth 和 filter 控制。

### 7. 安全基线

必须检查：

- `/api/source` 不能读取项目根目录外文件。
- API 参数需要基本校验。
- 不执行被分析项目中的代码。
- 不自动运行用户项目的命令。
- 不上传源码到外部服务。

---

## 验收方式

### 总体验收脚本

建议：

```bash
make check
```

等价于：

```bash
go test ./...
./scripts/smoke.sh
cd web && pnpm lint && pnpm typecheck && pnpm build
```

通过标准：

- 所有命令成功。
- 没有 panic。
- 没有无法解析的 JSON。
- 没有 TypeScript 错误。

### API 回归

启动：

```bash
go run ./cmd/codemap serve ./examples/layered-service --port 8080
```

检查：

```bash
curl -s http://localhost:8080/api/health
curl -s http://localhost:8080/api/symbols
curl -s "http://localhost:8080/api/graph?entry=main.main&depth=5"
```

通过标准：

- 三个接口返回合法 JSON。
- graph 中至少包含一个 node 和一个 edge。

### UI 回归

人工检查：

1. 页面正常加载。
2. 搜索可用。
3. 图可见。
4. 点击节点后源码出现。
5. warnings 或 empty state 显示正常。
6. 刷新页面后不崩溃。

---

## 成功完成标准

本阶段完成时，应满足：

1. `go test ./...` 通过。
2. CLI smoke test 通过。
3. `pnpm lint` 通过。
4. `pnpm typecheck` 通过。
5. `pnpm build` 通过。
6. API 回归通过。
7. UI 手动回归通过。
8. README demo 命令仍然可运行。

---

## 常见失败模式

- 只测试 happy path，没有测试 invalid entry、unresolved call、interface call。
- 修改 schema 后没有同步前端类型。
- 前端能 dev 运行，但 build 失败。
- source API 存在 path traversal 风险。
- 大项目图不受 depth 控制。
- 项目 README demo 已经过期。
