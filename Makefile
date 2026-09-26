# BOCE 域名检测工具 — 统一开发入口
#
#   make          # 列出全部 target（等价于 make help）
#   make check    # PR 门禁：vet + test + lint + audit
#
# 【硬约束】main.go 用 //go:embed all:frontend/dist，
# frontend/dist 不存在时 go build / go test / go vet 全部编译失败。
# 因此下面 vet / test 一律依赖 frontend，把这层关系显式化——请不要去掉。
# （frontend 必须声明为 .PHONY：仓库里有同名目录 frontend/，
#   否则 make 会认为目标已是最新的，永远不执行 npm run build。）

.DEFAULT_GOAL := help

# wails CLI 钉死版本，不要用 @latest —— 会与 go.mod 的 wails 模块版本漂移。
WAILS_VERSION      := v2.16.0
GOVULNCHECK_VERSION := v1.8.0

.PHONY: help deps-go deps-fe frontend vet test lint audit fmt check dev clean

help: ## 列出全部 target
	@grep -E '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-10s %s\n", $$1, $$2}'

deps-go: ## 安装 Go 依赖 + 钉死 wails CLI 与 govulncheck
	go mod download
	go install github.com/wailsapp/wails/v2/cmd/wails@$(WAILS_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

deps-fe: ## 安装前端依赖（走 package-lock.json，可复现）
	cd frontend && npm ci

frontend: ## 构建前端，产出 frontend/dist（一切 go 命令的硬前置）
	cd frontend && npm run build

vet: frontend ## go vet（依赖 frontend，因 //go:embed）
	go vet ./...

test: frontend ## 后端单元测试
	go test ./...

lint: ## golangci-lint 静态检查
	golangci-lint run ./...

audit: ## 漏洞扫描（Go govulncheck + npm audit high 及以上）
	govulncheck ./...
	cd frontend && npm audit --audit-level=high

fmt: ## gofmt 格式化
	gofmt -w .

check: vet test lint audit ## PR 门禁一条命令：vet + test + lint + audit

dev: ## 开发热更新
	wails dev

clean: ## 清理构建产物
	rm -rf build/bin build/release frontend/dist

# ---------------------------------------------------------------------------
# 发布（计划 007）：一平台一次构建 → 打包 zip/DMG → 校验和
#
# 【Wails v2.16.0 硬约束，改这里之前先读】
#   1. 一平台一次 wails build，各自带 -clean。多 target 会产连字符名
#      （boce_tool_app-darwin-arm64），与 README 契约的下划线名对不上；
#      且 -clean 只对第一个 target 生效，后续产物会串台。
#   2. 每次构建都会清空 build/bin，所以构建完立刻暂存到
#      build/release/.stage/<平台>/，否则后一个平台会覆盖前一个的 .app。
#   3. 不用 -nopackage：darwin 要的是 build/bin/boce_tool_app.app bundle。
#   4. 一律加 -m -nosyncgomod，否则 wails build 会改写 go.mod / go.sum。
#   5. 不开 linux target：从 macOS 交叉编译 linux 会被 wails 静默跳过
#      （只打 warning、exit 0、无产物）。
#   6. -buildvcs=false 是 wails 硬编码，版本只能从 wails.json 的
#      info.productVersion + -ldflags -X main.Version 这两条通道注入。
#
# 【命名契约】zip / dmg 文件名只在 package-* 里出现，其余 target 一律不得重复。
# ---------------------------------------------------------------------------

VERSION ?= $(shell git describe --tags --always --dirty)
RELEASE := build/release
STAGE   := $(RELEASE)/.stage

.PHONY: build-darwin-arm64 build-darwin-amd64 build-windows-amd64 build-all \
        package-darwin package-windows package checksums release

# 产物都落在 build/bin 里，必须串行，不能并行跑这几个 target。
.NOTPARALLEL: build-all package-darwin package-windows package

build-darwin-arm64: ## 构建 macOS (Apple Silicon) .app → build/bin/boce_tool_app.app
	wails build -clean -platform darwin/arm64 -m -nosyncgomod -ldflags "-X main.Version=$(VERSION)"
	rm -rf $(STAGE)/darwin-arm64
	mkdir -p $(STAGE)/darwin-arm64
	cp -R build/bin/boce_tool_app.app $(STAGE)/darwin-arm64/
	cp scripts/fix_mac_quarantine.sh $(STAGE)/darwin-arm64/

build-darwin-amd64: ## 构建 macOS (Intel) .app → build/bin/boce_tool_app.app
	wails build -clean -platform darwin/amd64 -m -nosyncgomod -ldflags "-X main.Version=$(VERSION)"
	rm -rf $(STAGE)/darwin-amd64
	mkdir -p $(STAGE)/darwin-amd64
	cp -R build/bin/boce_tool_app.app $(STAGE)/darwin-amd64/
	cp scripts/fix_mac_quarantine.sh $(STAGE)/darwin-amd64/

build-windows-amd64: ## 构建 Windows 64 位 → build/bin/boce_tool_app.exe
	wails build -clean -platform windows/amd64 -m -nosyncgomod -ldflags "-X main.Version=$(VERSION)"
	rm -rf $(STAGE)/windows-amd64
	mkdir -p $(STAGE)/windows-amd64
	cp build/bin/boce_tool_app.exe $(STAGE)/windows-amd64/

build-all: build-darwin-arm64 build-darwin-amd64 build-windows-amd64 ## 串行构建三平台

package-darwin: build-darwin-arm64 build-darwin-amd64 ## 打包两个 macOS 平台（zip + DMG）
	mkdir -p $(RELEASE)
	rm -f $(RELEASE)/boce_tool_app_darwin_arm64.zip $(RELEASE)/boce_tool_app_darwin_arm64.dmg
	(cd $(STAGE)/darwin-arm64 && zip -ry $(CURDIR)/$(RELEASE)/boce_tool_app_darwin_arm64.zip boce_tool_app.app fix_mac_quarantine.sh)
	hdiutil create -volname "BOCE 检测工具" -srcfolder $(STAGE)/darwin-arm64 -ov -format UDZO $(RELEASE)/boce_tool_app_darwin_arm64.dmg
	rm -f $(RELEASE)/boce_tool_app_darwin_amd64.zip $(RELEASE)/boce_tool_app_darwin_amd64.dmg
	(cd $(STAGE)/darwin-amd64 && zip -ry $(CURDIR)/$(RELEASE)/boce_tool_app_darwin_amd64.zip boce_tool_app.app fix_mac_quarantine.sh)
	hdiutil create -volname "BOCE 检测工具" -srcfolder $(STAGE)/darwin-amd64 -ov -format UDZO $(RELEASE)/boce_tool_app_darwin_amd64.dmg

package-windows: build-windows-amd64 ## 打包 Windows 平台（zip）
	mkdir -p $(RELEASE)
	rm -f $(RELEASE)/boce_tool_app_windows_amd64.zip
	(cd $(STAGE)/windows-amd64 && zip -y $(CURDIR)/$(RELEASE)/boce_tool_app_windows_amd64.zip boce_tool_app.exe)

package: package-darwin package-windows ## 打包全部平台（命名契约唯一所在地）
	rm -rf $(STAGE)
	ls -la $(RELEASE)

checksums: ## 生成 build/release/SHA256SUMS.txt（必须在 package 之后）
	cd $(RELEASE) && shasum -a 256 boce_tool_app_*.zip boce_tool_app_*.dmg > SHA256SUMS.txt
	cat $(RELEASE)/SHA256SUMS.txt

release: ## 一键发布：注入版本 → 三平台构建打包 → SHA256SUMS（结束还原 wails.json）
	@set -e; \
	backup=$$(mktemp); cp wails.json "$$backup"; \
	trap 'cp "$$backup" wails.json; rm -f "$$backup"' EXIT INT TERM; \
	sed -e 's/"productVersion"[[:space:]]*:[[:space:]]*"[^"]*"/"productVersion": "$(VERSION)"/' wails.json > wails.json.tmp && mv wails.json.tmp wails.json; \
	grep -q '"productVersion": "$(VERSION)"' wails.json || { echo "版本注入失败：wails.json 的 productVersion 未被改写为 $(VERSION)" >&2; exit 1; }; \
	echo "productVersion → $(VERSION)"; \
	$(MAKE) --no-print-directory VERSION=$(VERSION) package; \
	$(MAKE) --no-print-directory checksums
