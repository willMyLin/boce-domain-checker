package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 检查更新的两条取数路径 + 一条降级。
//
// 【为什么不直接打 GitHub API】未认证限速 60 次/小时/IP（响应头 X-RateLimit-Limit），
// 用户一多就互相挤兑，还会撞上 403/429。所以便宜的那条（网页 302）当主路，API 当副路。
//
// 主路 GET github.com/<owner>/<repo>/releases/latest
//
//	—— 不跟随重定向，直接读 Location: .../releases/tag/vX.Y.Z 就能拿到版本号。
//	   实测响应头里没有任何 rate-limit 头，且它走 GitHub 网页边缘而非 API 限速池。
//
// 副路 GET api.github.com/repos/<owner>/<repo>/releases/latest
//
//	—— 主路挂了才走。能多拿更新说明（body）和各平台资产大小，但吃限速。
//
// 两条都挂 → 报错，由调用方决定静默吞掉（启动时自动检查）还是提示用户（手动检查）。
// 刻意不做本地缓存：版本检查本来就轻，缓存带来的陈旧问题比省下的那点请求更烦。
var (
	updateRedirectURL = "https://github.com/willMyLin/boce-domain-checker/releases/latest"
	updateAPIURL      = "https://api.github.com/repos/willMyLin/boce-domain-checker/releases/latest"
)

// UpdateInfo 是给界面看的检查结果。Source 标明数据来自哪条路，便于排障。
type UpdateInfo struct {
	Current    string `json:"current"`
	Latest     string `json:"latest"`
	HasUpdate  bool   `json:"hasUpdate"`
	ReleaseURL string `json:"releaseUrl"`
	Notes      string `json:"notes"`
	Source     string `json:"source"`
}

// errUpdateRateLimited 让调用方分辨「限速」和「彻底断网」——限速时提示等一会儿，断网时提示检查网络。
var errUpdateRateLimited = errors.New("GitHub API 限速")

// checkUpdate 按「302 → API」的顺序取最新版本，与当前版本比较后返回。
// 调用方要自己带超时；本函数不额外造 context。
func checkUpdate(ctx context.Context, client *http.Client) (UpdateInfo, error) {
	current := Version

	tag, notes, source, err := latestViaRedirect(ctx, client)
	if err != nil {
		tag, notes, source, err = latestViaAPI(ctx, client)
	}
	if err != nil {
		return UpdateInfo{Current: displayVersion()}, err
	}

	latest := normalizeVersion(tag)
	hasUpdate := compareVersions(latest, normalizeVersion(current)) > 0

	return UpdateInfo{
		Current:    displayVersion(),
		Latest:     tag,
		HasUpdate:  hasUpdate,
		ReleaseURL: releasePageURL(tag),
		Notes:      notes,
		Source:     source,
	}, nil
}

// latestViaRedirect —— 主路：吃 302 的 Location 头拿 tag，不占 API 限速。
// 跟随重定向是错的：跟到底就到 html 页面了，反而解析不出 tag。
func latestViaRedirect(ctx context.Context, client *http.Client) (tag, notes, source string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, updateRedirectURL, nil)
	if err != nil {
		return "", "", "", err
	}
	// 不跟随 3xx：下面要读 Location。
	noRedirect := *client
	noRedirect.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := noRedirect.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusMovedPermanently {
		return "", "", "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	loc := resp.Header.Get("Location")
	tag = tagFromLocation(loc)
	if tag == "" {
		return "", "", "", fmt.Errorf("重定向里没有版本号: %q", loc)
	}
	// 302 这条路拿不到更新说明，留空由界面自己处理。
	return tag, "", "redirect", nil
}

// latestViaAPI —— 副路：限速时给出可辨认的错误，让提示能对症。
func latestViaAPI(ctx context.Context, client *http.Client) (tag, notes, source string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, updateAPIURL, nil)
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	// 403 也可能被限速打出来；GitHub 限速时会带 X-RateLimit-Remaining: 0。
	if resp.StatusCode == http.StatusTooManyRequests ||
		(resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0") {
		return "", "", "", errUpdateRateLimited
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var payload struct {
		TagName string `json:"tag_name"`
		Body    string `json:"body"`
		Draft   bool   `json:"draft"`
		Pre     bool   `json:"prerelease"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", "", "", err
	}
	if payload.TagName == "" {
		return "", "", "", errors.New("响应里没有 tag_name")
	}
	// 预发布不当「最新版」推给用户。
	if payload.Draft || payload.Pre {
		return "", "", "", errors.New("最新发布是预发布版")
	}
	return payload.TagName, strings.TrimSpace(payload.Body), "api", nil
}

// tagFromLocation 从 .../releases/tag/v1.2.3 里抠出 v1.2.3。
// 只认末段，避免把 https://github.com 里的片段误当版本。
func tagFromLocation(loc string) string {
	const marker = "/releases/tag/"
	i := strings.LastIndex(loc, marker)
	if i < 0 {
		return ""
	}
	tag := strings.TrimSpace(loc[i+len(marker):])
	if tag == "" || strings.ContainsAny(tag, "/?#") {
		return ""
	}
	return tag
}

// releasePageURL —— 下载页地址由 tag 构造，不必等 API 给 browser_download_url。
// 这样即使只走 302 那条路，界面也能给出可点的链接。
func releasePageURL(tag string) string {
	return strings.TrimSuffix(updateRedirectURL, "/releases/latest") + "/releases/tag/" + tag
}

// normalizeVersion 把可比的数字部分抠出来：
//
//	"v1.2.3"      → "1.2.3"
//	"1.2.3-dirty" → "1.2.3"   （本地未提交改动的构建）
//	"v1.2.3-rc1"  → "1.2.3"   （预发布按同版本算，不误报更新）
//	"dev"         → ""        （未注入版本，无从比较）
//
// 不引入 semver 依赖：这里只要「最新版是不是比当前新」，三段数字比大小够用。
func normalizeVersion(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "V")
	if i := strings.IndexByte(s, '-'); i >= 0 {
		s = s[:i]
	}
	if i := strings.IndexByte(s, '+'); i >= 0 {
		s = s[:i]
	}
	for _, part := range strings.Split(s, ".") {
		if part == "" {
			return ""
		}
		if _, err := strconv.Atoi(part); err != nil {
			return ""
		}
	}
	return s
}

// compareVersions 比较两段已 normalize 的版本号：>0 表示 a 更新，<0 表示 b 更新，0 相同。
// 段数不齐按 0 补（1.2 视作 1.2.0）。任一方不是合法版本号时返回 0，即「当作没更新」，
// 宁可少提示一次，也不要在 dev 构建上弹一堆「请更新」。
func compareVersions(a, b string) int {
	if a == "" || b == "" {
		return 0
	}
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		av, bv := 0, 0
		if i < len(as) {
			av, _ = strconv.Atoi(as[i])
		}
		if i < len(bs) {
			bv, _ = strconv.Atoi(bs[i])
		}
		if av != bv {
			return av - bv
		}
	}
	return 0
}

// updateCheckTimeout —— 检查更新不能卡住启动或界面。8 秒拿不到就当失败。
const updateCheckTimeout = 8 * time.Second
