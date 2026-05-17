# Stage 04 — Graph Query and Filtering

## 目标

基于 symbols 和 calls 构建调用图，并支持从入口函数按深度展开。这个阶段需要实现中粒度业务调用链展示所需的核心图查询能力。

---

## 本阶段需要完成的内容

### 1. 构建图

建议位置：

```text
internal/graph/builder.go
internal/graph/traversal.go
internal/graph/filter.go
```

输入：

- symbols
- calls
- options

输出：

- Graph JSON

### 2. 支持入口 symbol

命令：

```bash
codemap graph <path> --entry <symbol-or-query> --depth 5
```

要求：

- entry 可以先要求完整 symbol ID。
- 后续可以支持模糊匹配。
- entry 不存在时返回明确错误，不 panic。

### 3. 支持 depth

示例：

```bash
codemap graph ./examples/layered-service --entry main.main --depth 3
```

要求：

- depth 为 0 时只返回入口节点。
- depth 为 1 时返回入口直接调用。
- 默认 depth 建议为 4 或 5。
- depth 必须防止图爆炸。

### 4. 支持过滤

默认策略：

- 不显示测试代码。
- 不显示标准库调用。
- 不显示第三方依赖调用。
- 可以保留 unresolved 作为可选开关。

建议选项：

```bash
--show-external=false
--show-unresolved=true
--depth=5
```

### 5. 循环保护

必须防止递归无限展开。

建议：

- 使用 visited set。
- 对 cycle edge 标记或至少保留一次边。
- 不要求 v1 做复杂 cycle UI。

---

## 验收方式

### layered service graph

```bash
go run ./cmd/codemap graph ./examples/layered-service --entry main.main --depth 5
```

通过标准：

- 输出合法 Graph JSON。
- `nodes` 非空。
- `edges` 非空。
- 图中包含 main、handler、service、repository 中的预期节点。
- 默认不出现大量 `fmt.Println`、`net/http` 等标准库节点。

### depth 检查

```bash
go run ./cmd/codemap graph ./examples/layered-service --entry main.main --depth 0
```

通过标准：

- 只返回 entry 节点。
- 不返回调用边，或仅返回空 edges。

```bash
go run ./cmd/codemap graph ./examples/layered-service --entry main.main --depth 1
```

通过标准：

- 返回 entry 及其直接下游节点。
- 不展开更深层级。

### invalid entry

```bash
go run ./cmd/codemap graph ./examples/layered-service --entry not.exists --depth 5
```

通过标准：

- 返回明确错误信息。
- exit code 非 0。
- 不 panic。
- 不输出破坏 JSON 的混合日志。

### 自动化测试

```bash
go test ./internal/graph/... -run TestGraphTraversal
```

建议断言：

- depth 限制有效。
- visited 去重有效。
- cycle 不导致死循环。
- filtering 能隐藏 external nodes。
- edge 的 from / to 对应图中节点。

---

## 成功完成标准

本阶段完成时，应满足：

1. `codemap graph <path> --entry ... --depth ...` 可用。
2. 能从入口函数展开调用链。
3. depth 参数有效。
4. 默认隐藏标准库和第三方依赖。
5. 图遍历不会因循环调用死循环。
6. 输出 Graph JSON 与 Stage 00 schema 一致。
7. `go test ./internal/graph/...` 通过。

---

## 常见失败模式

- 不限制 depth，导致大项目图爆炸。
- 只返回所有节点和所有边，前端难以展示。
- entry 使用短名导致多 package 冲突。
- 循环调用导致无限递归。
- 过滤 external 后留下悬空 edge。
- CLI 输出中混入调试日志，JSON 无法解析。
