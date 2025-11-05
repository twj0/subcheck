package storage

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/twj0/subcheck/utils"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

func defaultPath() (string, error) {
	dir := filepath.Join(utils.GetExecutablePath(), "data")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "subcheck.db"), nil
}

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

func Migrate() error {
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

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
