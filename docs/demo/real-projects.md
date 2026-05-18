# Real Project Demos

本页记录 `codemap` 从 fixture demo 进入真实 Go 项目分析后的 v0.3 baseline。Stage 18 不新增功能，不改变 Graph JSON Schema，只固定 `v0.2.0` 能力边界，并为 v0.3 的主题建立起点。

## v0.3 Theme

`v0.3` 主题是 **Focused Graph Exploration**。

真实项目扫描已经能产出可用的 packages、symbols、calls 和 warnings 数据，但从 `main` 入口展开时，图会很快变密。v0.3 应聚焦在大图探索体验：让用户更容易从完整静态调用关系中缩小范围、聚焦子图、理解路径，而不是扩大静态分析承诺。

## Baseline

- Baseline tag：`v0.2.0`
- Baseline commit：`f82ba67`
- 记录日期：2026-05-18
- 默认行为：隐藏标准库和第三方调用，隐藏 unresolved 调用，不展开 interface candidates。
- 分析边界：静态可解析调用关系，不是运行时精确调用链。

## Summary

| Project | Packages | Symbols | Calls | Warnings | Primary entry |
| --- | ---: | ---: | ---: | ---: | --- |
| `contentflow` | 28 | 345 | 1572 | 0 | `github.com/tepzxl/contentflow/cmd/server.main` |
| `raft-kv-extended` | 7 | 161 | 753 | 0 | `raft-kv-extended/cmd/demo.main` |

## contentflow

`contentflow` 是一个包含 HTTP、module、cache、database、observability、scheduler 等包的 Go 服务项目。该项目适合验证 `codemap` 在常见分层后端项目中的默认图效果。

### Scan Result

| Metric | Value |
| --- | ---: |
| Packages | 28 |
| Symbols | 345 |
| Calls | 1572 |
| Warnings | 0 |

### Entries

| Entry | Depth | Nodes | Edges | Warnings | Notes |
| --- | ---: | ---: | ---: | ---: | --- |
| `github.com/tepzxl/contentflow/cmd/server.main` | 5 | 86 | 102 | 0 | 从 server 入口展开，会覆盖配置、日志、数据库、缓存、HTTP router、业务 module 初始化和部分 worker 路径。 |
| `github.com/tepzxl/contentflow/internal/app.Run` | 4 | 85 | 101 | 0 | 与 `main` 接近，更适合观察应用启动逻辑本身。 |

### Commands

```bash
CONTENTFLOW=/path/to/contentflow

go run ./cmd/codemap scan "$CONTENTFLOW"
go run ./cmd/codemap symbols "$CONTENTFLOW"
go run ./cmd/codemap calls "$CONTENTFLOW"
go run ./cmd/codemap graph "$CONTENTFLOW" --entry github.com/tepzxl/contentflow/cmd/server.main --depth 5
go run ./cmd/codemap graph "$CONTENTFLOW" --entry github.com/tepzxl/contentflow/internal/app.Run --depth 4
```

### Observations

- `main` 入口展开后的节点数量已经明显超过 fixture demo，默认视图可用但阅读密度偏高。
- 应用启动链路会同时触达基础设施初始化和多个业务 module，用户需要更好的聚焦能力来只看某个 module 或某条路径。
- 当前 package filter、depth 和 node limit 已经有帮助，但 v0.3 需要进一步降低从大图中定位关键链路的成本。

## raft-kv-extended

`raft-kv-extended` 是一个包含 Raft、KV server、HTTP API、sharding router 和 test utilities 的 Go 项目。该项目适合验证 `codemap` 在分布式系统代码中的调用图密度和热点路径探索。

### Scan Result

| Metric | Value |
| --- | ---: |
| Packages | 7 |
| Symbols | 161 |
| Calls | 753 |
| Warnings | 0 |

### Entries

| Entry | Depth | Nodes | Edges | Warnings | Notes |
| --- | ---: | ---: | ---: | ---: | --- |
| `raft-kv-extended/cmd/demo.main` | 5 | 63 | 85 | 0 | 从 demo 入口展开，会覆盖 KV server、Raft node、memory network、client request 和 apply loop 相关路径。 |
| `raft-kv-extended/internal/kvraft.(*Server).submit` | 4 | 18 | 28 | 0 | 聚焦 KV request 提交到 Raft 的核心路径，图规模更适合人工阅读。 |
| `raft-kv-extended/internal/raft.(*Raft).Start` | 4 | 18 | 29 | 0 | 聚焦 Raft log proposal 入口，适合观察共识模块内部调用。 |

### Commands

```bash
RAFT_KV=/path/to/raft-kv-extended

go run ./cmd/codemap scan "$RAFT_KV"
go run ./cmd/codemap symbols "$RAFT_KV"
go run ./cmd/codemap calls "$RAFT_KV"
go run ./cmd/codemap graph "$RAFT_KV" --entry raft-kv-extended/cmd/demo.main --depth 5
go run ./cmd/codemap graph "$RAFT_KV" --entry 'raft-kv-extended/internal/kvraft.(*Server).submit' --depth 4
go run ./cmd/codemap graph "$RAFT_KV" --entry 'raft-kv-extended/internal/raft.(*Raft).Start' --depth 4
```

### Observations

- `cmd/demo.main` 能展示真实项目级别的端到端链路，但一旦进入 Raft 内部，图会快速连接到较多状态机和日志处理方法。
- 相比从 `main` 入口观察，直接以 `kvraft.(*Server).submit` 或 `raft.(*Raft).Start` 作为入口更符合“聚焦探索”的使用方式。
- 该项目说明 v0.3 不应只优化整体大图布局，还需要支持用户围绕关键 entry、路径和 package 逐步缩小视图。

## Screenshot Status

当前仓库只包含 fixture demo 截图 `docs/demo/codemap-ui.png`。本阶段未新增真实项目截图；如后续补充截图，建议保存为：

- `docs/demo/contentflow-callgraph.png`
- `docs/demo/raft-kv-extended-callgraph.png`
