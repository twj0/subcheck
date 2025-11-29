package proxies

import (
	"fmt"
)

// DeduplicateProxies 是一个去重函数，用于根据代理服务器的特定属性去重
// 参数 proxies 是一个包含代理服务器信息的map切片，每个map代表一个代理
// 返回值是去重后的代理服务器map切片
func DeduplicateProxies(proxies []map[string]any) []map[string]any {
	// seenKeys 用于记录已经出现过的代理组合，key是代理的组合字符串
	seenKeys := make(map[string]bool)
	// result 用于存储去重后的代理列表，预分配足够空间以提高性能
	result := make([]map[string]any, 0, len(proxies))

	// 遍历所有代理
	for _, proxy := range proxies {
		// 获取服务器地址，如果为空则跳过该代理
		server, _ := proxy["server"].(string)
		if server == "" {
			continue
		}
		// 获取服务器名称
		servername, _ := proxy["servername"].(string)

		// 获取密码，如果为空则尝试使用uuid作为密码
		password, _ := proxy["password"].(string)
		if password == "" {
			password, _ = proxy["uuid"].(string)
		}

		// 创建组合键，包含服务器地址、端口、服务器名称和密码
		key := fmt.Sprintf("%s:%v:%s:%s", server, proxy["port"], servername, password)
		// 如果该组合键未出现过，则将其加入结果集
		if !seenKeys[key] {
			seenKeys[key] = true
			result = append(result, proxy)
		}
	}

	// 返回去重后的代理列表
	return result
}
