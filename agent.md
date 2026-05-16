# Agent Guide

本文档面向维护者和自动化代理，说明 ClipLite 仓库的结构、常用命令和变更注意事项。

## 项目概览

ClipLite 是一个 Windows 剪贴板管理工具，技术栈为 Go + Wails v2 + 原生 HTML/CSS/JavaScript。

仓库根目录主要用于文档和发布规划。实际 Wails 应用位于：

```text
ClipLite/
```

## 关键路径

```text
README.md                         # 面向用户的主说明
TODOLIST.md                       # 开源发布与分发计划
docs/cross-compile-windows.md     # Windows 交叉编译说明
ClipLite/wails.json               # Wails 配置
ClipLite/go.mod                   # Go 模块依赖
ClipLite/main.go                  # Wails 应用入口
ClipLite/app.go                   # 应用核心逻辑和绑定方法
ClipLite/clipboard_monitor.go     # 剪贴板轮询逻辑
ClipLite/clipboard_windows.go     # Windows 剪贴板 API
ClipLite/tray_windows.go          # Windows 托盘和键盘 hook
ClipLite/hotkey_windows.go        # 快捷键相关逻辑
ClipLite/frontend/dist/           # 当前前端资源
ClipLite/build/appicon.png        # Wails 应用图标
ClipLite/build/bin/icon.ico       # 运行目录图标
ClipLite/build/windows/icon.ico   # Windows 构建图标
```

## 当前依赖

以 `ClipLite/go.mod` 为准：

```text
Go 1.22.0
github.com/wailsapp/wails/v2 v2.12.0
modernc.org/sqlite v1.35.0
```

## 常用命令

所有 Go/Wails 命令应在 `ClipLite/` 子目录执行。

下载 Go 依赖：

```bash
cd ClipLite
go mod download
```

安装前端依赖：

```bash
cd ClipLite/frontend
npm install
```

开发模式：

```bash
cd ClipLite
wails dev
```

Windows 本机构建：

```bash
cd ClipLite
wails build -o ClipLite.exe
```

Linux/macOS 交叉编译 Windows amd64：

```bash
cd ClipLite
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
export CGO_ENABLED=1
wails build -platform windows/amd64 -o ClipLite.exe
```

Go 测试：

```bash
cd ClipLite
go test ./...
```

## 前端注意事项

当前前端直接使用 `ClipLite/frontend/dist/` 中的静态文件。

重要文件：

```text
ClipLite/frontend/dist/index.html
ClipLite/frontend/dist/style.css
ClipLite/frontend/dist/app.js
ClipLite/frontend/dist/wails/ipc.js
ClipLite/frontend/dist/wails/runtime.js
```

`frontend/dist/wails/ipc.js` 是当前构建产物中的 Wails 调用桥接。修改或新增后端绑定方法时，要确认这里也暴露了对应方法，否则前端按钮可能调用 `undefined`。

## 后端绑定方法

Wails 绑定对象是 `App`。常见前端调用方法包括：

```text
GetRecords
GetFavorites
ToggleFavorite
DeleteRecord
ClearAllRecords
GetConfig
UpdateConfig
SelectDirectory
CopyToClipboard
HideWindow
ShowWindow
ExitApp
RecordFromPaste
```

新增绑定方法后，需要同步检查前端桥接和 TypeScript/JS 调用。

## Windows 行为注意事项

ClipLite 涉及以下敏感能力：

- 剪贴板读取和写入。
- 全局快捷键。
- 键盘 hook。
- 系统托盘驻留。

这些能力容易触发安全软件关注。修改相关逻辑时，应保持行为透明，不要增加静默联网、隐藏进程、自启动等高风险行为。

## 图标与资源

图标源稿：

```text
ClipLite/build/icon-source.svg
```

发布和运行使用的图标：

```text
ClipLite/build/appicon.png
ClipLite/build/bin/icon.ico
ClipLite/build/windows/icon.ico
```

更新图标时，需要同步检查 PNG 尺寸和 ICO 多尺寸内容。

## 发布要求

正式 release 应至少包含：

```text
ClipLite.exe
ClipLite.exe.sha256
源码压缩包
构建环境说明
```

推荐后续补充：

- GitHub Actions 自动构建。
- 代码签名。
- winget manifest。
- Scoop manifest。
- SmartScreen 误报反馈流程。

发布路线图见：

```text
TODOLIST.md
```

交叉编译说明见：

```text
docs/cross-compile-windows.md
```

## 变更守则

- 不要回滚用户已有修改，除非用户明确要求。
- 修改 Windows API、托盘、快捷键、剪贴板逻辑后，优先验证真实 Windows 行为。
- 文档中的版本号应与 `ClipLite/go.mod` 和 `ClipLite/wails.json` 保持一致。
- release 相关文档必须写明产物路径和校验方式。
- 不要把本地绝对路径写入用户文档，除非是在故障排查记录中明确说明。
