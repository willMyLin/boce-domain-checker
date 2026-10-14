package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const boceBatchSize = 20

// boceBaseURL 是 BOCE API 根地址。做成变量而不是常量，是为了让 boce_client_test.go
// 能把请求指到 net/http/httptest 假服务器；生产取值不变。
var boceBaseURL = "https://api.boce.com/v3/"

func queryPollution(ctx context.Context, client *http.Client, apiKey string, host string) (int, error) {
	endpoint, err := url.Parse(boceBaseURL + "task/create/pollute")
	if err != nil {
		return 0, err
	}

	query := endpoint.Query()
	setBoceTaskQuery(query, apiKey, host)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var result bocePolluteResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	if result.ErrorCode != 0 {
		if result.Error != "" {
			return 0, errors.New(result.Error)
		}
		return 0, fmt.Errorf("BOCE error_code %d", result.ErrorCode)
	}

	if status, ok := result.Data[host]; ok {
		return status, nil
	}
	for _, status := range result.Data {
		return status, nil
	}

	return 0, fmt.Errorf("响应缺少检测结果")
}

func queryQQBatch(ctx context.Context, client *http.Client, apiKey string, hosts []string) (map[string]int, error) {
	return queryBlockBatch(ctx, client, apiKey, hosts, "task/create/qq")
}

func queryPollutionBatch(ctx context.Context, client *http.Client, apiKey string, hosts []string) (map[string]int, error) {
	return queryBlockBatch(ctx, client, apiKey, hosts, "task/create/pollute")
}

func queryWechatBatch(ctx context.Context, client *http.Client, apiKey string, hosts []string) (map[string]int, error) {
	return queryBlockBatch(ctx, client, apiKey, hosts, "task/create/wechat")
}

func queryBlacklistBatch(ctx context.Context, client *http.Client, apiKey string, hosts []string) (map[string]int, error) {
	return queryBlockBatch(ctx, client, apiKey, hosts, "task/create/blacklist")
}

func queryWallBatch(ctx context.Context, client *http.Client, apiKey string, hosts []string) (map[string]int, error) {
	return queryBlockBatch(ctx, client, apiKey, hosts, "task/create/wall")
}

func queryICP(ctx context.Context, client *http.Client, apiKey string, host string) (boceICPInfo, error) {
	endpoint, err := url.Parse(boceBaseURL + "task/create/icp")
	if err != nil {
		return boceICPInfo{}, err
	}

	query := endpoint.Query()
	setBoceTaskQuery(query, apiKey, host)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return boceICPInfo{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return boceICPInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return boceICPInfo{}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var result boceICPResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return boceICPInfo{}, err
	}
	if result.ErrorCode != 0 {
		if result.Error != "" {
			return boceICPInfo{}, errors.New(result.Error)
		}
		return boceICPInfo{}, fmt.Errorf("BOCE error_code %d", result.ErrorCode)
	}

	if info, ok := result.Data[host]; ok {
		return info, nil
	}
	for _, info := range result.Data {
		return info, nil
	}

	return boceICPInfo{}, fmt.Errorf("响应缺少备案结果")
}

func queryBlockBatch(ctx context.Context, client *http.Client, apiKey string, hosts []string, taskPath string) (map[string]int, error) {
	endpoint, err := url.Parse(boceBaseURL + taskPath)
	if err != nil {
		return nil, err
	}

	query := endpoint.Query()
	setBoceTaskQuery(query, apiKey, strings.Join(hosts, ","))
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var result boceQQResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.ErrorCode != 0 {
		if result.Error != "" {
			return nil, errors.New(result.Error)
		}
		return nil, fmt.Errorf("BOCE error_code %d", result.ErrorCode)
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("响应缺少检测结果")
	}

	return result.Data, nil
}

func setBoceTaskQuery(query url.Values, apiKey string, host string) {
	query.Set("key", apiKey)
	query.Set("host", host)
}
