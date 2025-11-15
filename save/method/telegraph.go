package method

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/twj0/subcheck/config"
)

const (
	telegraphAPIURL = "https://api.telegra.ph"
)

// TelegraphPage 表示 Telegraph 页面响应
type TelegraphPage struct {
	Ok     bool `json:"ok"`
	Result struct {
		Path string `json:"path"`
		URL  string `json:"url"`
	} `json:"result"`
}

// TelegraphContent 表示 Telegraph 内容节点
type TelegraphContent struct {
	Tag      string `json:"tag"`
	Children []any  `json:"children,omitempty"`
}

// TelegraphPayload 表示创建/编辑页面的请求
type TelegraphPayload struct {
	AccessToken string             `json:"access_token,omitempty"`
	Path        string             `json:"path,omitempty"`
	Title       string             `json:"title"`
	Content     []TelegraphContent `json:"content"`
	AuthorName  string             `json:"author_name,omitempty"`
}

// UploadToTelegraph 上传数据到 Telegraph
func UploadToTelegraph(data []byte, filename string) error {
	// 只上传 mihomo.yaml 到远程存储
	if filename != "mihomo.yaml" {
		return nil
	}
	
	client := &http.Client{Timeout: 60 * time.Second}

	content := []TelegraphContent{
		{Tag: "pre", Children: []any{string(data)}},
	}

	payload := TelegraphPayload{
		Title:   filename,
		Content: content,
	}

	// 如果配置了 access_token 和 path，则编辑已有页面
	if config.GlobalConfig.TelegraphToken != "" && config.GlobalConfig.TelegraphPath != "" {
		payload.AccessToken = config.GlobalConfig.TelegraphToken
		payload.Path = config.GlobalConfig.TelegraphPath
		return editPage(client, payload, filename)
	}

	// 否则创建新页面
	return createPage(client, payload, filename)
}

// createPage 创建新页面
func createPage(client *http.Client, payload TelegraphPayload, filename string) error {
	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", telegraphAPIURL+"/createPage", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("创建Telegraph页面失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result TelegraphPage
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if !result.Ok {
		return fmt.Errorf("Telegraph返回错误: %s", string(body))
	}

	slog.Info("Telegraph上传成功", "filename", filename, "url", result.Result.URL)
	return nil
}

// editPage 编辑已有页面
func editPage(client *http.Client, payload TelegraphPayload, filename string) error {
	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", telegraphAPIURL+"/editPage/"+payload.Path, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("编辑Telegraph页面失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result TelegraphPage
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if !result.Ok {
		return fmt.Errorf("Telegraph返回错误: %s", string(body))
	}

	slog.Info("Telegraph更新成功", "filename", filename, "url", result.Result.URL)
	return nil
}

// ValiTelegraphConfig 验证Telegraph配置
func ValiTelegraphConfig() error {
	// Telegraph 可以匿名创建，无需验证
	return nil
}
