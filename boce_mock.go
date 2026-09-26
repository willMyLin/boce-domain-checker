package main

import (
	"fmt"
	"strings"
	"time"
)

func makeMockDetectionResponse(message string, enablePollution bool, enableQQ bool, enableWechat bool, enableICP bool, enableBlacklist bool, enableWall bool) DetectResponse {
	rows := makeMockRows(enablePollution, enableQQ, enableWechat, enableICP, enableBlacklist, enableWall)
	if enablePollution && !enableQQ && !enableWechat && !enableICP && !enableBlacklist && !enableWall {
		total := 93874
		return DetectResponse{
			Rows: rows,
			Summary: DetectSummary{
				Total:        total,
				Checked:      total,
				Pollution:    1498,
				Normal:       92155,
				Unregistered: 203,
				Failed:       18,
			},
			Progress:   100,
			ExportPath: "",
			Message:    message,
		}
	}

	summary := summarizeRows(rows)
	return DetectResponse{
		Rows:       rows,
		Summary:    summary,
		Progress:   100,
		ExportPath: "",
		Message:    message,
	}
}

func makeMockRows(enablePollution bool, enableQQ bool, enableWechat bool, enableICP bool, enableBlacklist bool, enableWall bool) []DetectRow {
	targets := mockTargets()
	base := time.Date(2026, 6, 9, 8, 39, 40, 0, time.Local)
	rows := make([]DetectRow, 0, len(targets)*6)

	if enablePollution {
		for idx, target := range targets {
			rows = append(rows, DetectRow{
				ID:        len(rows) + 1,
				Type:      TypePollution,
				Target:    target,
				Status:    StatusNormal,
				CheckedAt: base.Add(time.Duration(idx/2) * time.Second).Format("2006-01-02 15:04:05"),
			})
		}
	}

	if enableQQ {
		for idx, target := range targets {
			status := StatusNormal
			if idx%9 == 0 {
				status = StatusIntercepted
			}
			rows = append(rows, DetectRow{
				ID:        len(rows) + 1,
				Type:      TypeQQ,
				Target:    target,
				Status:    status,
				CheckedAt: base.Add(time.Duration(idx/2) * time.Second).Format("2006-01-02 15:04:05"),
			})
		}
	}

	if enableWechat {
		for idx, target := range targets {
			status := StatusNormal
			if idx%7 == 0 {
				status = StatusIntercepted
			}
			rows = append(rows, DetectRow{
				ID:        len(rows) + 1,
				Type:      TypeWechat,
				Target:    target,
				Status:    status,
				CheckedAt: base.Add(time.Duration(idx/2) * time.Second).Format("2006-01-02 15:04:05"),
			})
		}
	}

	if enableICP {
		for idx, target := range targets {
			status := StatusFiled
			if idx%8 == 0 {
				status = StatusUnfiled
			}
			rows = append(rows, DetectRow{
				ID:        len(rows) + 1,
				Type:      TypeICP,
				Target:    target,
				Status:    status,
				CheckedAt: base.Add(time.Duration(idx/2) * time.Second).Format("2006-01-02 15:04:05"),
				Domain:    target,
				BeianCode: fmt.Sprintf("测试ICP备案号-%d", idx+1),
				SiteName:  fmt.Sprintf("测试网站%d", idx+1),
			})
		}
	}

	if enableBlacklist {
		for idx, target := range targets {
			status := StatusNormal
			if idx%10 == 0 {
				status = StatusBlacklisted
			}
			rows = append(rows, DetectRow{
				ID:        len(rows) + 1,
				Type:      TypeBlacklist,
				Target:    target,
				Status:    status,
				CheckedAt: base.Add(time.Duration(idx/2) * time.Second).Format("2006-01-02 15:04:05"),
			})
		}
	}

	if enableWall {
		for idx, target := range targets {
			status := StatusNormal
			if idx%11 == 0 {
				status = StatusWalled
			} else if idx%7 == 0 {
				status = StatusSuspectWalled
			}
			rows = append(rows, DetectRow{
				ID:        len(rows) + 1,
				Type:      TypeWall,
				Target:    target,
				Status:    status,
				CheckedAt: base.Add(time.Duration(idx/2) * time.Second).Format("2006-01-02 15:04:05"),
			})
		}
	}

	return rows
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func mockTargets() []string {
	return []string{
		"80823456.com",
		"8088w.com",
		"80906.vip",
		"8090xiaoyaoju.com",
		"80969t.com",
		"8099r.com",
		"80gu.com",
		"80suncity.com",
		"80ych.top",
		"810180.com",
		"810461.com",
		"810msc.com",
		"8111777.com",
		"811k3.com",
		"8156d.com",
		"815749.org",
		"816002.com",
		"817890.com",
		"8181138.com",
		"818428.com",
		"81850zq.com",
		"81852.cc",
		"818737.cc",
		"8188q.com",
		"818988c.com",
		"818yule.net",
		"81966.com",
		"81999.cc",
		"820003.com",
		"82023.net",
		"821138.com",
		"82226.vip",
	}
}
