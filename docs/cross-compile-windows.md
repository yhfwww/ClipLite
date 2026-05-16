# ClipLite Windows 交叉编译方法

本文档说明如何在 Linux/macOS 环境中交叉编译 ClipLite 的 Windows 版本。

ClipLite 是基于 Go + Wails v2 的 Windows 剪贴板管理工具。当前仓库实际依赖版本以 `ClipLite/go.mod` 为准：

- Go：`1.22.0`
- Wails：`github.com/wailsapp/wails/v2 v2.12.0`
- SQLite：`modernc.org/sqlite v1.35.0`

## 一、编译目标

生成 Windows 64 位可执行文件：

```text
ClipLite/build/bin/ClipLite.exe
```

相关图标文件：

```text
ClipLite/build/bin/icon.ico
ClipLite/build/appicon.png
```

## 二、环境要求

### 必需组件

- Go 1.22+
- Wails CLI v2
- Node.js 与 npm
- mingw-w64 交叉编译工具链
- Git

### 推荐版本

- Go：1.22.x 或更高 1.x 稳定版
- Wails CLI：v2.12.x
- Node.js：20 LTS 或更高 LTS
- npm：随 Node.js LTS 安装即可

## 三、安装 Go

Linux x86_64 示例：

```bash
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc
```

验证：

```bash
go version
```

期望输出类似：

```text
go version go1.22.0 linux/amd64
```

## 四、安装 Wails CLI

建议安装与项目依赖接近的 Wails v2 CLI：

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
```

验证：

```bash
wails version
```

如果提示找不到 `wails`，确认 `$HOME/go/bin` 已加入 `PATH`。

## 五、安装 Node.js 与 npm

Ubuntu/Debian 可使用 NodeSource 或系统包。推荐 Node.js LTS。

验证：

```bash
node --version
npm --version
```

## 六、安装 Windows 交叉编译工具链

Ubuntu/Debian：

```bash
sudo apt update
sudo apt install -y gcc-mingw-w64-x86-64
```

设置交叉编译器：

```bash
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
export CGO_ENABLED=1
```

验证：

```bash
x86_64-w64-mingw32-gcc --version
```

## 七、获取源码

```bash
git clone <repository-url> ClipLite
cd ClipLite/ClipLite
```

如果已经在仓库根目录：

```bash
cd ClipLite
```

注意：项目实际 Go/Wails 工程目录是仓库中的 `ClipLite/` 子目录，不是仓库根目录。

## 八、安装项目依赖

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

说明：当前 `wails.json` 中配置的是：

```json
"frontend:install": "npm install",
"frontend:build": "npm run build"
```

因此使用 `wails build` 时，Wails 会按配置执行前端安装和构建。

## 九、交叉编译 Windows exe

推荐命令：

```bash
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
export CGO_ENABLED=1

wails build -platform windows/amd64 -o ClipLite.exe
```

编译成功后，产物通常位于：

```text
build/bin/ClipLite.exe
```

也就是仓库路径：

```text
ClipLite/build/bin/ClipLite.exe
```

## 十、调试版本编译

如果需要保留调试信息：

```bash
wails build -platform windows/amd64 -o ClipLite.exe -debug
```

调试版本体积更大，不建议作为正式 release 发布。

## 十一、开发模式

本机开发运行：

```bash
wails dev
```

注意：`wails dev` 是开发模式，通常用于当前系统本机运行，不等同于交叉编译 Windows release。

## 十二、校验编译产物

计算 SHA256：

```bash
sha256sum build/bin/ClipLite.exe
```

Windows PowerShell 等价命令：

```powershell
Get-FileHash .\build\bin\ClipLite.exe -Algorithm SHA256
```

建议发布时同时提供：

```text
ClipLite.exe
ClipLite.exe.sha256
```

## 十三、运行时文件说明

编译后的主程序：

```text
build/bin/ClipLite.exe
```

托盘图标：

```text
build/bin/icon.ico
```

Wails/Windows 构建图标：

```text
build/windows/icon.ico
```

首次运行后程序会创建数据目录，包含：

```text
data/cliplite.db
data/config.json
data/log.txt
```

具体数据目录由程序运行位置和权限决定，程序会优先尝试可写路径。

## 十四、常见问题

### 1. 找不到 wails 命令

确认 Go bin 目录在 `PATH` 中：

```bash
export PATH=$PATH:$HOME/go/bin
```

然后重新验证：

```bash
wails version
```

### 2. Windows 交叉编译失败，提示 gcc 或 cc 不存在

确认安装了 mingw-w64，并设置了：

```bash
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
export CGO_ENABLED=1
```

### 3. Go 依赖下载失败

如果网络无法访问 `proxy.golang.org`，可以配置 Go 代理。

中国大陆网络常用示例：

```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

再执行：

```bash
go mod download
```

### 4. 前端依赖安装失败

进入前端目录单独安装：

```bash
cd frontend
npm install
cd ..
```

然后重新执行：

```bash
wails build -platform windows/amd64 -o ClipLite.exe
```

### 5. SmartScreen 提示发布者未知

这是 Windows 对未签名或信誉不足二进制文件的提示，不是交叉编译步骤本身导致的。

正式发布建议：

- 使用固定 release 渠道。
- 为 `ClipLite.exe` 做代码签名。
- 发布 SHA256 校验值。
- 提供源码与可复现构建说明。

## 十五、完整命令示例

以下示例假设在 Ubuntu/Debian 上从零开始交叉编译 Windows amd64 版本：

```bash
# 1. 安装系统依赖
sudo apt update
sudo apt install -y git gcc-mingw-w64-x86-64

# 2. 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
export PATH=$PATH:$HOME/go/bin

# 3. 获取源码
git clone <repository-url> ClipLite
cd ClipLite/ClipLite

# 4. 下载依赖
go mod download
cd frontend
npm install
cd ..

# 5. 设置交叉编译环境
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
export CGO_ENABLED=1

# 6. 编译 Windows exe
wails build -platform windows/amd64 -o ClipLite.exe

# 7. 校验产物
sha256sum build/bin/ClipLite.exe > build/bin/ClipLite.exe.sha256
```

最终产物：

```text
build/bin/ClipLite.exe
build/bin/ClipLite.exe.sha256
```
