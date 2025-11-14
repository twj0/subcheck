package method

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/twj0/subcheck/config"
)

var githubAPIURL = "https://api.github.com"

// GitHubFileResponse 表示 GitHub 文件 API 响应
type GitHubFileResponse struct {
	SHA string `json:"sha"`
}

// GitHubFilePayload 表示创建/更新文件的请求
type GitHubFilePayload struct {
	Message string `json:"message"`
	Content string `json:"content"`
	Branch  string `json:"branch,omitempty"`
	SHA     string `json:"sha,omitempty"`
}

// UploadToGitHubRaw 上传数据到 GitHub 仓库
func UploadToGitHubRaw(data []byte, filename string) error {
	if config.GlobalConfig.GithubAPIMirror != "" {
		githubAPIURL = config.GlobalConfig.GithubAPIMirror
	}

	client := &http.Client{Timeout: 30 * time.Second}

	// 获取文件当前 SHA（如果存在）
	sha, _ := getFileSHA(client, filename)

	// 准备请求
	content := base64.StdEncoding.EncodeToString(data)
	payload := GitHubFilePayload{
		Message: fmt.Sprintf("Update %s", filename),
		Content: content,
		Branch:  config.GlobalConfig.GithubRawBranch,
		SHA:     sha,
	}

	jsonData, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s%s",
		githubAPIURL,
		config.GlobalConfig.GithubRawOwner,
		config.GlobalConfig.GithubRawRepo,
		config.GlobalConfig.GithubRawPath,
		filename,
	)

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+config.GlobalConfig.GithubRawToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("上传到GitHub失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub返回错误(状态码: %d): %s", resp.StatusCode, string(body))
	}

	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s%s",
		config.GlobalConfig.GithubRawOwner,
		config.GlobalConfig.GithubRawRepo,
		config.GlobalConfig.GithubRawBranch,
		config.GlobalConfig.GithubRawPath,
		filename,
	)

	slog.Info("GitHub Raw上传成功", "filename", filename, "url", rawURL)
	return nil
}

// getFileSHA 获取文件当前的 SHA
func getFileSHA(client *http.Client, filename string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s%s?ref=%s",
		githubAPIURL,
		config.GlobalConfig.GithubRawOwner,
		config.GlobalConfig.GithubRawRepo,
		config.GlobalConfig.GithubRawPath,
		filename,
		config.GlobalConfig.GithubRawBranch,
	)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+config.GlobalConfig.GithubRawToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", nil // 文件不存在
	}

	body, _ := io.ReadAll(resp.Body)
	var result GitHubFileResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.SHA, nil
}

// ValiGitHubRawConfig 验证GitHub Raw配置
func ValiGitHubRawConfig() error {
	if config.GlobalConfig.GithubRawToken == "" {
		return fmt.Errorf("github-raw-token未配置")
	}
	if config.GlobalConfig.GithubRawOwner == "" {
		return fmt.Errorf("github-raw-owner未配置")
	}
	if config.GlobalConfig.GithubRawRepo == "" {
		return fmt.Errorf("github-raw-repo未配置")
	}
	if config.GlobalConfig.GithubRawBranch == "" {
		config.GlobalConfig.GithubRawBranch = "main"
	}
	if config.GlobalConfig.GithubRawPath != "" && config.GlobalConfig.GithubRawPath[len(config.GlobalConfig.GithubRawPath)-1] != '/' {
		config.GlobalConfig.GithubRawPath += "/"
	}
	return nil
}
