# Stage 24 — Large Graph UI Polish

## 目标

改善真实项目中大图的阅读体验。该阶段不新增静态分析能力，主要优化前端交互。

## 需要解决的问题

真实项目从 `main` 展开 depth=5 后，节点和边可能很多，用户需要：

- 在当前图中搜索节点。
- 快速跳转到节点。
- 高亮上下游邻居。
- 折叠或扩大左侧 sidebar。
- 更清楚地区分 function/method/external/unresolved/interface。
- 更容易看 selected node/edge 的详情。

## 允许修改范围

可以修改：

- `web/*`
- 前端类型和组件
- README 中 UI 说明

不要修改：

- Go analyzer
- server API，除非只是补充非破坏性字段
- Graph JSON Schema 必填字段

## 需要完成

### 1. 当前图节点搜索

新增 `Search current graph` 输入框：

- 按 label、id、package、file 搜索。
- 搜索结果列表显示匹配节点。
- 点击结果：
  - 选中节点。
  - 将 React Flow viewport 移动到该节点。
  - 高亮该节点。

### 2. 邻居高亮

点击节点后：

- 当前节点高亮。
- 直接上游/下游节点高亮。
- 相关边高亮。
- 非相关节点可轻微淡化。

### 3. Sidebar collapse

允许折叠左侧控制面板，只保留一个展开按钮，给大图更多空间。

### 4. 更清晰的节点/边样式

根据类型区分：

- function
- method
- external
- unresolved
- interface candidate

注意：不要使用过多颜色，保持简洁。

### 5. Empty/loading/error 状态统一

- graph loading。
- source loading。
- callsite loading。
- path search loading。
- API error。

## 验收命令

```bash
make check
make web-build
make build
```

手动验收：

```bash
./bin/codemap serve ./examples/layered-service --port 8080
```

在真实项目大图中测试：

- 搜索节点。
- 跳转节点。
- 折叠 sidebar。
- 点击节点看高亮。
- reset layout。

## 成功标准

- 大图中可以快速找到节点。
- 选中节点后上下游关系更清楚。
- sidebar 可折叠。
- 前端 build/lint/typecheck 通过。

## 常见失败模式

- 搜索只搜 label，不搜完整 id。
- 节点跳转后没有选中。
- 高亮逻辑依赖 display label，导致冲突。
- 样式过度复杂，图反而更难读。
