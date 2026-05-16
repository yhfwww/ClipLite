# ClipLite Application

本目录是 ClipLite 的 Wails 应用工程目录。

面向用户的完整说明请查看仓库根目录：

```text
../README.md
```

Windows 交叉编译说明请查看：

```text
../docs/cross-compile-windows.md
```

## 技术栈

- Go 1.22.0
- Wails v2.12.0
- SQLite via `modernc.org/sqlite`
- 原生 HTML/CSS/JavaScript 前端

## 常用命令

下载 Go 依赖：

```bash
go mod download
```

安装前端依赖：

```bash
cd frontend
npm install
cd ..
```

开发模式：

```bash
wails dev
```

Windows 本机构建：

```bash
wails build -o ClipLite.exe
```

Linux/macOS 交叉编译 Windows amd64：

```bash
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
export CGO_ENABLED=1
wails build -platform windows/amd64 -o ClipLite.exe
```

构建产物通常位于：

```text
build/bin/ClipLite.exe
```

## 关键文件

```text
wails.json                 # Wails 构建配置
go.mod                     # Go 依赖
main.go                    # 应用入口
app.go                     # 核心业务逻辑与 Wails 绑定
clipboard_monitor.go       # 剪贴板轮询
clipboard_windows.go       # Windows 剪贴板 API
tray_windows.go            # 系统托盘和键盘 hook
frontend/dist/             # 前端静态资源
build/appicon.png          # 应用图标
build/bin/icon.ico         # 运行目录图标
build/windows/icon.ico     # Windows 构建图标
```

## 注意事项

- 修改后端绑定方法时，同步检查 `frontend/dist/wails/ipc.js`。
- 修改剪贴板、全局快捷键、键盘 hook 或托盘逻辑后，应在 Windows 上验证真实行为。
- 正式发布前应生成 SHA256 校验文件，并尽量使用代码签名。
