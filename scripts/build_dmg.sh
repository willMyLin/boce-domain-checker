#!/usr/bin/env bash
set -euo pipefail

# scripts/build_dmg.sh - 构建高颜值 macOS DMG 镜像
#
# 用法: ./scripts/build_dmg.sh <卷标名称> <源文件夹路径> <输出DMG路径>
# 示例: ./scripts/build_dmg.sh "BOCE 检测工具" build/release/.stage/darwin-arm64 build/release/boce_tool_app_darwin_arm64.dmg

VOLNAME="${1:-BOCE 检测工具}"
SOURCE_DIR="${2:?缺少源文件夹路径}"
OUTPUT_DMG="${3:?缺少输出DMG路径}"

# 确保输出目录存在
mkdir -p "$(dirname "$OUTPUT_DMG")"
rm -f "$OUTPUT_DMG"

# 准备仅包含 .app 的干净源目录，确保 DMG 视图纯净无杂质、无滚动条
DMG_SRC_DIR=$(mktemp -d /tmp/boce_dmg_src.XXXXXX)
trap 'rm -rf "$DMG_SRC_DIR"' EXIT

if [ -d "$SOURCE_DIR/Boce.app" ]; then
  cp -R "$SOURCE_DIR/Boce.app" "$DMG_SRC_DIR/"
else
  cp -R "$SOURCE_DIR"/* "$DMG_SRC_DIR/"
fi

# 查找卷标图标
ICON_ARGS=()
if [ -f "build/bin/Boce.app/Contents/Resources/iconfile.icns" ]; then
  ICON_ARGS=(--volicon "build/bin/Boce.app/Contents/Resources/iconfile.icns")
elif [ -f "build/darwin/iconfile.icns" ]; then
  ICON_ARGS=(--volicon "build/darwin/iconfile.icns")
fi

# 背景图路径
BG_IMG="build/darwin/dmg-background.png"

# 优先使用 create-dmg 打包经典漂亮的 macOS 拖拽安装镜像
if command -v create-dmg >/dev/null 2>&1 && [ -f "$BG_IMG" ]; then
  echo "==> 使用 create-dmg 打包精美 DMG: $OUTPUT_DMG"
  set +e
  create-dmg \
    --volname "$VOLNAME" \
    "${ICON_ARGS[@]}" \
    --background "$BG_IMG" \
    --window-pos 200 120 \
    --window-size 700 440 \
    --icon-size 110 \
    --icon "Boce.app" 160 135 \
    --app-drop-link 540 135 \
    --hide-extension "Boce.app" \
    --format UDZO \
    --no-internet-enable \
    --overwrite \
    "$OUTPUT_DMG" \
    "$DMG_SRC_DIR"
  STATUS=$?
  set -e

  # 若 create-dmg 成功生成产物，则直接返回
  if [ $STATUS -eq 0 ] && [ -f "$OUTPUT_DMG" ]; then
    echo "✓ create-dmg 打包成功: $OUTPUT_DMG"
    exit 0
  fi

  echo "警告: create-dmg 执行未完成 (退出码: $STATUS)，自动降级回退到 hdiutil 兜底打包..."
  rm -f "$OUTPUT_DMG"
fi

# 兜底降级方案：使用 macOS 内置原生 hdiutil
echo "==> 使用原生 hdiutil 打包 DMG: $OUTPUT_DMG"
hdiutil create -volname "$VOLNAME" -srcfolder "$DMG_SRC_DIR" -ov -format UDZO "$OUTPUT_DMG"
echo "✓ hdiutil 打包完成: $OUTPUT_DMG"
