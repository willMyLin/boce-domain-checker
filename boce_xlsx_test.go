package main

import (
	"encoding/xml"
	"strings"
	"testing"
)

// TestEscapeXMLStripsControlChars 断言转义前剔除 XML 1.0 非法字符：
// 0x00-0x08、0x0B、0x0C、0x0E-0x1F 等必须删掉，\t \n \r 原样交给 xml.EscapeText。
// 不改动表头与列结构。
func TestEscapeXMLStripsControlChars(t *testing.T) {
	var got strings.Builder
	escapeXML(&got, "a\x00b\x1fc")
	if want := "abc"; got.String() != want {
		t.Fatalf("escapeXML(控制字符) = %q, want %q", got.String(), want)
	}

	const legal = "a\tb\nc\rd"
	var gotLegal strings.Builder
	escapeXML(&gotLegal, legal)
	var wantLegal strings.Builder
	if err := xml.EscapeText(&wantLegal, []byte(legal)); err != nil {
		t.Fatalf("xml.EscapeText() error = %v", err)
	}
	if gotLegal.String() != wantLegal.String() {
		t.Fatalf("escapeXML(\t\n\r) = %q, want %q（合法字符不得被删）", gotLegal.String(), wantLegal.String())
	}

	worksheet := buildWorksheetXML([]DetectRow{{
		Type:        TypePollution,
		Target:      "a\x00b\x1fc",
		Status:      StatusNormal,
		CheckedAt:   "2026-06-12 12:00:00",
		ErrorRemark: "ok\x07&\x0bchecked",
	}})
	for _, illegal := range []string{"\x00", "\x07", "\x0b", "\x1f"} {
		if strings.Contains(worksheet, illegal) {
			t.Fatalf("worksheet 仍含非法控制字符 %q: %q", illegal, worksheet)
		}
	}
	for _, want := range []string{"<t>abc</t>", "<t>ok&amp;checked</t>", "类型", "检测目标", "状态", "检测时间", "错误 / 备注"} {
		if !strings.Contains(worksheet, want) {
			t.Fatalf("worksheet 缺少 %q（表头与列结构不得改动）: %q", want, worksheet)
		}
	}
}

// TestStripInvalidXMLCharsKeepsLegalRanges 断言只删 XML 1.0 Char 之外的字符。
// 非 BMP / C1 / 代理区字符在源码里一律写成 Go 转义，不落不可见字符。
func TestStripInvalidXMLCharsKeepsLegalRanges(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "控制字符全删", in: "\x00\x01\x08\x0b\x0c\x0e\x1f", want: ""},
		{name: "保留 tab/换行/回车", in: "\t\n\r", want: "\t\n\r"},
		{name: "保留普通文本与中文", in: "a<b>&'\"中文", want: "a<b>&'\"中文"},
		{name: "删 U+FFFE/U+FFFF", in: "a\ufffe\uffffb", want: "ab"},
		{name: "保留 C1 区（XML 1.0 允许）", in: "a\u0080\u009fb", want: "a\u0080\u009fb"},
		{name: "保留 U+007F 与 U+00A0", in: "a\u007f\u00a0b", want: "a\u007f\u00a0b"},
		{name: "保留增补平面", in: "a\U00010000b", want: "a\U00010000b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripInvalidXMLChars(tt.in); got != tt.want {
				t.Fatalf("stripInvalidXMLChars(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
