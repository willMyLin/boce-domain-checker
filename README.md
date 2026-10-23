# BOCE 域名检测工具 - DNS 污染检测 / QQ 微信拦截 / ICP 备案查询 / 被墙检测

一个基于 Go + Wails + Vue 构建的 BOCE 桌面检测工具，用于批量检测域名状态。API Key 可在 [BOCE 官网](https://www.boce.com) 申请。

## 功能

- DNS 污染检测
- QQ 拦截检测
- 微信拦截检测
- ICP 备案查询
- 备案黑名单检测
- 被墙检测
- TXT 批量导入检测目标
- 按状态筛选检测结果
- 导出当前列表为 Excel
- 检查新版本（启动时静默检查，也可从「帮助 → 检查更新」手动查）

## 界面预览

![BOCE 检测工具界面](https://github.com/user-attachments/assets/7be5ad3c-58b2-4cbd-9ff4-77ae023302a7)

[查看原图](https://github.com/user-attachments/assets/7be5ad3c-58b2-4cbd-9ff4-77ae023302a7)

## 接口说明

当前 BOCE 请求会统一携带以下参数：

```text
key=申请的 API Key
host=检测目标，多个域名用英文逗号分隔
```

已接入的 BOCE API：

```text
/v3/task/create/pollute
/v3/task/create/qq
/v3/task/create/wechat
/v3/task/create/icp
/v3/task/create/blacklist
/v3/task/create/wall
```

## 开发环境

需要先安装：

- Go 1.26+
- Node.js >= 22.12.0（版本见根目录 `.nvmrc`）
- GNU Make

Wails CLI 不用手动安装，`make deps-go` 会钉死版本装好。Wails 安装文档：

https://wails.io/zh-Hans/docs/gettingstarted/installation

安装依赖（Go 依赖、Wails CLI、前端依赖）：

```bash
make deps-go
make deps-fe
```

回到项目根目录运行开发模式（热更新）：

```bash
make dev
```

其它常用入口：

```bash
make help    # 列出全部 target
make fmt     # gofmt 格式化
make clean   # 清理构建产物
```

## 测试

跑 PR 门禁一条命令（`go vet` + 后端测试 + `golangci-lint` + 漏洞扫描）：

```bash
make check
```

只跑后端测试：

```bash
make test
```

只跑静态检查：

```bash
make lint
```

漏洞扫描（Go `govulncheck` + `npm audit`）：

```bash
make audit
```

构建前端（产出 `frontend/dist`；`go build` / `go test` 依赖它，Makefile 已把这层依赖显式化）：

```bash
make frontend
```

## 构建

按平台构建（**一平台一次**，不要一次指定多个平台）：

```bash
make build-darwin-arm64   # macOS Apple Silicon（M 系列）→ build/bin/Boce.app
make build-darwin-amd64   # macOS Intel            → build/bin/Boce.app
make build-windows-amd64  # Windows 64 位          → build/bin/Boce.exe
make build-all            # 串行构建上面三个
```

> **请走 `make build-*`，不要手敲裸 `wails build`。** 裸命令会漏掉版本注入，
> 打出来的包「关于」里显示「版本 dev」、Finder「显示简介」里版本也是旧的：
> `make build-*` 会先跑 `inject-version`（写 `wails.json` 的 `info.productVersion`），
> 再带上 `-ldflags "-X main.Version=$(VERSION)"`，两条通道一起灌进去。
> 确实要绕开 make 时，完整等价命令是：
>
> ```bash
> wails build -clean -platform darwin/arm64 -m -nosyncgomod -ldflags "-X main.Version=$(git describe --tags --always --dirty)"
> ```
>
> （还要先把 `wails.json` 的 `info.productVersion` 改成同一个值，否则系统里显示的版本仍是旧的。）
>
> **不要**写成 `wails build -platform darwin/amd64,darwin/arm64,windows/amd64`：
> 多 target 会产出带连字符的文件名（如 `Boce-darwin-arm64`），与下面的
> 发布物命名契约对不上；而且 `-clean` 只对第一个平台生效，产物会串台。
>
> **Linux 目前没有构建 target**：`wails build -platform linux/amd64` **仅限 Linux 主机**，
> 从 macOS 执行只打印一行 warning 就静默跳过（exit 0、无产物）。

构建产物默认输出到：

```text
build/bin
```

### 发布物说明

```bash
make release   # 注入版本 → 三平台构建 → 打包 → 校验和
```

`make release` 在 `build/release/` 下产出（文件名是对外契约，请勿改动）：

```text
build/release/
├── boce_tool_app_darwin_arm64.zip   # 内含 Boce.app + fix_mac_quarantine.sh
├── boce_tool_app_darwin_arm64.dmg   # 推荐给 Mac 用户，可直接拖进「应用程序」
├── boce_tool_app_darwin_amd64.zip   # 内含 Boce.app + fix_mac_quarantine.sh
├── boce_tool_app_darwin_amd64.dmg
├── boce_tool_app_windows_amd64.zip  # 内含 Boce.exe
└── SHA256SUMS.txt
```

版本号注入有两条独立通道，都由 `make release` 完成：

- OS 级元数据：`wails.json` 的 `info.productVersion`（`wails build` 没有任何 flag 能覆盖它，
  `release` target 会临时把它改成 `git describe` 的结果，构建结束后自动还原）。
- 进程内字符串：`-ldflags "-X main.Version=..."`，供界面「关于」对话框显示。

已压缩好的发布包可在 GitHub Release 下载：

[https://github.com/willMyLin/boce-domain-checker/releases/latest](https://github.com/willMyLin/boce-domain-checker/releases/latest)

## Mac 使用与排障

下载 DMG 打开后，把 `BOCE 检测工具` 拖进「应用程序」，再从「应用程序」双击启动；
也可以在解压出来的目录里执行：

```bash
open Boce.app
```

> 发布物是 **`.app` 应用包**，不是裸二进制：没有 `chmod +x` 这一步，
> 也不要 `./Boce` 这样直接跑。

如果 Mac 提示「文件已损坏」或「无法打开」，是下载文件被 Gatekeeper 隔离标记导致。
在解压后的文件目录执行：

```bash
xattr -dr com.apple.quarantine Boce.app
```

或直接运行同目录下自带的脚本：

```bash
./fix_mac_quarantine.sh
```

> **DMG 不解决 Gatekeeper 隔离**。DMG 只是容器格式，未签名应用从 DMG 拖出来后仍会被拦，
> 未签名应用仍需此步骤。Apple Developer ID 签名 + 公证已明确延后，落地之前这一步不可省略。

Apple Silicon 与 Intel Mac 的命令相同（两边的应用都叫 `Boce.app`）。

如果 zip / dmg 解压时就提示损坏，请先确认文件大小和校验值是否一致：

```bash
shasum -a 256 boce_tool_app_darwin_arm64.dmg
```

校验值见（构建时写入 `build/release/SHA256SUMS.txt`，随 Release 一起发布）：

```text
SHA256SUMS.txt
```

在发布物所在目录一次校验全部文件（`SHA256SUMS.txt` 是 `<hash>  <文件名>`，两个空格）：

```bash
shasum -a 256 -c SHA256SUMS.txt
```

## 自动更新

本工具**只检查并提示，不自动下载安装**。发现新版本时弹窗给出当前/最新版本和更新说明，
点「前往下载」用系统浏览器打开 GitHub Release 页面自行下载 —— 未签名应用被 macOS
Gatekeeper 拦截的那一步（见上一节）在自动安装里同样躲不掉，所以刻意不做静默替换。

检查入口两个：

- **启动时**后台静默查一次。真有新版本才弹窗；查不到、查失败一律不打扰用户。
- **帮助 → 检查更新**手动查。没有新版也提示「已是最新版本」，失败会告诉你是限速还是网络问题。

取数按下面的顺序降级（`boce_update.go`），目的是**尽量不消耗 GitHub API 的限速额度**：

1. `https://github.com/<owner>/<repo>/releases/latest` —— 不跟随重定向，读 `Location`
   里的 tag。实测响应头没有 rate-limit 头，走 GitHub 网页边缘而非 API 限速池。
2. `https://api.github.com/repos/<owner>/<repo>/releases/latest` —— 主路挂了才走。
   未认证限速 **60 次/小时/IP**，能多拿更新说明和资产列表，但会吃额度。
3. 两条都挂就报错，由调用方决定静默吞掉还是提示。

限速（HTTP 429，或 403 且 `X-RateLimit-Remaining: 0`）会被识别成专门的错误，
提示用户稍后再试，而不是含糊的「检查失败」。检查超时 8 秒，绝不拖慢启动或卡住界面。

版本比较会剥掉 `v` 前缀和 `-dirty` / `-rc1` 之类的后缀，本地未提交改动的构建（`v1.1.1-dirty`）
不会因为在开发机上就弹「有新版本」。预发布版不当「最新版」推给用户。

## 目录结构

```text
.
├── app.go              # Wails 后端逻辑和 BOCE API 调用
├── app_test.go         # 后端测试
├── frontend/           # Vue 前端
├── build/              # 图标、平台配置和构建产物目录
├── main.go             # 应用入口
├── wails.json          # Wails 配置
└── README.md
```
