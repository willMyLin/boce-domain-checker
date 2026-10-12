package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context

	detectMu     sync.Mutex
	detectCancel context.CancelFunc
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// beginDetection 注册一次可取消的检测，返回的 finish 必须在 StartDetection 里 defer，
// 否则并发第二次检测会泄漏上一次的 cancel 函数。
func (a *App) beginDetection() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	a.detectMu.Lock()
	a.detectCancel = cancel
	a.detectMu.Unlock()
	return ctx, func() {
		a.detectMu.Lock()
		a.detectCancel = nil
		a.detectMu.Unlock()
		cancel()
	}
}

func (a *App) emitDetectStart(total int) {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, "detect:start", DetectStartEvent{Total: total})
}

func (a *App) emitDetectRow(row DetectRow) {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, "detect:row", row)
}

// ImportTXT opens a file picker and parses detection targets from a TXT file.
func (a *App) ImportTXT() ImportResponse {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择检测目标TXT文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "TXT 文件 (*.txt)", Pattern: "*.txt"},
			{DisplayName: "所有文件 (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return ImportResponse{
			Message: fmt.Sprintf("导入失败: %v", err),
		}
	}
	if path == "" {
		return ImportResponse{
			Message:  "已取消导入",
			Canceled: true,
		}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return ImportResponse{
			Message: fmt.Sprintf("读取文件失败: %v", err),
		}
	}

	targets := parseTargetsText(string(content))
	if len(targets) == 0 {
		return ImportResponse{
			Message: "未从TXT中解析到有效检测目标",
		}
	}
	return ImportResponse{
		Targets: targets,
		Message: fmt.Sprintf("已导入 %d 条检测目标", len(targets)),
	}
}

// StartDetection runs the selected BOCE checks. An empty or placeholder key
// keeps returning mock data for UI testing.
func (a *App) StartDetection(req DetectRequest) DetectResponse {
	req.APIKey = strings.TrimSpace(req.APIKey)
	if !req.EnablePollution && !req.EnableHijack && !req.EnableWechat && !req.EnableICP && !req.EnableBlacklist && !req.EnableWall {
		return DetectResponse{
			Summary:  DetectSummary{},
			Progress: 100,
			Message:  "请至少选择一个检测类型",
		}
	}

	if req.APIKey == "" || strings.Contains(req.APIKey, "*") {
		response := makeMockDetectionResponse(
			fmt.Sprintf("未填写有效 API Key，返回测试数据，并发 %d，超时 %d 秒", req.Concurrency, req.TimeoutSeconds),
			req.EnablePollution,
			req.EnableHijack,
			req.EnableWechat,
			req.EnableICP,
			req.EnableBlacklist,
			req.EnableWall,
		)
		a.emitDetectStart(len(response.Rows))
		for _, row := range response.Rows {
			a.emitDetectRow(row)
		}
		return response
	}

	targets := normalizeTargets(req.Targets)
	if len(targets) == 0 {
		targets = mockTargets()
	}

	timeout := req.TimeoutSeconds
	if timeout <= 0 {
		timeout = 60
	}

	concurrency := req.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	if concurrency > len(targets) {
		concurrency = len(targets)
	}

	if concurrency > 20 {
		concurrency = 20
	}

	ctx, finish := a.beginDetection()
	defer finish()

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	expectedRows := expectedDetectRows(req, len(targets))
	a.emitDetectStart(expectedRows)
	emitRow := a.emitDetectRow
	rows := make([]DetectRow, 0, len(targets)*6)
	if req.EnablePollution {
		rows = append(rows, runPollutionDetection(ctx, client, req.APIKey, targets, concurrency, len(rows), emitRow)...)
	}
	if req.EnableHijack {
		rows = append(rows, runQQDetection(ctx, client, req.APIKey, targets, concurrency, len(rows), emitRow)...)
	}
	if req.EnableWechat {
		rows = append(rows, runWechatDetection(ctx, client, req.APIKey, targets, concurrency, len(rows), emitRow)...)
	}
	if req.EnableICP {
		rows = append(rows, runICPDetection(ctx, client, req.APIKey, targets, concurrency, len(rows), emitRow)...)
	}
	if req.EnableBlacklist {
		rows = append(rows, runBlacklistDetection(ctx, client, req.APIKey, targets, concurrency, len(rows), emitRow)...)
	}
	if req.EnableWall {
		rows = append(rows, runWallDetection(ctx, client, req.APIKey, targets, concurrency, len(rows), emitRow)...)
	}

	summary := summarizeRows(rows)
	message := detectionMessage(req.EnablePollution, req.EnableHijack, req.EnableWechat, req.EnableICP, req.EnableBlacklist, req.EnableWall, concurrency, timeout)
	if ctx.Err() != nil {
		message = "已中止，" + message
	}

	return DetectResponse{
		Rows:       rows,
		Summary:    summary,
		Progress:   100,
		ExportPath: "C:/Users/xjw/Desktop/全部污染.xlsx",
		Message:    message,
	}
}

// CancelDetection aborts a running StartDetection call.
func (a *App) CancelDetection() bool {
	a.detectMu.Lock()
	cancel := a.detectCancel
	a.detectMu.Unlock()
	if cancel == nil {
		return false
	}
	cancel()
	return true
}

// ExportPollutionExcel exports the rows currently shown in the frontend.
func (a *App) ExportPollutionExcel(rows []DetectRow) ExportResponse {
	if len(rows) == 0 {
		return ExportResponse{
			Message: "当前列表没有可导出的数据",
		}
	}

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "导出当前列表数据",
		DefaultFilename: defaultExportFilename(rows),
		Filters: []runtime.FileFilter{
			{DisplayName: "Excel 工作簿 (*.xlsx)", Pattern: "*.xlsx"},
		},
	})
	if err != nil {
		return ExportResponse{
			Message: fmt.Sprintf("导出失败: %v", err),
		}
	}
	if path == "" {
		return ExportResponse{
			Message: "已取消导出",
		}
	}
	if filepath.Ext(path) == "" {
		path += ".xlsx"
	}

	if err := writeRowsXLSX(path, rows); err != nil {
		return ExportResponse{
			Message: fmt.Sprintf("导出失败: %v", err),
		}
	}

	return ExportResponse{
		Path:    path,
		Message: fmt.Sprintf("已导出 %d 条当前列表数据", len(rows)),
	}
}
