package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Subscription struct {
	ID        int64
	Name      string
	URL       string
	Enabled   bool
	CreatedAt time.Time
}

func QueryTopNProxyJSONs(ctx context.Context, selectBy string, n int, windowHours int) ([]string, error) {
	if n <= 0 {
		n = 10
	}
	if windowHours <= 0 {
		windowHours = 24
	}
	field := "download_speed"
	switch strings.ToLower(selectBy) {
	case "delay":
		field = "delay"
	case "download_speed":
		field = "download_speed"
	default:
		// fallback
		field = "download_speed"
	}
	order := "DESC"
	if field == "delay" {
		order = "ASC"
	}
	q := `SELECT proxy_json FROM speed_test_results WHERE proxy_json IS NOT NULL AND proxy_json != '' AND test_time >= datetime('now', ? ) ORDER BY ` + field + ` ` + order + ` LIMIT ?`
	argWindow := fmt.Sprintf("-%d hour", windowHours)
	rows, err := DB.QueryContext(ctx, q, argWindow, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s sql.NullString
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		if s.Valid && s.String != "" {
			out = append(out, s.String)
		}
	}
	return out, nil
}

type SpeedResult struct {
	ID             int64
	SubscriptionID sql.NullInt64
	NodeName       string
	Delay          sql.NullInt64
	DownloadSpeed  sql.NullFloat64
	UploadSpeed    sql.NullFloat64
	IPAddress      sql.NullString
	ProxyJSON      sql.NullString
	TestTime       time.Time
}

type IPQualityResult struct {
	ID             int64
	SubscriptionID sql.NullInt64
	IPAddress      string
	FraudScore     sql.NullInt64
	RiskLevel      sql.NullString
	IsProxy        sql.NullBool
	IsVPN          sql.NullBool
	IsTor          sql.NullBool
	CountryCode    sql.NullString
	TestTime       time.Time
}

func CreateSubscription(ctx context.Context, name, url string, enabled bool) (int64, error) {
	res, err := DB.ExecContext(ctx, `INSERT INTO subscriptions (name, url, enabled) VALUES (?, ?, ?)`, name, url, enabled)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdateSubscription(ctx context.Context, id int64, name, url string, enabled bool) error {
	_, err := DB.ExecContext(ctx, `UPDATE subscriptions SET name=?, url=?, enabled=? WHERE id=?`, name, url, enabled, id)
	return err
}

func DeleteSubscription(ctx context.Context, id int64) error {
	_, err := DB.ExecContext(ctx, `DELETE FROM subscriptions WHERE id=?`, id)
	return err
}

func ListSubscriptions(ctx context.Context, page, pageSize int) ([]Subscription, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	var total int64
	if err := DB.QueryRowContext(ctx, `SELECT COUNT(1) FROM subscriptions`).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := DB.QueryContext(ctx, `SELECT id,name,url,enabled,created_at FROM subscriptions ORDER BY id DESC LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []Subscription
	for rows.Next() {
		var s Subscription
		if err := rows.Scan(&s.ID, &s.Name, &s.URL, &s.Enabled, &s.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, s)
	}
	return list, total, nil
}

func SaveSpeedResult(ctx context.Context, subscriptionID sql.NullInt64, nodeName string, delay sql.NullInt64, download float64, upload sql.NullFloat64, ipAddr sql.NullString, proxyJSON sql.NullString) error {
	_, err := DB.ExecContext(ctx, `INSERT INTO speed_test_results (subscription_id, node_name, delay, download_speed, upload_speed, ip_address, proxy_json) VALUES (?,?,?,?,?,?,?)`, subscriptionID, nodeName, delay, download, upload, ipAddr, proxyJSON)
	return err
}

func QuerySpeedResults(ctx context.Context, page, pageSize int, nodeLike string, minSpeed, maxSpeed *float64, sortBy, sortDir string) ([]SpeedResult, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	validSort := map[string]bool{"test_time": true, "download_speed": true, "node_name": true}
	if !validSort[sortBy] {
		sortBy = "test_time"
	}
	sortDir = strings.ToUpper(sortDir)
	if sortDir != "ASC" && sortDir != "DESC" {
		sortDir = "DESC"
	}
	var where []string
	var args []any
	if nodeLike != "" {
		where = append(where, "node_name LIKE ?")
		args = append(args, "%"+nodeLike+"%")
	}
	if minSpeed != nil {
		where = append(where, "download_speed >= ?")
		args = append(args, *minSpeed)
	}
	if maxSpeed != nil {
		where = append(where, "download_speed <= ?")
		args = append(args, *maxSpeed)
	}
	queryWhere := ""
	if len(where) > 0 {
		queryWhere = " WHERE " + strings.Join(where, " AND ")
	}
	var total int64
	if err := DB.QueryRowContext(ctx, `SELECT COUNT(1) FROM speed_test_results`+queryWhere, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := DB.QueryContext(ctx, `SELECT id,subscription_id,node_name,delay,download_speed,upload_speed,ip_address,proxy_json,test_time FROM speed_test_results`+queryWhere+` ORDER BY `+sortBy+` `+sortDir+` LIMIT ? OFFSET ?`, append(args, pageSize, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []SpeedResult
	for rows.Next() {
		var r SpeedResult
		if err := rows.Scan(&r.ID, &r.SubscriptionID, &r.NodeName, &r.Delay, &r.DownloadSpeed, &r.UploadSpeed, &r.IPAddress, &r.ProxyJSON, &r.TestTime); err != nil {
			return nil, 0, err
		}
		list = append(list, r)
	}
	return list, total, nil
}

func QueryIPQualityResults(ctx context.Context, page, pageSize int, ip, country, risk, sortBy, sortDir string) ([]IPQualityResult, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	validSort := map[string]bool{"test_time": true, "fraud_score": true, "ip_address": true}
	if !validSort[sortBy] {
		sortBy = "test_time"
	}
	sortDir = strings.ToUpper(sortDir)
	if sortDir != "ASC" && sortDir != "DESC" {
		sortDir = "DESC"
	}
	var where []string
	var args []any
	if ip != "" {
		where = append(where, "ip_address LIKE ?")
		args = append(args, "%"+ip+"%")
	}
	if country != "" {
		where = append(where, "country_code = ?")
		args = append(args, country)
	}
	if risk != "" {
		where = append(where, "risk_level = ?")
		args = append(args, risk)
	}
	queryWhere := ""
	if len(where) > 0 {
		queryWhere = " WHERE " + strings.Join(where, " AND ")
	}
	var total int64
	if err := DB.QueryRowContext(ctx, `SELECT COUNT(1) FROM ip_quality_results`+queryWhere, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := DB.QueryContext(ctx, `SELECT id,subscription_id,ip_address,fraud_score,risk_level,is_proxy,is_vpn,is_tor,country_code,test_time FROM ip_quality_results`+queryWhere+` ORDER BY `+sortBy+` `+sortDir+` LIMIT ? OFFSET ?`, append(args, pageSize, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []IPQualityResult
	for rows.Next() {
		var r IPQualityResult
		if err := rows.Scan(&r.ID, &r.SubscriptionID, &r.IPAddress, &r.FraudScore, &r.RiskLevel, &r.IsProxy, &r.IsVPN, &r.IsTor, &r.CountryCode, &r.TestTime); err != nil {
			return nil, 0, err
		}
		list = append(list, r)
	}
	return list, total, nil
}

type Dashboard struct {
	Subscriptions int64
	SpeedTests    int64
	IPChecks      int64
	AvgSpeed7d    float64
	RiskCounts    map[string]int64
}

func GetDashboard(ctx context.Context) (Dashboard, error) {
	var d Dashboard
	if err := DB.QueryRowContext(ctx, `SELECT COUNT(1) FROM subscriptions`).Scan(&d.Subscriptions); err != nil {
		return d, err
	}
	if err := DB.QueryRowContext(ctx, `SELECT COUNT(1) FROM speed_test_results WHERE test_time >= datetime('now','-7 day')`).Scan(&d.SpeedTests); err != nil {
		return d, err
	}
	if err := DB.QueryRowContext(ctx, `SELECT COUNT(1) FROM ip_quality_results WHERE test_time >= datetime('now','-30 day')`).Scan(&d.IPChecks); err != nil {
		return d, err
	}
	if err := DB.QueryRowContext(ctx, `SELECT COALESCE(AVG(download_speed),0) FROM speed_test_results WHERE test_time >= datetime('now','-7 day')`).Scan(&d.AvgSpeed7d); err != nil {
		return d, err
	}
	rows, err := DB.QueryContext(ctx, `SELECT risk_level, COUNT(1) FROM ip_quality_results WHERE test_time >= datetime('now','-30 day') GROUP BY risk_level`)
	if err != nil {
		return d, err
	}
	defer rows.Close()
	d.RiskCounts = map[string]int64{}
	for rows.Next() {
		var k sql.NullString
		var c int64
		if err := rows.Scan(&k, &c); err != nil {
			return d, err
		}
		key := k.String
		if key == "" {
			key = "Unknown"
		}
		d.RiskCounts[key] = c
	}
	return d, nil
}

func QueryTopNSpeedIPs(ctx context.Context, n int) ([]string, error) {
	rows, err := DB.QueryContext(ctx, `SELECT ip_address FROM speed_test_results WHERE ip_address IS NOT NULL AND ip_address != '' ORDER BY download_speed DESC LIMIT ?`, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var ip sql.NullString
		if err := rows.Scan(&ip); err != nil {
			return nil, err
		}
		if ip.Valid && ip.String != "" {
			out = append(out, ip.String)
		}
	}
	return out, nil
}
