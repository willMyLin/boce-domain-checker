package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	boceBaseURL   = "https://api.boce.com/v3/"
	boceBatchSize = 20
)

func queryQQBatch(client *http.Client, apiKey string, hosts []string) (map[string]int, error) {
	return queryBlockBatch(client, apiKey, hosts, "task/create/qq")
}

func queryPollutionBatch(client *http.Client, apiKey string, hosts []string) (map[string]int, error) {
	return queryBlockBatch(client, apiKey, hosts, "task/create/pollute")
}

func queryWechatBatch(client *http.Client, apiKey string, hosts []string) (map[string]int, error) {
	return queryBlockBatch(client, apiKey, hosts, "task/create/wechat")
}

func queryBlacklistBatch(client *http.Client, apiKey string, hosts []string) (map[string]int, error) {
	return queryBlockBatch(client, apiKey, hosts, "task/create/blacklist")
}

func queryWallBatch(client *http.Client, apiKey string, hosts []string) (map[string]int, error) {
	return queryBlockBatch(client, apiKey, hosts, "task/create/wall")
}

func queryICP(client *http.Client, apiKey string, host string) (boceICPInfo, error) {
	endpoint, err := url.Parse(boceBaseURL + "task/create/icp")
	if err != nil {
		return boceICPInfo{}, err
	}

	query := endpoint.Query()
	setBoceTaskQuery(query, apiKey, host)
	endpoint.RawQuery = query.Encode()

	resp, err := client.Get(endpoint.String())
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

func queryBlockBatch(client *http.Client, apiKey string, hosts []string, taskPath string) (map[string]int, error) {
	endpoint, err := url.Parse(boceBaseURL + taskPath)
	if err != nil {
		return nil, err
	}

	query := endpoint.Query()
	setBoceTaskQuery(query, apiKey, strings.Join(hosts, ","))
	endpoint.RawQuery = query.Encode()

	resp, err := client.Get(endpoint.String())
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
