package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func GetRequestWithoutParams(url string) (string, error) {
	// 创建 GET 请求
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("failed to send GET request: %v", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"request failed with status code: %d",
			resp.StatusCode,
		)
	}

	return string(body), nil
}

// getRequestWithParams 发送带有查询参数的 GET 请求
func GetRequestWithParams(
	baseURL string,
	params map[string]string,
) (string, error) {
	// 解析 URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}

	// 添加查询参数
	query := parsedURL.Query()
	for key, value := range params {
		query.Set(key, value)
	}

	parsedURL.RawQuery = query.Encode()

	// 发送 GET 请求
	resp, err := http.Get(parsedURL.String())

	if err != nil {
		return "", fmt.Errorf("failed to send GET request: %v", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// 返回响应内容
	return string(body), nil
}

// postRequestWithoutParams 发送不带参数的 POST 请求
func PostRequestWithoutParams(url string) (string, error) {
	// 发送空的 POST 请求（body 为 nil）
	resp, err := http.Post(url, "application/json", nil) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("failed to send POST request: %v", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	return string(body), nil
}

// postRequestWithParams 发送带有 JSON 参数的 POST 请求
func PostRequestWithParams(
	url string,
	params map[string]string,
) (string, error) {
	// 将参数编码为 JSON
	jsonData, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("failed to encode params as JSON: %v", err)
	}

	// 发送带 JSON body 的 POST 请求
	resp, err := http.Post( //nolint:gosec
		url,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("failed to send POST request: %v", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	return string(body), nil
}

// getRequestWithParams 发送带有查询参数的 GET 请求，并根据目标类型解析响应
func GetRequestWithParamsAndObject(
	baseURL string,
	params map[string]string,
	responseTarget interface{}, // 用于接收解析后的响应
) (int, error) {
	// 解析 URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return 0, fmt.Errorf("invalid URL: %v", err)
	}

	// 添加查询参数
	query := parsedURL.Query()

	for key, value := range params {
		query.Set(key, value)
	}

	parsedURL.RawQuery = query.Encode()

	// 发送 GET 请求
	resp, err := http.Get(parsedURL.String())

	if err != nil {
		return 0, fmt.Errorf("failed to send GET request: %v", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, fmt.Errorf(
			"failed to read response body: %v",
			err,
		)
	}

	// 如果 responseTarget 为 nil，只返回状态码
	if responseTarget == nil {
		return resp.StatusCode, nil
	}

	// 如果 responseTarget 是 *string 类型，直接返回响应的文本
	if target, ok := responseTarget.(*string); ok {
		*target = string(body)
		return resp.StatusCode, nil
	}
	fmt.Println()
	// 如果response_target为其他类型指针，则decode json为该类型结构体
	err = json.Unmarshal(body, responseTarget)
	if err != nil {
		return resp.StatusCode, fmt.Errorf(
			"failed to decode response body: %v",
			err,
		)
	}
	// 返回响应状态码
	return resp.StatusCode, nil
}

// postRequestWithParams 发送带有 JSON 参数的 POST 请求，并根据目标类型解析响应
func PostRequestWithParamsAndObject(
	url string,
	params map[string]string,
	responseTarget interface{}, // 用于接收解析后的响应
) (int, error) {
	// 将参数编码为 JSON
	jsonData, err := json.Marshal(params)

	if err != nil {
		return 0, fmt.Errorf("failed to encode params as JSON: %v", err)
	}

	// 发送带 JSON body 的 POST 请求
	resp, err := http.Post( //nolint:gosec
		url,
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return 0, fmt.Errorf("failed to send POST request: %v", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, fmt.Errorf(
			"failed to read response body: %v",
			err,
		)
	}

	// 如果 responseTarget 为 nil，只返回状态码
	if responseTarget == nil {
		return resp.StatusCode, nil
	}

	// 如果 responseTarget 是 *string 类型，直接返回响应的文本
	if target, ok := responseTarget.(*string); ok {
		*target = string(body)

		return resp.StatusCode, nil
	}

	// 如果 responseTarget 是其他类型，尝试解码为 JSON 格式
	err = json.Unmarshal(body, responseTarget)

	if err != nil {
		return resp.StatusCode, fmt.Errorf(
			"failed to decode response body: %v",
			err,
		)
	}

	return resp.StatusCode, nil
}
