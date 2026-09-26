package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/net/idna"
)

// TargetList 是检测目标的解析结果：Valid 为归一化后的合法 host，
// Invalid 为原样保留的非法项。非法项不再静默丢弃。
type TargetList struct {
	Valid   []string
	Invalid []string
}

var (
	// hostLabelPattern 对应前端 App.vue isValidHost 的标签规则（不区分大小写）。
	hostLabelPattern = regexp.MustCompile(`(?i)^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`)
	// httpSchemePattern 对应前端 normalizeTarget 的 http(s) scheme 分支。
	httpSchemePattern = regexp.MustCompile(`(?i)^https?://`)
	// schemePattern 对应前端 normalizeTarget 的「非 http(s) scheme 直接拒绝」分支。
	schemePattern = regexp.MustCompile(`(?i)^[a-z][a-z\d+.-]*://`)
	// ipv4PartPattern 对应前端 isValidIPv4 的分段规则。
	ipv4PartPattern = regexp.MustCompile(`^\d{1,3}$`)
)

// normalizeTargets 归一化并严格校验一批目标：合法项去重后进 Valid，
// 非法项进 Invalid。纯空白项沿用前端 filter(Boolean) 的约定直接忽略。
func normalizeTargets(targets []string) TargetList {
	result := TargetList{
		Valid:   make([]string, 0, len(targets)),
		Invalid: make([]string, 0),
	}
	seen := make(map[string]struct{}, len(targets))

	for _, target := range targets {
		target = trimTarget(target)
		if target == "" {
			continue
		}
		host := normalizeTarget(target)
		if host == "" {
			result.Invalid = append(result.Invalid, target)
			continue
		}
		if _, ok := seen[host]; ok {
			continue
		}
		seen[host] = struct{}{}
		result.Valid = append(result.Valid, host)
	}

	return result
}

// parseTargetsText 解析粘贴 / TXT 导入的自由文本，分隔与去引号规则沿用旧实现，
// 校验改走与前端一致的严格规则，非法项进 Invalid。
func parseTargetsText(content string) TargetList {
	content = strings.TrimPrefix(content, "\ufeff")
	parts := strings.FieldsFunc(content, func(r rune) bool {
		switch r {
		case '\n', '\r', '\t', ' ', ',', ';', '，', '；':
			return true
		default:
			return false
		}
	})

	for idx, part := range parts {
		parts[idx] = strings.Trim(part, `"'`)
	}

	return normalizeTargets(parts)
}

// normalizeTarget 把单条输入归一化成小写 host；非法时返回 ""。
// 规则逐条对应前端 App.vue 的 normalizeTarget / isValidHost / isValidIPv4。
//
// IDN 归一：前端用 WHATWG new URL 解析 host，它会把国际化域名转成 punycode
// （"中国.com" → "xn--fiqs8s.com"）；Go 的 net/url 不做这件事。因此这里在既有
// isValidHost / isValidIPv4 校验之前，先用 idna.Lookup.ToASCII 把 host 归一成
// punycode，再走原校验。转换失败一律当作非法 host 返回 ""。
// DNS 查询本来就该用 punycode 形式 —— BOCE API 的 host= 参数要的就是
// xn--fiqs8s.com，不是 "中国.com"，所以转换是正确行为，拒绝才是错的。
//
// 已知差异（都出在 URL 解析层，不在 isValidHost / isValidIPv4 —— 那两层是逐条一致的）：
//  1. 带前导零 / 十六进制 / 八进制的 IPv4 写法：前端 new URL("http://01.2.3.4")
//     会先归一化成 "1.2.3.4" 再校验，故放行；Go 侧按 isValidIPv4 逐字判定，拒绝。
//     这条差异是刻意保留的（Go 侧更严格），不要顺手改成放行。
//  2. idna.Lookup 是 UTS #46 lookup profile（MapForLookup + BidiRule + STD3），
//     与 WHATWG host 解析在边缘输入上不完全等价，见 toASCIIHost 的注释。
//     以「能吃下中文域名并转成 punycode」为首要目标。
func normalizeTarget(input string) string {
	raw := trimTarget(input)
	if raw == "" {
		return ""
	}
	// 对应前端 /[\s@]/ 的拒绝条件（含 U+FEFF，与 JS \s 一致）。
	if strings.IndexFunc(raw, isForbiddenTargetRune) >= 0 {
		return ""
	}

	var host string
	switch {
	case httpSchemePattern.MatchString(raw):
		parsed, err := url.Parse(raw)
		if err != nil || parsed.Host == "" {
			return ""
		}
		host = parsed.Hostname()
	case schemePattern.MatchString(raw):
		return ""
	default:
		// 对应前端 raw.replace(/^\/+|\/+$/g, '') 后再补 http:// 解析。
		parsed, err := url.Parse("http://" + strings.Trim(raw, "/"))
		if err != nil || parsed.Host == "" {
			return ""
		}
		host = parsed.Hostname()
	}

	host = strings.ToLower(host)
	host = toASCIIHost(host)
	if host == "" {
		return ""
	}
	if !isValidHost(host) {
		return ""
	}
	return host
}

// toASCIIHost 把 host 归一成 punycode / ASCII 小写形式；无法归一时返回 ""。
//
// 用 idna.Lookup（RFC 5891 §5 推荐的查询 profile）而不是裸 Punycode：
// 它自带 UTS #46 的 MapForLookup（大小写 / 全角 / 异体字映射）、BidiRule 与
// STD3 字符规则，行为与浏览器把 host 送进 DNS 前做的归一最接近。
//
// 与前端 WHATWG host 解析的残余差异（均已实测，见 normalizeTarget 注释）：
//   - Lookup 开了 checkHyphens，拒绝第 3、4 位同为连字符的标签（如 ab--cd.com、
//     aa--bb--cc.com）；WHATWG new URL 与既有 isValidHost 都放过。Go 侧略严。
//   - Lookup 会解码 xn-- 标签，非法 punycode（如 xn--a.com、xn--abc.com）判为
//     非法；WHATWG 同样拒绝，但既有 isValidHost 会放过 —— 这条是向前端靠拢。
//   - Lookup 不做 IPv4 的前导零 / 八进制 / 十六进制归一（那是 WHATWG 的
//     IPv4 parser），所以 01.2.3.4 仍会被 Go 侧拒绝 —— 这是刻意保留的差异。
//
// 上面三条之外，纯 ASCII 合法 host（isValidHost 放行的那套 LDH 标签）
// 不会被 Lookup 改写，也不会被它拒掉。
func toASCIIHost(host string) string {
	ascii, err := idna.Lookup.ToASCII(host)
	if err != nil {
		return ""
	}
	return ascii
}

// isValidHost 与前端 App.vue 的 isValidHost 逐条对应；host 须已转小写。
func isValidHost(host string) bool {
	if isValidIPv4(host) {
		return true
	}
	if len(host) > 253 || !strings.Contains(host, ".") || strings.Contains(host, "..") {
		return false
	}

	labels := strings.Split(host, ".")
	topLabel := labels[len(labels)-1]
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 || !hostLabelPattern.MatchString(label) {
			return false
		}
	}
	return len(topLabel) >= 2 && containsASCIILetter(topLabel)
}

// isValidIPv4 与前端 App.vue 的 isValidIPv4 逐条对应（含拒绝前导零）。
func isValidIPv4(host string) bool {
	parts := strings.Split(host, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if !ipv4PartPattern.MatchString(part) {
			return false
		}
		value, err := strconv.Atoi(part)
		if err != nil || value < 0 || value > 255 {
			return false
		}
		// 对应前端 String(value) === part：拒绝 "01" 这类前导零写法。
		if strconv.Itoa(value) != part {
			return false
		}
	}
	return true
}

// targetValidationMessage 按前端既有文案格式生成校验提示。
func targetValidationMessage(rawText string, parsed TargetList) string {
	if len(parsed.Invalid) > 0 {
		examples := parsed.Invalid
		if len(examples) > 3 {
			examples = examples[:3]
		}
		rest := ""
		if len(parsed.Invalid) > 3 {
			rest = fmt.Sprintf(" 等 %d 条", len(parsed.Invalid))
		}
		return fmt.Sprintf("检测目标格式错误: %s%s", strings.Join(examples, "、"), rest)
	}
	if trimTarget(rawText) != "" && len(parsed.Valid) == 0 {
		return "未解析到有效域名或 URL"
	}
	return ""
}

// trimTarget 对应前端 String.prototype.trim：去掉首尾空白与 U+FEFF。
func trimTarget(value string) string {
	return strings.TrimFunc(value, func(r rune) bool {
		return r == '\ufeff' || unicode.IsSpace(r)
	})
}

func isForbiddenTargetRune(r rune) bool {
	return r == '@' || r == '\ufeff' || unicode.IsSpace(r)
}

func containsASCIILetter(value string) bool {
	return strings.IndexFunc(value, func(r rune) bool {
		return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
	}) >= 0
}
