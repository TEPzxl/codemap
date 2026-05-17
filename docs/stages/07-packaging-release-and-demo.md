# Stage 07 — Packaging, Release, and Demo

## 目标

将 `codemap` 产品化为可演示的简历项目。用户应能通过一个 Go 二进制启动本地服务，浏览器查看调用图和源码。项目需要 README、截图或 GIF、示例项目和发布命令。

---

## 本阶段需要完成的内容

### 1. 前端静态构建

Next 项目需要支持静态输出，目标是让 Go 后端服务构建后的静态资源。

建议：

```bash
cd web
pnpm build
```

如果采用 static export，确保输出目录稳定，例如：

```text
web/out/
```

### 2. Go embed 静态资源

建议位置：

```text
internal/server/static.go
```

目标：

```go
//go:embed web/out/*
var webAssets embed.FS
```

实际路径根据仓库结构调整。

### 3. 单二进制启动

用户运行：

```bash
codemap serve ./examples/layered-service
```

浏览器打开：

```text
http://localhost:8080
```

应看到完整 UI。

### 4. README

README 必须包含：

- 项目一句话。
- 功能列表。
- 技术栈。
- 快速开始。
- CLI 示例。
- Web UI 示例。
- 架构图。
- 限制说明。
- 截图或 GIF。
- 简历描述。

### 5. Demo GIF 或截图

建议展示：

1. 启动 `codemap serve`。
2. 搜索 `main.main`。
3. 展示调用链图。
4. 点击 `UserService.CreateUser`。
5. 底部源码面板出现。

### 6. Release 命令

建议提供：

```bash
make build
make test
make web-build
make demo
```

或等价脚本。

---

## 验收方式

### 构建 Go 二进制

```bash
go build -o ./bin/codemap ./cmd/codemap
```

通过标准：

- 构建成功。
- 产物位于 `bin/codemap`。

### 构建前端

```bash
cd web
pnpm build
```

通过标准：

- 构建成功。
- 输出目录存在。
- 没有 TypeScript 错误。

### 单二进制启动

```bash
./bin/codemap serve ./examples/layered-service --port 8080
```

通过标准：

- 打开 `http://localhost:8080` 可以看到页面。
- 页面可请求 API。
- 图可渲染。
- 源码面板可用。

### README 验证

人工检查 README：

- 新用户可按 README 跑通 demo。
- README 没有声称支持 v1 非目标功能。
- 限制说明清楚，例如接口调用和动态调用不保证运行时精确。

---

## 成功完成标准

本阶段完成时，应满足：

1. Go 二进制可构建。
2. 前端可构建。
3. Go server 可服务前端静态资源。
4. `codemap serve` 可完成端到端 demo。
5. README 完整。
6. 至少有一张截图或一个 GIF。
7. 项目适合放入简历和 GitHub portfolio。

---

## 常见失败模式

- 发布版仍要求用户单独启动 Next dev server。
- README 只讲概念，没有可执行命令。
- demo 使用过于复杂的大型项目，导致图混乱。
- 忽略项目限制，过度承诺“完整运行时调用链”。
- 静态资源路径和 Go embed 路径不匹配。
