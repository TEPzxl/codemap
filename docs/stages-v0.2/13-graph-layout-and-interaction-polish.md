# Stage 13 — Graph Layout and Interaction Polish

## 目标

提升 React Flow 图的可读性、布局稳定性和交互体验，使其更适合展示真实 Go 项目的业务调用链。

本阶段主要修改前端，不修改 analyzer 核心逻辑。

## 需要完成

### 自动布局

1. 将调用图按照层级布局展示。
2. 同一 depth 的节点尽量排列在同一层。
3. 尽量减少边交叉。
4. layout 结果应稳定：同一 graph 多次加载位置一致。
5. graph 为空时展示 empty state。

可以自行选择轻量布局算法或手写简单层级布局。

### 交互增强

1. 点击节点：
   - 高亮选中节点。
   - 高亮与该节点直接相关的 edges。
   - 显示 source snippet。
2. 点击空白区域：
   - 可清空 selected，或保持当前 selected；行为要一致。
3. 支持 fit view。
4. 支持 minimap。
5. 支持 controls。
6. 提供 reset layout 按钮。

### 节点视觉区分

至少区分：

- function
- method
- external
- unresolved

可以用不同边框、角标、badge 或背景，但不要过度花哨。

### 边视觉区分

至少区分：

- resolved
- interface
- external
- unresolved

edge label 保留 resolution。

## 不做内容

- 不做 3D 图。
- 不做 force-directed 无限拖拽调优。
- 不引入大型图数据库。
- 不修改后端调用关系算法。

## 验收命令

```bash
make check
make web-build
make build
./bin/codemap serve ./examples/layered-service --port 8080
```

人工验收：

1. `layered-service` 图层级清晰。
2. `simple` 图不显得过度分散。
3. `interface-call` 在开启 interface expansion 后仍可读。
4. 点击节点后相关边高亮。
5. fit view / reset layout 可用。
6. source panel 不被图遮挡。

## 成功标准

- 图比 v0.1 更可读。
- 布局稳定。
- 大图不会明显卡死。
- 不破坏 API 和 Graph JSON Schema。

## 常见失败模式

- layout 每次刷新位置随机。
- 节点重叠。
- edge label 遮挡严重。
- 大量 state 更新导致页面卡顿。
- React Flow node id 与 graph node id 不一致，导致源码点击失败。
