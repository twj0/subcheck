package save

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/twj0/subcheck/check"
	"github.com/twj0/subcheck/config"
	"github.com/twj0/subcheck/save/method"
	"github.com/twj0/subcheck/utils"
	"gopkg.in/yaml.v3"
)

// ProxyCategory 定义代理分类
type ProxyCategory struct {
	Name    string                     // 分类名称
	Proxies []map[string]any           // 该分类下的代理列表
	Filter  func(result check.Result) bool // 用于过滤代理的函数
}

// ConfigSaver 处理配置保存的结构体
type ConfigSaver struct {
	results     []check.Result         // 检查结果列表
	categories  []ProxyCategory        // 代理分类列表
	saveMethods []func([]byte, string) error // 保存方法列表
}

// NewConfigSaver 创建新的配置保存器
//
// 参数:
//   - results: 检查结果列表
//
// 返回值:
//   - *ConfigSaver: 配置保存器实例
// NewConfigSaver 是一个构造函数，用于创建并初始化一个新的 ConfigSaver 实例
// 参数:
//   - results: check.Result 类型的切片，包含需要保存的检查结果
// 返回值:
//   - *ConfigSaver: 指向新创建的 ConfigSaver 实例的指针
func NewConfigSaver(results []check.Result) *ConfigSaver {
	return &ConfigSaver{
		results:     results, // 将传入的检查结果赋值给 ConfigSaver 的 results 字段
		saveMethods: chooseSaveMethods(), // 调用 chooseSaveMethods 函数选择保存方法
		categories: []ProxyCategory{ // 初始化代理分类切片，包含三个默认分类
			{
				Name:    "all.yaml", // 分类名称，表示所有代理的配置文件
				Proxies: make([]map[string]any, 0), // 初始化空代理列表
				Filter:  func(result check.Result) bool { return true }, // 过滤函数，接受所有代理
			},
			{
				Name:    "mihomo.yaml", // 分类名称，表示 mihomo 代理的配置文件
				Proxies: make([]map[string]any, 0), // 初始化空代理列表
				Filter:  func(result check.Result) bool { return true }, // 过滤函数，接受所有代理
			},
			{
				Name:    "base64.txt", // 分类名称，表示 base64 编码的代理列表
				Proxies: make([]map[string]any, 0), // 初始化空代理列表
				Filter:  func(result check.Result) bool { return true }, // 过滤函数，接受所有代理
			},
		},
	}
}

// SaveConfig 保存配置的入口函数
//
// 参数:
//   - results: 检查结果列表
func SaveConfig(results []check.Result) {
	saver := NewConfigSaver(results)
	if err := saver.Save(); err != nil {
		slog.Error(fmt.Sprintf("保存配置失败: %v", err))
	}
}

// Save 执行保存操作
//
// 参数:
//   - cs: 配置保存器实例
//
// 返回值:
//   - error: 保存过程中可能发生的错误
func (cs *ConfigSaver) Save() error {
	// 分类处理代理
	cs.categorizeProxies()

	// 保存各个类别的代理
	for _, category := range cs.categories {
		if err := cs.saveCategory(category); err != nil {
			slog.Error(fmt.Sprintf("保存失败: %v", err))
			continue
		}
	}

	return nil
}

// injectIPQualityToMihomo 在 mihomo.yaml 中为每个节点注入 IP 纯净度相关信息
//
// 参数:
//   - cs: 配置保存器实例
//   - data: 原始数据
//
// 返回值:
//   - []byte: 注入IP质量信息后的数据
func (cs *ConfigSaver) injectIPQualityToMihomo(data []byte) []byte {
	// 没有检测结果则直接返回原始内容
	if len(cs.results) == 0 {
		return data
	}

	// 根据节点名称构建 IP 风控索引
	type ipQuality struct {
		Risk      string // IP风险等级
		Country   string // 国家代码
		IP        string // IP地址
		SpeedKBps int    // 速度(KB/s)
	}

	index := make(map[string]ipQuality)
	for _, r := range cs.results {
		name, _ := r.Proxy["name"].(string)
		if name == "" {
			continue
		}
		if r.IPRisk == "" && r.Country == "" && r.IP == "" && r.SpeedKBps == 0 {
			continue
		}
		index[name] = ipQuality{
			Risk:      r.IPRisk,
			Country:   r.Country,
			IP:        r.IP,
			SpeedKBps: r.SpeedKBps,
		}
	}

	if len(index) == 0 {
		return data
	}

	// 解析 mihomo.yaml
	var cfg map[string]any
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		slog.Debug("解析 mihomo.yaml 失败，跳过 IP 纯净度注入", "error", err)
		return data
	}

	proxiesAny, ok := cfg["proxies"]
	if !ok {
		return data
	}

	proxies, ok := proxiesAny.([]any)
	if !ok {
		return data
	}

	changed := false
	for i, p := range proxies {
		mp, ok := p.(map[string]any)
		if !ok {
			continue
		}
		name, _ := mp["name"].(string)
		if name == "" {
			continue
		}
		q, ok := index[name]
		if !ok || q.Risk == "" {
			continue
		}
		// 将 IP 风控信息附加到节点上，供 mihomo/clash.meta 客户端或外部工具使用
		mp["ip_risk"] = q.Risk
		if q.Country != "" {
			mp["ip_country"] = q.Country
		}
		if q.IP != "" {
			mp["ip_address"] = q.IP
		}
		if q.SpeedKBps > 0 {
			mp["speed_kbps"] = q.SpeedKBps
		}
		proxies[i] = mp
		changed = true
	}

	if !changed {
		return data
	}

	cfg["proxies"] = proxies
	out, err := yaml.Marshal(cfg)
	if err != nil {
		slog.Debug("重新序列化 mihomo.yaml 失败，跳过 IP 纯净度注入", "error", err)
		return data
	}

	return out
}

// categorizeProxies 将代理按类别分类
//
// 参数:
//   - cs: 配置保存器实例
// categorizeProxies 方法用于将代理分类到不同的类别中
// 它遍历所有结果，并根据每个类别的过滤器将代理添加到相应的类别中
func (cs *ConfigSaver) categorizeProxies() {
    // 遍历所有结果
	for _, result := range cs.results {
        // 遍历所有类别
		for i := range cs.categories {
            // 使用类别的过滤器检查结果是否符合该类别
			if cs.categories[i].Filter(result) {
                // 如果符合，则将代理添加到该类别的代理列表中
				cs.categories[i].Proxies = append(cs.categories[i].Proxies, result.Proxy)
			}
		}
	}
}

// saveCategory 保存单个类别的代理
//
// 参数:
//   - cs: 配置保存器实例
//   - category: 要保存的代理分类
//
// 返回值:
//   - error: 保存过程中可能发生的错误
/*
 * saveCategory 方法用于保存代理配置
 * @param category: ProxyCategory 结构体，包含代理名称和代理列表
 * @return error: 错误信息，保存成功返回 nil
 */
func (cs *ConfigSaver) saveCategory(category ProxyCategory) error {
	// 检查代理列表是否为空，为空则跳过保存并记录警告日志
	if len(category.Proxies) == 0 {
		slog.Warn(fmt.Sprintf("yaml节点为空，跳过保存: %s", category.Name))
		return nil
	}

	var data []byte  // 用于存储要保存的数据
	var err error    // 用于存储过程中可能出现的错误

	// 根据不同类别名称执行不同的保存逻辑
	if category.Name == "all.yaml" {
		// 将代理列表序列化为 YAML 格式
		data, err = yaml.Marshal(map[string]any{
			"proxies": category.Proxies,
		})
		if err != nil {
			return fmt.Errorf("序列化yaml %s 失败: %w", category.Name, err)
		}
		// 更新substore（如果启用）
		if config.GlobalConfig.SubStorePort != "" {
			utils.UpdateSubStore(data)
		}
	} else if category.Name == "mihomo.yaml" && config.GlobalConfig.SubStorePort != "" {
		resp, err := http.Get(fmt.Sprintf("%s/api/file/%s", utils.BaseURL, utils.MihomoName))
		if err != nil {
			return fmt.Errorf("获取mihomo file请求失败: %w", err)
		}
		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("读取mihomo file失败: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("获取mihomo file失败, 状态码: %d, 错误信息: %s", resp.StatusCode, data)
		}
		// 基于最新检测结果，为 mihomo.yaml 中的节点注入 IP 纯净度相关信息
		data = cs.injectIPQualityToMihomo(data)
	} else if category.Name == "base64.txt" && config.GlobalConfig.SubStorePort != "" {
		resp, err := http.Get(fmt.Sprintf("%s/download/%s?target=V2Ray", utils.BaseURL, utils.SubName))
		if err != nil {
			return fmt.Errorf("获取base64.txt请求失败: %w", err)
		}
		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("读取base64.txt失败: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("获取base64.txt失败，状态码: %d, 错误信息: %s", resp.StatusCode, data)
		}
	} else {
		return nil
	}

	// 远程存储只保存 mihomo.yaml，本地保存所有格式
	for _, saveMethod := range cs.saveMethods {
		if err := saveMethod(data, category.Name); err != nil {
			slog.Error(fmt.Sprintf("保存 %s 失败: %v", category.Name, err))
		}
	}

	return nil
}

// chooseSaveMethods 根据配置选择保存方法（支持多个）
//
// 返回值:
//   - []func([]byte, string) error: 保存方法函数列表
// chooseSaveMethods 选择并返回可用的保存方法函数列表
// 返回值: 包含各种保存方法的函数切片，每个函数接受字节数组和文件路径作为参数并返回错误
func chooseSaveMethods() []func([]byte, string) error {
	var methods []func([]byte, string) error

	// 解析 save-method 配置（支持字符串或数组）
	// 从全局配置中获取保存方法配置，并处理多种可能的配置格式
	var methodNames []string
	switch v := config.GlobalConfig.SaveMethod.(type) {
	case string:
		// 如果配置是单个字符串，则将其转换为单元素切片
		methodNames = []string{v}
	case []any:
		// 如果配置是任意类型数组，则尝试将每个元素转换为字符串
		for _, item := range v {
			if s, ok := item.(string); ok {
				methodNames = append(methodNames, s)
			}
		}
	case []string:
		// 如果配置已经是字符串数组，则直接使用
		methodNames = v
	default:
		// 如果配置格式不支持，则输出警告并使用默认的 local 方法
		slog.Warn("save-method 配置格式错误，使用默认 local")
		methodNames = []string{"local"}
	}

	// 始终包含 local 方法（确保本地有备份）
	// 检查方法列表中是否已包含 local 方法
	hasLocal := false
	for _, name := range methodNames {
		if name == "local" {
			hasLocal = true
			break
		}
	}
	// 如果没有 local 方法，则将其添加到列表开头
	if !hasLocal {
		methodNames = append([]string{"local"}, methodNames...)
	}

	// 为每个方法名创建对应的保存函数
	// 根据方法名获取对应的保存函数，并添加到方法列表中
	for _, name := range methodNames {
		method := getSaveMethod(name)
		if method != nil {
			methods = append(methods, method)
		}
	}

	return methods
}

// getSaveMethod 根据方法名获取对应的保存函数
//
// 参数:
//   - name: 保存方法名称
//
// 返回值:
//   - func([]byte, string) error: 对应的保存函数，如果未找到则返回nil
// getSaveMethod 根据传入的名称返回对应的保存方法函数
// 参数:
//   name: 保存方法的名称，如 "r2", "gist", "webdav" 等
// 返回值:
//   func([]byte, string) error: 对应的保存方法函数，如果配置不完整或方法未知则返回 nil
func getSaveMethod(name string) func([]byte, string) error {
	switch name {
	case "r2":
		// 检查 R2 配置是否完整
		if err := method.ValiR2Config(); err != nil {
			// 如果配置不完整，记录错误日志并返回 nil
			slog.Error(fmt.Sprintf("R2配置不完整，跳过: %v", err))
			return nil
		}
		// 返回 R2 存储上传方法
		return method.UploadToR2Storage
	case "gist":
		// 检查 Gist 配置是否完整
		if err := method.ValiGistConfig(); err != nil {
			// 如果配置不完整，记录错误日志并返回 nil
			slog.Error(fmt.Sprintf("Gist配置不完整，跳过: %v", err))
			return nil
		}
		// 返回 Gist 上传方法
		return method.UploadToGist
	case "webdav":
		// 检查 WebDAV 配置是否完整
		if err := method.ValiWebDAVConfig(); err != nil {
			// 如果配置不完整，记录错误日志并返回 nil
			slog.Error(fmt.Sprintf("WebDAV配置不完整，跳过: %v", err))
			return nil
		}
		// 返回 WebDAV 上传方法
		return method.UploadToWebDAV
	case "local":
		// 本地保存方法不需要配置检查，直接返回
		return method.SaveToLocal
	case "s3":
		// 检查 S3 配置是否完整
		if err := method.ValiS3Config(); err != nil {
			// 如果配置不完整，记录错误日志并返回 nil
			slog.Error(fmt.Sprintf("S3配置不完整，跳过: %v", err))
			return nil
		}
		// 返回 S3 上传方法
		return method.UploadToS3
	case "telegraph":
		// 检查 Telegraph 配置是否完整
		if err := method.ValiTelegraphConfig(); err != nil {
			// 如果配置不完整，记录错误日志并返回 nil
			slog.Error(fmt.Sprintf("Telegraph配置不完整，跳过: %v", err))
			return nil
		}
		// 返回 Telegraph 上传方法
		return method.UploadToTelegraph
	case "github-raw":
		// 检查 GitHub Raw 配置是否完整
		if err := method.ValiGitHubRawConfig(); err != nil {
			// 如果配置不完整，记录错误日志并返回 nil
			slog.Error(fmt.Sprintf("GitHub Raw配置不完整，跳过: %v", err))
			return nil
		}
		// 返回 GitHub Raw 上传方法
		return method.UploadToGitHubRaw
	default:
		// 如果传入的保存方法名称未知，记录警告日志并返回 nil
		slog.Warn(fmt.Sprintf("未知的保存方法: %s", name))
		return nil
	}
}