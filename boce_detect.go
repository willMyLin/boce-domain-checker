package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

func runPollutionDetection(client *http.Client, apiKey string, targets []string, concurrency int, idOffset int, emit func(DetectRow)) []DetectRow {
	return runBlockDetection(client, apiKey, targets, concurrency, idOffset, TypePollution, queryPollutionBatch, polluteStatusText, emit)
}

func runQQDetection(client *http.Client, apiKey string, targets []string, concurrency int, idOffset int, emit func(DetectRow)) []DetectRow {
	return runBlockDetection(client, apiKey, targets, concurrency, idOffset, TypeQQ, queryQQBatch, blockStatusText, emit)
}

func runWechatDetection(client *http.Client, apiKey string, targets []string, concurrency int, idOffset int, emit func(DetectRow)) []DetectRow {
	return runBlockDetection(client, apiKey, targets, concurrency, idOffset, TypeWechat, queryWechatBatch, blockStatusText, emit)
}

func runBlacklistDetection(client *http.Client, apiKey string, targets []string, concurrency int, idOffset int, emit func(DetectRow)) []DetectRow {
	return runBlockDetection(client, apiKey, targets, concurrency, idOffset, TypeBlacklist, queryBlacklistBatch, blacklistStatusText, emit)
}

func runWallDetection(client *http.Client, apiKey string, targets []string, concurrency int, idOffset int, emit func(DetectRow)) []DetectRow {
	return runBlockDetection(client, apiKey, targets, concurrency, idOffset, TypeWall, queryWallBatch, wallStatusText, emit)
}

func runICPDetection(client *http.Client, apiKey string, targets []string, concurrency int, idOffset int, emit func(DetectRow)) []DetectRow {
	rows := make([]DetectRow, len(targets))
	jobs := make(chan int)
	var wg sync.WaitGroup

	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				target := targets[idx]
				info, err := queryICP(client, apiKey, target)
				domain := firstNonEmpty(info.Domain, target)
				row := DetectRow{
					ID:          idOffset + idx + 1,
					Type:        TypeICP,
					Target:      target,
					Status:      icpStatusText(info.Status, err),
					CheckedAt:   time.Now().Format("2006-01-02 15:04:05"),
					ErrorRemark: errorRemark(err),
					Domain:      domain,
					BeianCode:   firstNonEmpty(info.SiteBeianCode, info.CompanyBeianCode),
					SiteName:    info.Name,
				}
				rows[idx] = row
				if emit != nil {
					emit(row)
				}
			}
		}()
	}

	for idx := range targets {
		jobs <- idx
	}
	close(jobs)
	wg.Wait()

	return rows
}

func runBlockDetection(client *http.Client, apiKey string, targets []string, concurrency int, idOffset int, rowType string, query func(*http.Client, string, []string) (map[string]int, error), statusText func(int, error) string, emit func(DetectRow)) []DetectRow {
	rows := make([]DetectRow, len(targets))
	batches := make([][]int, 0, (len(targets)+boceBatchSize-1)/boceBatchSize)
	for start := 0; start < len(targets); start += boceBatchSize {
		end := start + boceBatchSize
		if end > len(targets) {
			end = len(targets)
		}

		indexes := make([]int, 0, end-start)
		for idx := start; idx < end; idx++ {
			indexes = append(indexes, idx)
		}
		batches = append(batches, indexes)
	}

	if concurrency > len(batches) {
		concurrency = len(batches)
	}

	jobs := make(chan []int)
	var wg sync.WaitGroup
	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for indexes := range jobs {
				hosts := make([]string, 0, len(indexes))
				for _, idx := range indexes {
					hosts = append(hosts, targets[idx])
				}

				result, err := query(client, apiKey, hosts)
				checkedAt := time.Now().Format("2006-01-02 15:04:05")
				for _, idx := range indexes {
					target := targets[idx]
					statusCode, statusErr := lookupHostStatus(result, target)
					if err != nil {
						statusErr = err
					} else if statusCode == 3 {
						statusErr = errors.New("检测失败")
					}
					row := DetectRow{
						ID:          idOffset + idx + 1,
						Type:        rowType,
						Target:      target,
						Status:      statusText(statusCode, statusErr),
						CheckedAt:   checkedAt,
						ErrorRemark: errorRemark(statusErr),
					}
					rows[idx] = row
					if emit != nil {
						emit(row)
					}
				}
			}
		}()
	}

	for _, batch := range batches {
		jobs <- batch
	}
	close(jobs)
	wg.Wait()

	return rows
}

func lookupHostStatus(result map[string]int, host string) (int, error) {
	if status, ok := result[host]; ok {
		return status, nil
	}
	if status, ok := result["www."+host]; ok {
		return status, nil
	}
	trimmed := strings.TrimPrefix(host, "www.")
	if status, ok := result[trimmed]; ok {
		return status, nil
	}

	return 0, fmt.Errorf("响应缺少 %s 的检测结果", host)
}
