package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNormalizeVersion(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"v1.2.3", "1.2.3"},
		{"V1.2.3", "1.2.3"},
		{"1.2.3", "1.2.3"},
		{"v1.2.3-dirty", "1.2.3"},
		{"v1.2.3-rc1", "1.2.3"},
		{"1.2.3+build", "1.2.3"},
		{"dev", ""},
		{"", ""},
		{"1.2", "1.2"},
		{"1.2.x", ""},
		{"本地构建", ""},
	}
	for _, c := range cases {
		if got := normalizeVersion(c.in); got != c.want {
			t.Errorf("normalizeVersion(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int // sign
	}{
		{"1.2.3", "1.2.3", 0},
		{"1.2.4", "1.2.3", 1},
		{"1.2.3", "1.2.4", -1},
		{"2.0.0", "1.9.9", 1},
		{"1.2", "1.2.0", 0},
		{"1.10.0", "1.9.0", 1},
		// 任一方无版本号 → 0，宁可不提示
		{"", "1.2.3", 0},
		{"1.2.3", "", 0},
	}
	for _, c := range cases {
		got := compareVersions(c.a, c.b)
		sign := func(n int) int {
			switch {
			case n > 0:
				return 1
			case n < 0:
				return -1
			}
			return 0
		}
		if sign(got) != c.want {
			t.Errorf("compareVersions(%q, %q) = %d, want sign %d", c.a, c.b, got, c.want)
		}
	}
}

func TestTagFromLocation(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"https://github.com/willMyLin/boce-domain-checker/releases/tag/v1.2.3", "v1.2.3"},
		{"https://github.com/x/y/releases/tag/v2.0.0-rc1", "v2.0.0-rc1"},
		{"https://github.com/x/y/releases/latest", ""},
		{"", ""},
		{"/releases/tag/", ""},
		// 不能把路径里夹的斜杠当版本
		{"https://github.com/x/y/releases/tag/v1.0.0/extra", ""},
	}
	for _, c := range cases {
		if got := tagFromLocation(c.in); got != c.want {
			t.Errorf("tagFromLocation(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// withUpdateURLs 把两条取数路径指到假服务器，测完还原。
func withUpdateURLs(t *testing.T, redirectURL, apiURL string) {
	t.Helper()
	oldR, oldA := updateRedirectURL, updateAPIURL
	updateRedirectURL, updateAPIURL = redirectURL, apiURL
	t.Cleanup(func() { updateRedirectURL, updateAPIURL = oldR, oldA })
}

func TestLatestViaRedirect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://github.com/willMyLin/boce-domain-checker/releases/tag/v9.9.9", http.StatusFound)
	}))
	defer srv.Close()
	withUpdateURLs(t, srv.URL, srv.URL)

	tag, notes, source, err := latestViaRedirect(context.Background(), srv.Client())
	if err != nil {
		t.Fatalf("latestViaRedirect: %v", err)
	}
	if tag != "v9.9.9" {
		t.Errorf("tag = %q, want v9.9.9", tag)
	}
	if source != "redirect" {
		t.Errorf("source = %q, want redirect", source)
	}
	if notes != "" {
		t.Errorf("302 这条路不该有更新说明，got %q", notes)
	}
}

func TestLatestViaRedirectIgnoresFollowedBody(t *testing.T) {
	// 服务端 200（不是 302）时必须报错，不能把 HTML 当版本。
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	withUpdateURLs(t, srv.URL, srv.URL)

	if _, _, _, err := latestViaRedirect(context.Background(), srv.Client()); err == nil {
		t.Fatal("200 应当报错，而不是当作成功")
	}
}

func TestLatestViaAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v3.1.0","body":"修复若干问题","draft":false,"prerelease":false}`))
	}))
	defer srv.Close()
	withUpdateURLs(t, "http://127.0.0.1:1", srv.URL)

	tag, notes, source, err := latestViaAPI(context.Background(), srv.Client())
	if err != nil {
		t.Fatalf("latestViaAPI: %v", err)
	}
	if tag != "v3.1.0" {
		t.Errorf("tag = %q, want v3.1.0", tag)
	}
	if notes != "修复若干问题" {
		t.Errorf("notes = %q", notes)
	}
	if source != "api" {
		t.Errorf("source = %q, want api", source)
	}
}

func TestLatestViaAPIRejectsPrerelease(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v4.0.0-rc1","prerelease":true}`))
	}))
	defer srv.Close()
	withUpdateURLs(t, "http://127.0.0.1:1", srv.URL)

	if _, _, _, err := latestViaAPI(context.Background(), srv.Client()); err == nil {
		t.Fatal("预发布版不该当最新版推给用户")
	}
}

func TestLatestViaAPIDetectsRateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()
	withUpdateURLs(t, "http://127.0.0.1:1", srv.URL)

	_, _, _, err := latestViaAPI(context.Background(), srv.Client())
	if !errors.Is(err, errUpdateRateLimited) {
		t.Fatalf("err = %v, want errUpdateRateLimited", err)
	}
}

func TestLatestViaAPIDetectsTooManyRequests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()
	withUpdateURLs(t, "http://127.0.0.1:1", srv.URL)

	_, _, _, err := latestViaAPI(context.Background(), srv.Client())
	if !errors.Is(err, errUpdateRateLimited) {
		t.Fatalf("err = %v, want errUpdateRateLimited", err)
	}
}

// TestCheckUpdateFallsBackToAPI —— 主路 302 挂掉时必须自动切到 API，这就是「降级」。
func TestCheckUpdateFallsBackToAPI(t *testing.T) {
	var apiHit bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "redirect-down") {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		apiHit = true
		_, _ = w.Write([]byte(`{"tag_name":"v5.0.0","body":"说明","draft":false,"prerelease":false}`))
	}))
	defer srv.Close()
	withUpdateURLs(t, srv.URL+"/redirect-down", srv.URL)

	old := Version
	Version = "v1.0.0"
	t.Cleanup(func() { Version = old })

	info, err := checkUpdate(context.Background(), srv.Client())
	if err != nil {
		t.Fatalf("checkUpdate: %v", err)
	}
	if !apiHit {
		t.Error("主路失败后没有走 API 副路")
	}
	if info.Source != "api" {
		t.Errorf("source = %q, want api", info.Source)
	}
	if info.Latest != "v5.0.0" || !info.HasUpdate {
		t.Errorf("info = %+v, 想要 latest=v5.0.0 且 hasUpdate=true", info)
	}
}

// TestCheckUpdatePrefersRedirect —— 主路能用时不该去吃 API 的 60 次/小时额度。
func TestCheckUpdatePrefersRedirect(t *testing.T) {
	apiHit := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api" {
			apiHit = true
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "https://github.com/x/y/releases/tag/v2.2.2", http.StatusFound)
	}))
	defer srv.Close()
	withUpdateURLs(t, srv.URL, srv.URL+"/api")

	old := Version
	Version = "v2.2.2"
	t.Cleanup(func() { Version = old })

	info, err := checkUpdate(context.Background(), srv.Client())
	if err != nil {
		t.Fatalf("checkUpdate: %v", err)
	}
	if apiHit {
		t.Error("主路可用时不该打 API，会白白消耗限速额度")
	}
	if info.Source != "redirect" {
		t.Errorf("source = %q, want redirect", info.Source)
	}
	if info.HasUpdate {
		t.Errorf("同版本不该报有更新：%+v", info)
	}
}

// TestCheckUpdateBothPathsDown —— 两条路都挂要报错，由调用方决定静默还是提示。
func TestCheckUpdateBothPathsDown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	withUpdateURLs(t, srv.URL, srv.URL)

	old := Version
	Version = "v1.0.0"
	t.Cleanup(func() { Version = old })

	if _, err := checkUpdate(context.Background(), srv.Client()); err == nil {
		t.Fatal("两条路都失败时必须报错")
	}
}

// TestCheckUpdateDirtyBuild —— 本地未提交改动的构建（v1.1.1-dirty）应与 v1.1.1 视作同版本，
// 不要因为后缀就在开发机上弹「有新版本」。
func TestCheckUpdateDirtyBuild(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://github.com/x/y/releases/tag/v1.1.1", http.StatusFound)
	}))
	defer srv.Close()
	withUpdateURLs(t, srv.URL, srv.URL)

	old := Version
	Version = "v1.1.1-dirty"
	t.Cleanup(func() { Version = old })

	info, err := checkUpdate(context.Background(), srv.Client())
	if err != nil {
		t.Fatalf("checkUpdate: %v", err)
	}
	if info.HasUpdate {
		t.Errorf("-dirty 构建不该报有更新：%+v", info)
	}
}
