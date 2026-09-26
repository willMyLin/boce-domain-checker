package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// useTestBoceBaseURL 把 BOCE API 根地址指到 httptest 假服务器，测试结束自动还原。
func useTestBoceBaseURL(t *testing.T, serverURL string) {
	t.Helper()
	original := boceBaseURL
	boceBaseURL = serverURL + "/"
	t.Cleanup(func() {
		boceBaseURL = original
	})
}

func TestBeginDetectionCancel(t *testing.T) {
	app := NewApp()
	ctx, finish := app.beginDetection()

	if ctx.Err() != nil {
		t.Fatal("beginDetection() 返回的 ctx 不应已经取消")
	}
	if !app.CancelDetection() {
		t.Fatal("检测进行中 CancelDetection() 应返回 true")
	}
	if ctx.Err() == nil {
		t.Fatal("CancelDetection() 之后 ctx.Err() 应非 nil")
	}

	finish()
	if ctx.Err() == nil {
		t.Fatal("finish() 之后 ctx.Err() 应非 nil")
	}
	if app.CancelDetection() {
		t.Fatal("finish() 之后 CancelDetection() 应返回 false")
	}
}

func TestCancelDetectionNoopWhenIdle(t *testing.T) {
	app := NewApp()
	if app.CancelDetection() {
		t.Fatal("空闲时 CancelDetection() 应返回 false")
	}
}

func TestQueryBlockBatchRespectsCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(10 * time.Second):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"error_code":0,"data":{"a.com":1}}`))
		case <-r.Context().Done():
		}
	}))
	defer server.Close()
	useTestBoceBaseURL(t, server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	result, err := queryBlockBatch(ctx, &http.Client{}, "test-key", []string{"a.com"}, "task/create/qq")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("取消后 queryBlockBatch() 应报错，实际返回 %#v", result)
	}
	if elapsed >= time.Second {
		t.Fatalf("取消后 queryBlockBatch() 耗时 %v，应小于 1s", elapsed)
	}
}

func TestQueryBlockBatchHappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error_code":0,"data":{"a.com":1}}`))
	}))
	defer server.Close()
	useTestBoceBaseURL(t, server.URL)

	result, err := queryBlockBatch(context.Background(), &http.Client{}, "test-key", []string{"a.com"}, "task/create/qq")
	if err != nil {
		t.Fatalf("queryBlockBatch() error = %v", err)
	}
	if got := result["a.com"]; got != 1 {
		t.Fatalf("queryBlockBatch() = %#v, want map hit a.com -> 1", result)
	}
}
