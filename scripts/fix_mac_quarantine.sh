#!/bin/sh
# 去掉 macOS 下载隔离属性。未签名应用首次打开会提示「文件已损坏」或「无法打开」。
#
# 用法：
#   ./fix_mac_quarantine.sh              # 自动处理同目录下的 boce_tool_app.app
#   ./fix_mac_quarantine.sh <路径>       # 显式指定 .app 目录或可执行文件
#
# 注意：DMG 只是容器格式，**不能**解除 Gatekeeper 隔离。
# 未签名应用从 DMG 拖出来后仍会被拦，必须执行本脚本（或手动 xattr）。
set -eu

target="${1:-}"

if [ -z "$target" ]; then
	cd "$(dirname "$0")"

	# 发布物是 .app bundle；裸二进制名保留是为了兼容旧 zip 与源码直建产物。
	for candidate in boce_tool_app.app boce_tool_app_darwin_arm64 boce_tool_app_darwin_amd64; do
		if [ -e "$candidate" ]; then
			target="$candidate"
			break
		fi
	done
fi

if [ -z "$target" ] || [ ! -e "$target" ]; then
	echo "未找到 boce_tool_app.app。"
	echo "请确认本脚本和 boce_tool_app.app 在同一目录后重新运行，"
	echo "或显式指定路径： ./fix_mac_quarantine.sh /Applications/BOCE\ 检测工具.app"
	exit 1
fi

if command -v xattr >/dev/null 2>&1; then
	xattr -dr com.apple.quarantine "$target" 2>/dev/null || true
fi

if [ -d "$target" ]; then
	# .app 是目录，内部可执行文件同样需要执行位
	chmod -R +x "$target/Contents/MacOS" 2>/dev/null || true
else
	chmod +x "$target"
fi

echo "已处理: $target"
echo "现在可以打开: $target"
