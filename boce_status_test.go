package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// statusConstants 是中文状态文案的全集（顺序与 boce_status.go 的常量块一致）。
func statusConstants() []string {
	return []string{
		StatusPollution,
		StatusIntercepted,
		StatusUnfiled,
		StatusBlacklisted,
		StatusWalled,
		StatusSuspectWalled,
		StatusBadDomain,
		StatusNormal,
		StatusFiled,
		StatusUnregistered,
		StatusFailed,
		StatusUnknown,
	}
}

// typeConstants 是检测类型文案的全集。
func typeConstants() []string {
	return []string{
		TypePollution,
		TypeQQ,
		TypeWechat,
		TypeICP,
		TypeBlacklist,
		TypeWall,
	}
}

// TestJSStatusConstants 把 Go 与前端共用的中文文案契约变成机器可检：
// frontend/src/lib/constants.js 里必须逐字出现 Go 的 18 个常量值，
// 且 STATUS_LIST 的元素集合与 Go 的 12 个状态常量集合完全相等（无多无少）。
func TestJSStatusConstants(t *testing.T) {
	const jsPath = "frontend/src/lib/constants.js"

	raw, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v（go test 的工作目录应为包目录即仓库根）", jsPath, err)
	}
	js := string(raw)

	for _, value := range append(statusConstants(), typeConstants()...) {
		if !strings.Contains(js, "'"+value+"'") && !strings.Contains(js, `"`+value+`"`) {
			t.Fatalf("%s 中找不到字符串字面量 %q —— 与 Go 常量不一致", jsPath, value)
		}
	}

	jsStatuses := jsStatusListValues(t, js)
	want := map[string]bool{}
	for _, value := range statusConstants() {
		want[value] = true
	}
	got := map[string]bool{}
	for _, value := range jsStatuses {
		got[value] = true
	}

	for value := range want {
		if !got[value] {
			t.Fatalf("STATUS_LIST 缺少状态 %q（JS %v）", value, jsStatuses)
		}
	}
	for value := range got {
		if !want[value] {
			t.Fatalf("STATUS_LIST 多出状态 %q（JS %v）", value, jsStatuses)
		}
	}
	if len(jsStatuses) != len(statusConstants()) {
		t.Fatalf("STATUS_LIST 元素个数 = %d, want %d（%v）", len(jsStatuses), len(statusConstants()), jsStatuses)
	}
}

// jsStatusListValues 解析 constants.js 的 STATUS_LIST 数组，
// 元素可以是字符串字面量，也可以是同文件 export const 的标识符。
func jsStatusListValues(t *testing.T, js string) []string {
	t.Helper()

	list := regexp.MustCompile(`export const STATUS_LIST\s*=\s*\[([^\]]*)\]`).FindStringSubmatch(js)
	if list == nil {
		t.Fatalf("constants.js 中找不到 export const STATUS_LIST = [...] ")
	}

	values := map[string]string{}
	for _, pair := range regexp.MustCompile(`export const (\w+)\s*=\s*['"]([^'"]+)['"]`).FindAllStringSubmatch(js, -1) {
		values[pair[1]] = pair[2]
	}

	resolved := []string{}
	for _, item := range regexp.MustCompile(`['"][^'"]+['"]|\w+`).FindAllString(list[1], -1) {
		if strings.HasPrefix(item, "'") || strings.HasPrefix(item, `"`) {
			resolved = append(resolved, item[1:len(item)-1])
			continue
		}
		value, ok := values[item]
		if !ok {
			t.Fatalf("STATUS_LIST 元素 %s 既不是字符串字面量也不是 constants.js 的常量", item)
		}
		resolved = append(resolved, value)
	}
	return resolved
}

// TestStatusConstantsCoverSummarizeBuckets 断言 summarizeRows 的 4 个分类
// 覆盖的文案集合与常量表一致：除 StatusUnknown（设计上不入任何分类）外无遗漏，
// 常量表之外的文案无多余。
func TestStatusConstantsCoverSummarizeBuckets(t *testing.T) {
	constants := statusConstants()
	seen := map[string]bool{}
	for _, status := range constants {
		if status == "" {
			t.Fatalf("状态常量出现空值")
		}
		if seen[status] {
			t.Fatalf("状态常量值 %q 重复", status)
		}
		seen[status] = true
	}

	unbucketed := []string{}
	for _, status := range constants {
		got := summarizeRows([]DetectRow{{Status: status}})
		buckets := 0
		if got.Pollution > 0 {
			buckets++
		}
		if got.Normal > 0 {
			buckets++
		}
		if got.Unregistered > 0 {
			buckets++
		}
		if got.Failed > 0 {
			buckets++
		}
		if buckets == 0 {
			unbucketed = append(unbucketed, status)
			continue
		}
		if buckets != 1 {
			t.Fatalf("summarizeRows(%q) 落入 %d 个分类, want 1: %#v", status, buckets, got)
		}
		if got.Total != 1 || got.Checked != 1 {
			t.Fatalf("summarizeRows(%q) Total/Checked = %d/%d, want 1/1", status, got.Total, got.Checked)
		}
	}

	// 无遗漏：只有 StatusUnknown 设计上不入分类。
	if len(unbucketed) != 1 || unbucketed[0] != StatusUnknown {
		t.Fatalf("不入分类的状态 = %v, want 仅 [%q]", unbucketed, StatusUnknown)
	}

	// 无多余：常量表之外的文案不入任何分类。
	unknown := summarizeRows([]DetectRow{{Status: "不在常量表中的文案"}})
	if unknown.Pollution+unknown.Normal+unknown.Unregistered+unknown.Failed != 0 {
		t.Fatalf("常量表之外的文案被分类了: %#v", unknown)
	}
}

// TestTypeConstantsMatchDetectionLabels 断言 6 个类型常量与 runXxxDetection
// 写入 DetectRow.Type 的 label 一一对应，且常量值两两不同。
func TestTypeConstantsMatchDetectionLabels(t *testing.T) {
	tests := []struct {
		fn     string
		consts string
	}{
		{fn: "runPollutionDetection", consts: "TypePollution"},
		{fn: "runQQDetection", consts: "TypeQQ"},
		{fn: "runWechatDetection", consts: "TypeWechat"},
		{fn: "runICPDetection", consts: "TypeICP"},
		{fn: "runBlacklistDetection", consts: "TypeBlacklist"},
		{fn: "runWallDetection", consts: "TypeWall"},
	}

	raw, err := os.ReadFile("boce_detect.go")
	if err != nil {
		t.Fatalf("ReadFile(boce_detect.go) error = %v", err)
	}
	bodies := functionBodies(string(raw))

	names := []string{"TypePollution", "TypeQQ", "TypeWechat", "TypeICP", "TypeBlacklist", "TypeWall"}
	for _, tt := range tests {
		body, ok := bodies[tt.fn]
		if !ok {
			t.Fatalf("boce_detect.go 中找不到 func %s", tt.fn)
		}
		used := []string{}
		for _, name := range names {
			if regexp.MustCompile(`\b` + name + `\b`).MatchString(body) {
				used = append(used, name)
			}
		}
		if len(used) != 1 || used[0] != tt.consts {
			t.Fatalf("func %s 使用的类型常量 = %v, want [%s]", tt.fn, used, tt.consts)
		}
	}

	seen := map[string]bool{}
	for _, value := range typeConstants() {
		if value == "" {
			t.Fatalf("类型常量出现空值")
		}
		if seen[value] {
			t.Fatalf("类型常量值 %q 重复", value)
		}
		seen[value] = true
	}
	if len(typeConstants()) != len(tests) {
		t.Fatalf("类型常量个数 = %d, want %d", len(typeConstants()), len(tests))
	}
}

// functionBodies 按顶层 func 切开源码，返回函数名 -> 函数文本。
func functionBodies(source string) map[string]string {
	parts := regexp.MustCompile(`(?m)^func `).Split(source, -1)
	bodies := map[string]string{}
	for _, part := range parts[1:] {
		name := regexp.MustCompile(`^(\w+)`).FindStringSubmatch(part)
		if name == nil {
			continue
		}
		bodies[name[1]] = part
	}
	return bodies
}
