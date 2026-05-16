# ClipLite

轻量级 Windows 剪贴板管理工具

## 功能特性

- 📋 自动记录剪贴板历史（纯文本）
- ⭐ 收藏重要内容，永久保存
- 🔍 关键词快速搜索
- ⌨️ 全局快捷键唤醒（默认 `Ctrl+Shift+V`）
- 🗂️ 自定义保存天数（默认 30 天）
- 💾 自定义数据存储位置
- 🚀 轻量级，单文件便携版（约 9MB）
- 🪟 系统托盘驻留

## 系统要求

- Windows 10 / Windows 11
- WebView2 运行时（Windows 11 已内置，Windows 10 可自动下载或[手动安装](https://developer.microsoft.com/zh-cn/microsoft-edge/webview2/)）

## 下载使用

### 方法一：直接下载（推荐）

1. 从 [Releases](../../releases) 页面下载最新版本的 `ClipLite.exe`
2. 双击运行即可

### 方法二：自行编译

#### 环境要求

- Go 1.21 或更高版本
- GCC 编译器（如 MinGW-w64）
- Wails CLI（可选，用于开发）

#### 编译步骤

**Windows 环境：**

```batch
# 克隆项目
git clone <repository-url>
cd ClipLite

# 安装依赖
go mod tidy

# 编译
go build -ldflags="-H windowsgui -s -w" -o ClipLite.exe .
```

**Linux/macOS 交叉编译 Windows 版本：**

```bash
# 安装 GCC（Linux）
sudo apt install gcc-mingw-w64

# 克隆项目
git clone <repository-url>
cd ClipLite

# 安装依赖
go mod tidy

# 交叉编译
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -ldflags="-H windowsgui -s -w" -o ClipLite.exe .
```

**使用 Makefile（Linux/macOS）：**

```bash
make build
```

## 使用说明

### 首次运行

1. 双击 `ClipLite.exe` 启动程序
2. 程序会自动最小化到系统托盘
3. 托盘图标显示即表示运行成功

### 快捷键

- `Ctrl+Shift+V` 唤醒主面板（可在设置中自定义）
- `Escape` 隐藏主面板

### 托盘菜单

右键点击托盘图标：
- **显示主面板** - 打开历史记录面板
- **收藏列表** - 快速查看收藏内容
- **设置** - 打开设置面板
- **退出** - 退出程序

### 主面板操作

- **复制内容**：点击历史记录项即可复制到剪贴板
- **收藏**：点击"收藏"按钮，收藏内容永久保存
- **删除**：点击"删除"按钮删除单条记录
- **清空**：点击"清空全部"按钮清空非收藏记录
- **搜索**：在搜索框输入关键词实时过滤

### 设置说明

| 设置项 | 说明 | 默认值 |
|--------|------|--------|
| 快捷键 | 唤醒主面板的快捷键 | Ctrl+Shift+V |
| 保存天数 | 非收藏记录的保存天数 | 30 天 |
| 存储位置 | 数据库和配置文件的存储路径 | 程序同目录 |

## 数据存储

数据文件位于设置指定的存储目录：

```
<存储位置>/
├── cliplite.db      # SQLite 数据库（历史记录）
└── config.json      # 配置文件（设置）
```

## 常见问题

### Q: 程序启动后没有反应？

A: 程序默认最小化到托盘。请查看右下角系统托盘是否有 ClipLite 图标。

### Q: 快捷键不生效？

A: 
1. 检查快捷键是否与其他软件冲突
2. 在设置中更换快捷键组合
3. 确保以管理员权限运行（部分系统需要）

### Q: 如何完全卸载？

A: 删除 `ClipLite.exe` 和数据目录（如使用默认位置，还需删除同目录下的 `data` 文件夹）。

### Q: 数据保存在哪里？

A: 默认保存在程序同目录的 `data` 文件夹中，可在设置中自定义存储位置。

## 开发指南

### 项目结构

```
ClipLite/
├── main.go                 # 应用入口
├── app.go                  # 核心业务逻辑
├── clipboard_monitor.go    # 剪贴板监听
├── clipboard_windows.go    # Windows 剪贴板 API
├── tray_windows.go         # 系统托盘实现
├── frontend/dist/          # 前端资源
│   ├── index.html
│   ├── style.css
│   └── app.js
└── go.mod                  # Go 模块
```

### 技术栈

- **后端**：Go + Wails v2
- **前端**：原生 HTML/CSS/JavaScript
- **数据库**：SQLite
- **窗口**：Frameless 窗口 + 自定义标题栏

### 运行开发版本

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 开发模式运行
wails dev
```

## 许可证

MIT License

## 更新日志

### v1.0.0
- 初始版本
- 支持纯文本剪贴板记录
- 支持收藏功能
- 支持关键词搜索
- 支持自定义快捷键
- 支持自定义保存天数
- 支持自定义存储位置
