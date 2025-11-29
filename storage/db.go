package storage

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/twj0/subcheck/utils"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

/**
 * defaultPath 函数用于获取默认的数据文件路径
 * @return {string} 返回默认的数据库文件路径
 * @return {error} 返回可能出现的错误
 */
func defaultPath() (string, error) {
	// 拼接数据目录路径，使用可执行文件所在目录下的data文件夹
	dir := filepath.Join(utils.GetExecutablePath(), "data")
	// 尝试创建目录（包括所有必要的父目录），权限设置为0755
	if err := os.MkdirAll(dir, 0755); err != nil {
		// 如果创建目录失败，返回空字符串和错误信息
		return "", err
	}
	// 返回完整的数据库文件路径
	return filepath.Join(dir, "subcheck.db"), nil
}

// Init 函数用于初始化数据库连接
// 参数:
//   - path: 数据库文件的路径，如果为空则使用默认路径
// 返回值:
//   - error: 初始化过程中发生的错误，如果成功则为nil
func Init(path string) error {
	if path == "" {
		p, err := defaultPath()
		if err != nil {
			return err
		}
		path = p
	}
	dsn := "file:" + path + "?_pragma=busy_timeout=5000&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	DB = db
	return nil
}

// Migrate 函数用于执行数据库迁移，创建必要的表结构
func Migrate() error {
	// 定义需要执行的SQL语句数组，包含所有需要创建的表结构
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS subscriptions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(255) NOT NULL,
			url TEXT NOT NULL,
			enabled BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS speed_test_results (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			subscription_id INTEGER,
			node_name VARCHAR(255),
			delay INTEGER,
			download_speed REAL,
			upload_speed REAL,
			ip_address TEXT,
			proxy_json TEXT,
			test_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (subscription_id) REFERENCES subscriptions(id)
		);`,
		`CREATE TABLE IF NOT EXISTS ip_quality_results (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			subscription_id INTEGER,
			ip_address VARCHAR(45),
			fraud_score INTEGER,
			risk_level VARCHAR(50),
			is_proxy BOOLEAN,
			is_vpn BOOLEAN,
			is_tor BOOLEAN,
			country_code VARCHAR(10),
			test_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (subscription_id) REFERENCES subscriptions(id)
		);`,
		`CREATE TABLE IF NOT EXISTS system_config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			speed_test_interval INTEGER DEFAULT 86400,
			ip_quality_test_interval INTEGER DEFAULT 2592000,
			max_concurrent_tests INTEGER DEFAULT 5
		);`,
	}
	for _, s := range stmts {
		if _, err := DB.Exec(s); err != nil {
			return err
		}
	}
	_, _ = DB.Exec(`ALTER TABLE speed_test_results ADD COLUMN proxy_json TEXT`)
	return nil
}

// Close 函数用于关闭数据库连接
// 如果 DB 不为 nil，则调用 DB 的 Close 方法关闭数据库连接
// 如果 DB 为 nil，则直接返回 nil
func Close() error {
	if DB != nil {  // 检查 DB 是否已经初始化
		return DB.Close()  // 调用数据库的 Close 方法关闭连接
	}
	return nil  // 如果 DB 未初始化，返回 nil
}
