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
