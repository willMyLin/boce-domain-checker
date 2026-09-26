package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const (
	StatusPollution     = "污染"
	StatusIntercepted   = "拦截"
	StatusUnfiled       = "未备案"
	StatusBlacklisted   = "黑名单"
	StatusWalled        = "被墙"
	StatusSuspectWalled = "疑似被墙"
	StatusBadDomain     = "域名格式错误"
	StatusNormal        = "正常"
	StatusFiled         = "已备案"
	StatusUnregistered  = "未注册"
	StatusFailed        = "失败"
	StatusUnknown       = "未知"

	TypePollution = "污染"
	TypeQQ        = "QQ拦截"
	TypeWechat    = "微信拦截"
	TypeICP       = "备案查询"
	TypeBlacklist = "备案黑名单"
	TypeWall      = "被墙检测"
)

func polluteStatusText(statusCode int, err error) string {
	if err != nil {
		return StatusFailed
	}

	switch statusCode {
	case -1:
		return StatusUnregistered
	case 0:
		return StatusNormal
	case 1:
		return StatusPollution
	default:
		return StatusUnknown
	}
}

func qqStatusText(statusCode int, err error) string {
	return blockStatusText(statusCode, err)
}

func wechatStatusText(statusCode int, err error) string {
	return blockStatusText(statusCode, err)
}

func icpStatusText(statusCode int, err error) string {
	if err != nil {
		return StatusFailed
	}

	switch statusCode {
	case 1:
		return StatusFiled
	case 0:
		return StatusUnfiled
	default:
		return StatusUnknown
	}
}

func blockStatusText(statusCode int, err error) string {
	if err != nil {
		return StatusFailed
	}

	switch statusCode {
	case 1:
		return StatusNormal
	case 2:
		return StatusIntercepted
	case 3:
		return StatusFailed
	default:
		return StatusUnknown
	}
}

func blacklistStatusText(statusCode int, err error) string {
	if err != nil {
		return StatusFailed
	}

	switch statusCode {
	case 1:
		return StatusNormal
	case 2:
		return StatusBlacklisted
	case 3:
		return StatusFailed
	default:
		return StatusUnknown
	}
}

func wallStatusText(statusCode int, err error) string {
	if err != nil {
		return StatusFailed
	}

	switch statusCode {
	case 0:
		return StatusNormal
	case 1:
		return StatusWalled
	case 2:
		return StatusSuspectWalled
	case 3:
		return StatusFailed
	case 4:
		return StatusBadDomain
	default:
		return StatusUnknown
	}
}

func detectionMessage(enablePollution bool, enableQQ bool, enableWechat bool, enableICP bool, enableBlacklist bool, enableWall bool, concurrency int, timeout int) string {
	checks := make([]string, 0, 6)
	if enablePollution {
		checks = append(checks, "污染检测")
	}
	if enableQQ {
		checks = append(checks, "QQ拦截检测")
	}
	if enableWechat {
		checks = append(checks, "微信拦截检测")
	}
	if enableICP {
		checks = append(checks, TypeICP)
	}
	if enableBlacklist {
		checks = append(checks, "备案黑名单检测")
	}
	if enableWall {
		checks = append(checks, TypeWall)
	}

	return fmt.Sprintf("%s完成，并发 %d，超时 %d 秒", strings.Join(checks, "、"), concurrency, timeout)
}

func errorRemark(err error) string {
	if err == nil {
		return ""
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return fmt.Sprintf("请求失败: %v", urlErr.Err)
	}
	return err.Error()
}

func expectedDetectRows(req DetectRequest, targetCount int) int {
	checks := 0
	if req.EnablePollution {
		checks++
	}
	if req.EnableHijack {
		checks++
	}
	if req.EnableWechat {
		checks++
	}
	if req.EnableICP {
		checks++
	}
	if req.EnableBlacklist {
		checks++
	}
	if req.EnableWall {
		checks++
	}
	return checks * targetCount
}

func summarizeRows(rows []DetectRow) DetectSummary {
	summary := DetectSummary{
		Total:   len(rows),
		Checked: len(rows),
	}

	for _, row := range rows {
		switch row.Status {
		case StatusPollution, StatusIntercepted, StatusUnfiled, StatusBlacklisted, StatusWalled, StatusSuspectWalled, StatusBadDomain:
			summary.Pollution++
		case StatusNormal, StatusFiled:
			summary.Normal++
		case StatusUnregistered:
			summary.Unregistered++
		case StatusFailed:
			summary.Failed++
		}
	}

	return summary
}
