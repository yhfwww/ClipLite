# ClipLite Open Source Release TODO

目标：让 ClipLite 作为完全开源的 Windows 工具发布时，尽量降低 Microsoft Defender SmartScreen 的拦截概率，并让用户能够验证二进制文件可信、源码可审计、构建可复现。

## 阶段 1：开源透明度

- [ ] 确认仓库包含完整源码、构建脚本、前端产物来源说明。
- [ ] 在 `README.md` 明确说明 ClipLite 的敏感能力：
  - 读取剪贴板文本。
  - 写入剪贴板文本。
  - 使用全局快捷键/键盘 hook。
  - 驻留系统托盘。
  - 数据默认仅保存在本机。
- [ ] 增加隐私说明：
  - 不上传剪贴板内容。
  - 不收集遥测。
  - 不静默联网。
  - 数据文件位置和删除方式。
- [ ] 增加安全说明：
  - 为什么需要剪贴板权限。
  - 为什么需要全局快捷键。
  - 如何完全退出程序。
- [x] 增加 `LICENSE` 文件，明确开源协议。
- [ ] 增加 `SECURITY.md`，说明漏洞报告方式。
- [ ] 增加 `CHANGELOG.md`，记录每次 release 的用户可见变化。

验收标准：

- 新用户只看 README 就能理解程序会访问什么、本地保存什么、如何退出和卸载。
- 仓库具备基础开源项目信任文件：`LICENSE`、`SECURITY.md`、`CHANGELOG.md`。

## 阶段 2：可复现构建说明

- [ ] 固定构建环境要求：
  - Go 版本。
  - Wails 版本。
  - Node.js/npm 版本。
  - Windows SDK / MinGW 要求。
- [ ] 增加 `docs/reproducible-build.md`，写清楚从干净环境构建 release 的完整步骤。
- [ ] 明确依赖安装命令：
  - `go mod download`
  - `npm ci`
  - Wails CLI 安装方式。
- [ ] 明确前端构建命令。
- [ ] 明确 Windows exe 构建命令。
- [ ] 明确 release 产物路径。
- [ ] 记录如何计算 SHA256：
  ```powershell
  Get-FileHash .\ClipLite.exe -Algorithm SHA256
  ```
- [ ] 每个 GitHub Release 附带：
  - `ClipLite.exe`
  - `ClipLite.exe.sha256`
  - 源码压缩包
  - 构建环境版本说明

验收标准：

- 其他开发者能按文档从源码构建出功能一致的 exe。
- 每个 release 都能用 SHA256 校验下载文件是否被篡改。

## 阶段 3：自动化 Release

- [ ] 增加 GitHub Actions 构建工作流。
- [ ] 工作流触发条件：
  - tag push，例如 `v1.0.0`
  - 手动 `workflow_dispatch`
- [ ] 工作流步骤：
  - checkout
  - setup Go
  - setup Node.js
  - 安装依赖
  - 构建前端
  - 构建 Windows exe
  - 计算 SHA256
  - 上传 artifacts
- [ ] 增加 release 草稿生成步骤。
- [ ] 在 release notes 中自动写入：
  - 版本号
  - commit hash
  - SHA256
  - 构建环境
- [ ] 对工作流权限做最小化配置。

验收标准：

- 打 tag 后能自动生成可下载 exe 和 sha256 文件。
- release 页面能追溯产物来自哪个 commit。

## 阶段 4：签名 Release

- [ ] 决定签名方案：
  - 首选：Microsoft Trusted Signing / Azure Artifact Signing。
  - 备选：传统 OV Code Signing 证书。
  - 不建议依赖自签名证书对普通用户分发。
- [ ] 确定发布者名称，后续保持稳定。
- [ ] 在 CI 或本地 release 脚本中加入签名步骤。
- [ ] 使用时间戳服务，避免证书过期后签名失效。
- [ ] 签名后执行验证：
  ```powershell
  signtool verify /pa /v .\ClipLite.exe
  ```
- [ ] release notes 中说明：
  - 发布者名称。
  - 如何查看数字签名。
  - SHA256 校验值。
- [ ] 保持同一证书/发布者持续发布，积累 SmartScreen 信誉。

验收标准：

- Windows 文件属性中能看到可信数字签名。
- SmartScreen 不再显示“发布者未知”。

## 阶段 5：winget 分发

- [ ] 确定 GitHub Release 下载 URL 规则稳定。
- [ ] 准备 winget manifest：
  - PackageIdentifier
  - PackageVersion
  - Publisher
  - PackageName
  - License
  - ShortDescription
  - InstallerUrl
  - InstallerSha256
- [ ] 决定安装形式：
  - 如果继续发布单文件 exe，评估 winget portable package。
  - 如果改为安装包，准备 installer。
- [ ] 向 `microsoft/winget-pkgs` 提交 PR。
- [ ] 每个版本发布后同步更新 winget manifest。

验收标准：

- 用户可以通过以下命令安装：
  ```powershell
  winget install <PackageIdentifier>
  ```

## 阶段 6：Scoop 分发

- [ ] 准备 Scoop manifest。
- [ ] 确定 bucket：
  - 先放入自有 bucket。
  - 稳定后考虑提交到主流 bucket。
- [ ] manifest 包含：
  - version
  - url
  - hash
  - bin
  - shortcuts
  - checkver
  - autoupdate
- [ ] 每个 release 后更新 hash。
- [ ] 验证安装、升级、卸载流程。

验收标准：

- 用户可以通过以下命令安装：
  ```powershell
  scoop install cliplite
  ```

## 阶段 7：SmartScreen 与误报处理

- [ ] 为每个 release 保留签名文件、hash、构建日志。
- [ ] 如果 Defender 或 SmartScreen 报恶意，提交 Microsoft Security Intelligence 分析。
- [ ] 在 issue 模板中增加安全软件误报反馈模板。
- [ ] 在 README 中说明正常 SmartScreen 风险提示和真实恶意报毒的区别。
- [ ] 记录每次误报提交编号和处理结果。

验收标准：

- 用户遇到拦截时有明确反馈入口。
- 维护者有固定流程处理误报。

## 推荐执行顺序

1. 完成阶段 1，先把项目透明度补齐。
2. 完成阶段 2，让用户和贡献者可以自行验证构建。
3. 完成阶段 3，减少人工发布错误。
4. 完成阶段 4，解决“发布者未知”。
5. 完成阶段 5 和阶段 6，扩展可信下载渠道。
6. 持续执行阶段 7，积累信誉和处理误报。

## 当前优先级

- P0：README 隐私/权限说明、SECURITY、可复现构建文档。
- P1：GitHub Actions 自动 release、SHA256 文件、签名方案选型。
- P2：winget、Scoop、误报处理模板。
