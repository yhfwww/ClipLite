# ClipLite

ClipLite 是一个轻量级 Windows 剪贴板管理工具，用于记录纯文本剪贴板历史、收藏常用内容，并通过快捷键快速唤起主面板。

项目目标是保持简单、透明、可本地运行。ClipLite 不上传剪贴板内容，不收集遥测数据，默认只把历史记录保存在本机。

## 功能

- 自动记录纯文本剪贴板历史。
- 收藏重要内容，收藏记录长期保留。
- 关键词搜索历史记录。
- 点击“使用”将历史记录复制回剪贴板。
- 默认快捷键 `Ctrl+Shift+V` 唤起主面板。
- `Escape` 隐藏主面板。
- 系统托盘驻留，支持显示、隐藏和退出。
- 可设置普通记录保存天数。
- 可配置数据存储位置。

## 系统要求

- Windows 10 / Windows 11
- Microsoft Edge WebView2 Runtime

Windows 11 通常已内置 WebView2。Windows 10 如果缺少运行时，可从 Microsoft 官方页面安装：

https://developer.microsoft.com/microsoft-edge/webview2/

## 下载与运行

从项目 Releases 页面下载 `ClipLite.exe`，双击运行。

首次运行后，程序会创建数据目录并初始化数据库。主窗口关闭时会退出程序；需要后台驻留时，请使用最小化或托盘菜单隐藏窗口。

## 使用方法

### 主面板

- 历史记录：查看最近的剪贴板文本。
- 收藏：查看已收藏的记录。
- 设置：修改保存天数和存储路径。
- 使用：把对应记录复制到系统剪贴板。
- 收藏：切换记录收藏状态。
- 删除：删除单条记录。
- 清空全部：清空非收藏记录。

### 快捷键

默认快捷键：

```text
Ctrl+Shift+V
```

快捷键用于唤起主面板。若快捷键不生效，请检查是否被其他软件占用。

### 托盘

右键托盘图标可打开菜单：

- 显示主面板
- 隐藏窗口
- 退出

## 隐私与安全说明

ClipLite 的核心功能需要访问剪贴板，因此它会读取和写入系统剪贴板文本。

ClipLite 的行为边界：

- 不上传剪贴板内容。
- 不收集遥测数据。
- 不静默联网。
- 不隐藏进程。
- 数据默认保存在本机。

本项目使用全局快捷键和系统托盘能力，这些行为可能被安全软件重点关注。开源源码、可复现构建说明、SHA256 校验和代码签名会逐步完善，以降低用户安装和运行时的疑虑。

## 数据存储

程序会优先尝试在可写目录下创建数据文件。常见文件包括：

```text
data/cliplite.db
data/config.json
data/log.txt
```

文件说明：

- `cliplite.db`：SQLite 数据库，保存剪贴板历史。
- `config.json`：用户配置。
- `log.txt`：运行日志。

卸载时，删除 `ClipLite.exe` 和对应数据目录即可。

## 从源码构建

项目的 Wails 工程位于仓库的 `ClipLite/` 子目录。

当前主要依赖版本：

- Go `1.22.0`
- Wails `v2.12.0`
- `modernc.org/sqlite v1.35.0`

### Windows 本机构建

```powershell
cd ClipLite
go mod download
cd frontend
npm install
cd ..
wails build -o ClipLite.exe
```

构建产物通常位于：

```text
ClipLite/build/bin/ClipLite.exe
```

### Linux/macOS 交叉编译 Windows

请查看：

[docs/cross-compile-windows.md](docs/cross-compile-windows.md)

核心命令：

```bash
cd ClipLite
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
export CGO_ENABLED=1
wails build -platform windows/amd64 -o ClipLite.exe
```

## 开发模式

```bash
cd ClipLite
wails dev
```

前端资源位于：

```text
ClipLite/frontend/dist/
```

后端入口和主要逻辑：

```text
ClipLite/main.go
ClipLite/app.go
ClipLite/clipboard_monitor.go
ClipLite/clipboard_windows.go
ClipLite/tray_windows.go
```

## 发布计划

为了让开源发布更可信，本项目计划逐步完善：

- 可复现构建文档。
- GitHub Actions 自动构建。
- Release SHA256 校验文件。
- 代码签名 release。
- winget / Scoop 分发。
- SmartScreen 误报处理流程。

详细计划见：

[TODOLIST.md](TODOLIST.md)

## SmartScreen 提示

如果 Windows Defender SmartScreen 提示“发布者未知”，通常是因为 exe 未进行可信代码签名，或新发布文件尚未积累 SmartScreen 信誉。

这不是开源项目独有的问题。后续正式发布应使用固定发布渠道、SHA256 校验、代码签名和稳定发布者身份。

## 许可证

本项目使用 MIT License。详见 [LICENSE](LICENSE)。
