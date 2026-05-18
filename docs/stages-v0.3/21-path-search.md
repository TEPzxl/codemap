# Stage 21 — Path Search

## 目标

支持查询两个 symbols 之间的调用路径。这个功能用于解决真实项目中“我想知道 A 如何调用到 B”的问题。

## 功能定义

用户输入：

```text
from: <symbol id or shortcut>
to: <symbol id or shortcut>
max_depth: N
limit: K
```

系统返回从 `from` 到 `to` 的调用路径。

## 允许修改范围

可以修改：

- `internal/graph/*`
- `internal/server/*`
- `internal/cli/*`
- `web/*`
- tests
- README

不要修改：

- analyzer extraction 核心逻辑
- Graph JSON Schema 必填字段

## 推荐 CLI

```bash
codemap path ./repo --from <symbol> --to <symbol> --max-depth 8 --limit 5
```

输出 JSON：

```json
{
  "from": "...",
  "to": "...",
  "paths": [
    {
      "nodes": ["A", "B", "C"],
      "edges": ["edge-1", "edge-2"]
    }
  ],
  "warnings": []
}
```

## 推荐 API

```text
GET /api/path?from=<symbol>&to=<symbol>&max_depth=8&limit=5
```

错误情况：

- unknown `from`。
- unknown `to`。
- invalid max_depth。
- invalid limit。
- path not found。

所有错误返回 JSON。

## UI 行为

新增 `Path search` 面板：

- From symbol 搜索。
- To symbol 搜索。
- Max depth。
- Limit。
- Find path 按钮。

结果行为：

- 如果找到路径，将当前图切换为 path graph 或高亮路径。
- 如果多条路径，显示 path list，可点击不同路径。
- 没找到时显示 empty state，不白屏。

## 算法要求

- 使用 BFS 找 shortest path。
- 可扩展支持 limit 条路径，但 v0.3 可以先返回最短的前 K 条简单路径。
- 必须有 max_depth，防止大图遍历失控。
- 必须有 visited/cycle 保护。
- 默认遵守 graph filter 选项。

## 测试要求

至少覆盖：

- simple fixture 有路径。
- layered-service 从 `main` 到 `UserRepository.Save` 有路径。
- to 不可达时返回 empty paths 或明确 JSON error。
- max_depth 不足时找不到路径。
- 循环图不会死循环。
- unknown symbol 返回错误。
- CLI 和 API 行为一致。

## 验收命令

```bash
make check
make web-build
make build
```

手动验收：

```bash
./bin/codemap path ./examples/layered-service --from main.main --to UserRepository.Save --max-depth 8
```

Web 中查找 `main` 到 `UserRepository.Save` 的路径，并确认图中高亮或只显示路径。

## 成功标准

- 能回答 A 到 B 是否存在调用路径。
- CLI/API/UI 都可用。
- 查询失败有清晰错误或 empty state。
- 大图中不会无限遍历。

## 常见失败模式

- 只按 display label 匹配导致冲突。
- 忽略 filters，返回用户当前隐藏的边。
- max_depth 无效。
- 多路径场景重复路径过多。
