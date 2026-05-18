# codemap v0.3 阶段索引

## 版本主题

`v0.3` 聚焦真实项目的大图探索体验。`v0.2` 已经证明 `codemap` 可以处理真实 Go 后端和 pure Go 项目，但从 `main` 以 depth=5 展开时图会迅速变大。

`v0.3` 不优先继续增加静态分析能力，而是优先解决：

- 大图如何聚焦。
- 如何从某个节点看上游/下游。
- 如何查找两个 symbol 之间的调用路径。
- 如何先看 package 级别结构，再进入函数级别。
- 如何自动发现合理入口点。
- 如何导出当前视图用于 README、文档和面试展示。

## 阶段顺序

| Stage | 文档 | 目标 |
|---:|---|---|
| 18 | `18-v0.3-baseline-and-real-project-demos.md` | 固定 v0.2 基线，加入真实项目 demo 文档 |
| 19 | `19-graph-summary-and-view-state.md` | 当前 graph summary panel 与 URL view state |
| 20 | `20-focus-mode.md` | 从选中节点聚焦上游、下游、邻域 |
| 21 | `21-path-search.md` | 查询两个 symbols 之间的调用路径 |
| 22 | `22-package-overview-and-collapse.md` | package-level graph 与 package navigation |
| 23 | `23-entrypoint-discovery.md` | main/exported/heuristic entrypoint discovery |
| 24 | `24-large-graph-ui-polish.md` | 大图 UI 交互：节点搜索、跳转、高亮、侧栏 |
| 25 | `25-export-and-share.md` | 导出 JSON/Mermaid/DOT 与分享当前视图 |
| 26 | `26-v0.3-quality-gate-and-release.md` | v0.3 质量门禁与 release |

## v0.3 兼容性约束

- 不破坏 v0.2 的 CLI 命令。
- 不破坏 v0.2 的 `/api/graph` 默认行为。
- 不破坏 Stage 00 Graph JSON Schema；新增字段必须是可选字段。
- 默认仍隐藏 external、unresolved、interface candidates。
- interface expansion 仍是静态候选，不宣称运行时真实分派。
- UI 可以显示短 label 和相对路径，但请求、查找、copy full id、source lookup 必须使用完整 id。
