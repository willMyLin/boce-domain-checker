package main

import (
	"reflect"
	"strings"
	"testing"
)

// TestValidateTargetsRejectsLooseCases 覆盖旧的宽松解析会放过、
// 但前端 isValidHost / isValidIPv4 一律拒绝的输入。
func TestValidateTargetsRejectsLooseCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "无点主机", input: "abc"},
		{name: "连续点", input: "a..b.com"},
		{name: "空 scheme", input: "http://"},
		{name: "非 http scheme", input: "ftp://x.com"},
		{name: "带假 scheme", input: "ahttp://x.com"},
		{name: "非法端口", input: "example.com:abc"},
		{name: "超长标签", input: strings.Repeat("a", 65) + ".com"},
		{name: "超长 host", input: strings.Repeat("ab.", 90) + "com"},
		{name: "TLD 无字母", input: "a.12"},
		{name: "TLD 单字符", input: "a.c"},
		{name: "标签首连字符", input: "-a.com"},
		{name: "标签尾连字符", input: "a-.com"},
		{name: "标签含下划线", input: "a_b.com"},
		{name: "前导零 IPv4（WHATWG 会归一化，见 normalizeTarget 注释）", input: "01.2.3.4"},
		{name: "越界 IPv4", input: "1.2.3.256"},
		{name: "五段伪 IPv4", input: "1.2.3.4.5"},
		{name: "含 @", input: "user@x.com"},
		// IDN 中文域名原形已移至 TestValidateTargetsAcceptsRealCases —— 操作者 2026-09-25 拍板 IDN 转 punycode 放行，见 normalizeTarget
	}

	app := &App{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := app.ValidateTargets(tt.input)
			if len(got.Valid) != 0 {
				t.Fatalf("ValidateTargets(%q).Valid = %v, want 空", tt.input, got.Valid)
			}
			if want := []string{tt.input}; !reflect.DeepEqual(got.Invalid, want) {
				t.Fatalf("ValidateTargets(%q).Invalid = %#v, want %#v", tt.input, got.Invalid, want)
			}
			if got.Message == "" {
				t.Fatalf("ValidateTargets(%q).Message = \"\", want 非空中文提示", tt.input)
			}
		})
	}
}

// TestValidateTargetsAcceptsRealCases 断言真实域名 / URL / IPv4 归一化后进 Valid。
func TestValidateTargetsAcceptsRealCases(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{name: "裸域名", text: "www.baidu.com", want: []string{"www.baidu.com"}},
		{name: "带路径查询的 URL", text: "https://www.baidu.com/path?q=1", want: []string{"www.baidu.com"}},
		{name: "IPv4", text: "1.2.3.4", want: []string{"1.2.3.4"}},
		{name: "带端口与路径", text: "boce.com:8080/x", want: []string{"boce.com"}},
		{name: "数字开头标签", text: "1.2.3.com", want: []string{"1.2.3.com"}},
		{
			name: "混合分隔符",
			text: "boce.com, example.com；http://foo.test",
			want: []string{"boce.com", "example.com", "foo.test"},
		},
		{
			name: "大小写与去重",
			text: "www.baidu.com\nhttps://WWW.BAIDU.COM/search?q=1",
			want: []string{"www.baidu.com"},
		},
		{name: "IDN 中文域名原形", text: "中国.com", want: []string{"xn--fiqs8s.com"}},
	}

	app := &App{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := app.ValidateTargets(tt.text)
			if !reflect.DeepEqual(got.Valid, tt.want) {
				t.Fatalf("ValidateTargets(%q).Valid = %#v, want %#v", tt.text, got.Valid, tt.want)
			}
			if len(got.Invalid) != 0 {
				t.Fatalf("ValidateTargets(%q).Invalid = %#v, want 空", tt.text, got.Invalid)
			}
			if got.Message != "" {
				t.Fatalf("ValidateTargets(%q).Message = %q, want 空", tt.text, got.Message)
			}
		})
	}
}

// TestNormalizeTargetsRejectsInlineJunk 断言带空白 / @ 的单条输入进 Invalid，
// 纯空白项按前端 filter(Boolean) 的约定直接忽略（parseTargetsText 会按空白切分，
// 所以这条单条路径只能从 normalizeTargets 覆盖）。
func TestNormalizeTargetsRejectsInlineJunk(t *testing.T) {
	got := normalizeTargets([]string{"a b.com", "x\ty.com", "user@x.com", "   ", "\ufeffok.com"})
	want := TargetList{
		Valid:   []string{"ok.com"},
		Invalid: []string{"a b.com", "x\ty.com", "user@x.com"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeTargets() = %#v, want %#v", got, want)
	}
}

// TestParseTargetsTextReportsInvalid 断言非法项不再被静默丢弃，而是原样进 Invalid。
func TestParseTargetsTextReportsInvalid(t *testing.T) {
	content := "\ufeffhttps://www.baidu.com/path\nabc\nboce.com, a..b.com；http://foo.test\nftp://x.com"
	got := parseTargetsText(content)

	wantValid := []string{"www.baidu.com", "boce.com", "foo.test"}
	wantInvalid := []string{"abc", "a..b.com", "ftp://x.com"}
	if !reflect.DeepEqual(got.Valid, wantValid) {
		t.Fatalf("parseTargetsText().Valid = %#v, want %#v", got.Valid, wantValid)
	}
	if !reflect.DeepEqual(got.Invalid, wantInvalid) {
		t.Fatalf("parseTargetsText().Invalid = %#v, want %#v", got.Invalid, wantInvalid)
	}
}

// TestTargetValidationMessage 断言 Go 侧提示文案与前端现有格式一致。
func TestTargetValidationMessage(t *testing.T) {
	many := TargetList{Valid: []string{"ok.com"}, Invalid: []string{"a", "b", "c", "d", "e"}}
	if got, want := targetValidationMessage("a b c d e ok.com", many), "检测目标格式错误: a、b、c 等 5 条"; got != want {
		t.Fatalf("targetValidationMessage() = %q, want %q", got, want)
	}

	few := TargetList{Valid: []string{"ok.com"}, Invalid: []string{"a", "b"}}
	if got := targetValidationMessage("a b ok.com", few); got != "检测目标格式错误: a、b" {
		t.Fatalf("targetValidationMessage() = %q", got)
	}

	if got := targetValidationMessage("   ", TargetList{}); got != "" {
		t.Fatalf("targetValidationMessage(空白) = %q, want 空", got)
	}
}

// TestNormalizeTargetEncodesIDN 断言 IDN 中文域名在既有 isValidHost 校验之前
// 先被归一成 punycode。这是产品回归修复：域名校验收口到 Go 侧后，中文域名
// 因 net/url 不做 host 规范化而被标签正则误拒，与前端 new URL 的行为不一致。
// punycode 形式才是 DNS / BOCE API host= 参数真正要的东西，转换是正确行为。
func TestNormalizeTargetEncodesIDN(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "中文域名", input: "中国.com", want: "xn--fiqs8s.com"},
		{name: "中文域名 + 中文 TLD", input: "例子.公司", want: "xn--fsqu00a.xn--55qx5d"},
		{name: "混合中英（中文在前）", input: "中国.baidu.com", want: "xn--fiqs8s.baidu.com"},
		{name: "混合中英（中文 TLD）", input: "baidu.中国", want: "baidu.xn--fiqs8s"},
		{name: "带 scheme 与路径", input: "http://例子.公司/x", want: "xn--fsqu00a.xn--55qx5d"},
		{name: "英文大写 + 中文", input: "WWW.中国.COM", want: "www.xn--fiqs8s.com"},
		{name: "表意句点 U+3002", input: "例子。公司", want: "xn--fsqu00a.xn--55qx5d"},
		{name: "全角拉丁（UTS #46 映射）", input: "ＢＡＩＤＵ.com", want: "baidu.com"},
		{name: "拉丁扩展（非中文 IDN）", input: "münchen.de", want: "xn--mnchen-3ya.de"},
		{name: "已是 punycode 原样保留", input: "xn--fiqs8s.com", want: "xn--fiqs8s.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeTarget(tt.input); got != tt.want {
				t.Fatalf("normalizeTarget(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestNormalizeTargetKeepsRejectingLooseCases 断言加了 IDN 归一之后，
// 既有「必须拒绝」的输入不会被顺手放行。
func TestNormalizeTargetKeepsRejectingLooseCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "无点主机", input: "abc"},
		{name: "连续点", input: "a..b.com"},
		// 前导零 IPv4：WHATWG 会归一化成 1.2.3.4 再放行，Go 侧刻意保持严格。
		{name: "前导零 IPv4", input: "01.2.3.4"},
		{name: "尾点", input: "中国.com."},
		{name: "非法 punycode 标签", input: "xn--a.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeTarget(tt.input); got != "" {
				t.Fatalf("normalizeTarget(%q) = %q, want \"\"", tt.input, got)
			}
		})
	}
}

// TestNormalizeTargetPunycodePassesHostCheck 断言 IDN 归一的产物仍然要过
// isValidHost —— punycode 不是后门，只是把 host 变成 ASCII 再走同一套规则。
func TestNormalizeTargetPunycodePassesHostCheck(t *testing.T) {
	inputs := []string{"中国.com", "例子.公司", "中国.baidu.com", "baidu.中国", "münchen.de", "点.中国"}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			got := normalizeTarget(input)
			if got == "" {
				t.Fatalf("normalizeTarget(%q) = \"\", want 非空 punycode", input)
			}
			if !isValidHost(got) {
				t.Fatalf("normalizeTarget(%q) = %q 未通过 isValidHost", input, got)
			}
			for _, r := range got {
				if r > 0x7F {
					t.Fatalf("normalizeTarget(%q) = %q 含非 ASCII 字符 %U，punycode 归一未生效", input, got, r)
				}
			}
		})
	}
}

// TestParseTargetsTextEncodesIDN 断言粘贴 / TXT 导入路径同样吃到 IDN 归一。
func TestParseTargetsTextEncodesIDN(t *testing.T) {
	got := parseTargetsText("\ufeff中国.com 例子.公司\nabc")
	want := TargetList{
		Valid:   []string{"xn--fiqs8s.com", "xn--fsqu00a.xn--55qx5d"},
		Invalid: []string{"abc"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseTargetsText() = %#v, want %#v", got, want)
	}
}
