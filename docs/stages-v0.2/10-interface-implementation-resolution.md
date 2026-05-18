# Stage 10 — Interface Implementation Resolution

## 目标

改进 v0.1 对 interface call 的处理。

v0.1 中，接口 receiver 调用会保守标记为：

```json
{
  "resolution": "interface"
}
```

v0.2 需要在保持保守性的前提下，找到项目内部可能实现该接口的具体方法，并在用户显式开启时展示候选实现。

## 设计原则

1. 不声称得到完整运行时调用结果。
2. 候选实现只能表示“静态可能实现”。
3. 默认行为保持 v0.1：不自动展开 interface candidates。
4. 只有用户开启 `expand_interface` 时才展示候选实现。
5. 不通过字符串猜测目标，优先使用 `go/types`。

## 推荐行为

默认：

```text
UserService.CreateUser -> UserRepository.Save
resolution = interface
```

开启 interface expansion 后：

```text
UserService.CreateUser -> UserRepository.Save
resolution = interface

UserService.CreateUser -> MemoryUserRepository.Save
resolution = interface
```

其中第二条边表示候选实现，不表示运行时一定发生。

## 允许的 Schema 扩展

可以做向后兼容扩展，例如：

```go
type Edge struct {
    ...
    Candidate bool `json:"candidate,omitempty"`
    SourceEdgeID string `json:"source_edge_id,omitempty"`
}
```

或等价的 optional 字段。

不允许：

- 删除 v0.1 字段。
- 修改 `resolution` 既有枚举语义。
- 把所有 interface call 都标记成 `resolved`。

## 需要完成

### 后端分析

1. 建立 interface implementation index。
2. 遍历项目内部 named types。
3. 判断 concrete type 是否实现某个 interface。
4. 同时处理 value receiver 和 pointer receiver：
   - `T` 实现 interface。
   - `*T` 实现 interface。
5. 根据 method name 和 method signature 找到候选方法 symbol。
6. 对 interface edge 保留原始 edge。
7. 在 `expand_interface` 开启时额外输出候选实现 edge。

### CLI

增加或扩展：

```bash
codemap graph <path> --entry <symbol> --depth 5 --expand-interface
```

也可以为 `calls` 提供：

```bash
codemap calls <path> --expand-interface
```

但 graph 优先。

### API

支持：

```text
GET /api/graph?entry=<symbol>&depth=5&expand_interface=true
```

默认 `expand_interface=false`。

### Web

本阶段可以只做后端和 API，不要求 UI toggle；UI toggle 在 Stage 12 做。

## 测试要求

至少覆盖：

1. `examples/interface-call` 中接口调用默认不展开候选实现。
2. 开启 `expand_interface` 后出现具体实现方法。
3. value receiver 实现 interface。
4. pointer receiver 实现 interface。
5. 多个实现类时全部列出。
6. 签名不匹配的方法不能作为候选实现。
7. 第三方类型不作为项目内部候选实现，除非项目内部 symbol table 中有对应方法。
8. `go test ./...` 通过。

## 验收命令

```bash
make check
go run ./cmd/codemap graph ./examples/interface-call --entry main.main --depth 5
go run ./cmd/codemap graph ./examples/interface-call --entry main.main --depth 5 --expand-interface
./bin/codemap serve ./examples/interface-call --port 8080
curl -s "http://localhost:8080/api/graph?entry=main.main&depth=5&expand_interface=true" | python -m json.tool
```

## 成功标准

- 默认 graph 输出与 v0.1 兼容。
- `--expand-interface` 后能看到候选实现 edge。
- interface candidate edge 可被前端安全展示。
- 不发生无限递归。
- 不把不相关同名方法误判为实现。

## 常见失败模式

- 只靠方法名匹配实现。
- 忽略 pointer receiver。
- 把 interface candidate 错标为 `resolved`。
- 修改 v0.1 默认行为。
- 导致 `examples/layered-service` 的普通 resolved edge 退化。

## Codex 提示词

```text
现在开始 Stage 10：Interface Implementation Resolution。

只实现 interface candidate resolution。默认行为必须保持 v0.1，只有 --expand-interface 或 expand_interface=true 时才展示候选实现。必须使用 go/types 进行接口实现判断，不要靠字符串猜测。完成后运行 make check 和 interface-call 验收命令。
```
