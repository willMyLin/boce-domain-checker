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

	parsed := parseTargetsText(string(content))
	if len(parsed.Valid) == 0 {
		return ImportResponse{
			Message: "未从TXT中解析到有效检测目标",
		}
	}
	message := fmt.Sprintf("已导入 %d 条检测目标", len(parsed.Valid))
	if len(parsed.Invalid) > 0 {
		message = fmt.Sprintf("%s，另有 %d 条格式错误已跳过", message, len(parsed.Invalid))
	}
	return ImportResponse{
		Targets: parsed.Valid,
		Message: message,
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

	targets := normalizeTargets(req.Targets).Valid
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

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	expectedRows := expectedDetectRows(req, len(targets))
	a.emitDetectStart(expectedRows)
	emitRow := a.emitDetectRow
	rows := make([]DetectRow, 0, len(targets)*6)
	if req.EnablePollution {
		rows = append(rows, runPollutionDetection(client, req.APIKey, targets, concurrency, len(rows), emitRow)...)
	}
	if req.EnableHijack {
		rows = append(rows, runQQDetection(client, req.APIKey, targets, concurrency, len(rows), emitRow)...)
	}
	if req.EnableWechat {
		rows = append(rows, runWechatDetection(client, req.APIKey, targets, concurrency, len(rows), emitRow)...)
	}
	if req.EnableICP {
		rows = append(rows, runICPDetection(client, req.APIKey, targets, concurrency, len(rows), emitRow)...)
	}
	if req.EnableBlacklist {
		rows = append(rows, runBlacklistDetection(client, req.APIKey, targets, concurrency, len(rows), emitRow)...)
	}
	if req.EnableWall {
		rows = append(rows, runWallDetection(client, req.APIKey, targets, concurrency, len(rows), emitRow)...)
	}

	summary := summarizeRows(rows)
	message := detectionMessage(req.EnablePollution, req.EnableHijack, req.EnableWechat, req.EnableICP, req.EnableBlacklist, req.EnableWall, concurrency, timeout)

	return DetectResponse{
		Rows:       rows,
		Summary:    summary,
		Progress:   100,
		ExportPath: "",
		Message:    message,
	}
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

// ValidateTargets 归一化并严格校验检测目标。域名校验已收口到 Go 侧：
// 前端只负责展示，最终判定以本方法的返回为准（TXT 导入也走同一套规则，
// 不再有「导入侧宽松、提交侧严格」的旁路）。
func (a *App) ValidateTargets(text string) ValidateTargetsResponse {
	parsed := parseTargetsText(text)
	return ValidateTargetsResponse{
		Valid:   parsed.Valid,
		Invalid: parsed.Invalid,
		Message: targetValidationMessage(text, parsed),
	}
}
