package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/twj0/subcheck/app/monitor"
	"github.com/twj0/subcheck/assets"
	"github.com/twj0/subcheck/check"
	"github.com/twj0/subcheck/config"
	"github.com/twj0/subcheck/ipcheck"
	"github.com/twj0/subcheck/save"
	plat "github.com/twj0/subcheck/check/platform"
	"github.com/twj0/subcheck/storage"
	proxyutils "github.com/twj0/subcheck/proxy"
	"github.com/twj0/subcheck/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron/v3"
)

// App 结构体用于管理应用程序状态
type App struct {
	configPath string
	interval   int
	watcher    *fsnotify.Watcher
	checkChan  chan struct{} // 触发检测的通道
	checking   atomic.Bool   // 检测状态标志
	ipChecking atomic.Bool   // IP质量检测状态标志
	ticker     *time.Ticker
	done       chan struct{} // 用于结束ticker goroutine的信号
	cron       *cron.Cron    // crontab调度器
	ipCron     *cron.Cron    // IP质量检测调度器（每月）
	version    string
}

// initIPCron 初始化每月IP质量检测任务
func (app *App) initIPCron() error {
	if app.ipCron != nil {
		app.ipCron.Stop()
		app.ipCron = nil
	}
	if !config.GlobalConfig.IpCheck.MonthlyRun {
		slog.Warn("Monthly IP quality check task is not enabled")
		return nil
	}
	app.ipCron = cron.New()
	spec := config.GlobalConfig.IpCheck.Cron
	if spec == "" {
		spec = "0 0 1 * *"
	}
	_, err := app.ipCron.AddFunc(spec, func() {
		go app.triggerIPCheck()
	})
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to parse IP quality check cron: %v", err))
		return err
	}
	app.ipCron.Start()
	slog.Info("IP quality check task started", "cron", spec)
	return nil
}

// triggerIPCheck 触发一次IP质量检测
func (app *App) triggerIPCheck() {
	if !config.GlobalConfig.IpCheck.Enabled {
		return
	}
	if !app.ipChecking.CompareAndSwap(false, true) {
		slog.Warn("IP quality check is already in progress, skipping this check")
		return
	}
	defer app.ipChecking.Store(false)

	slog.Info("Starting IP quality check")
	timeout := time.Duration(config.GlobalConfig.IpCheck.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 300 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if config.GlobalConfig.IpCheck.UseTopN {
		n := config.GlobalConfig.IpCheck.TopN
		if n <= 0 {
			n = 10
		}
		window := config.GlobalConfig.IpCheck.WindowHours
		items, err := storage.QueryTopNProxyJSONs(context.Background(), config.GlobalConfig.IpCheck.SelectBy, n, window)
		if err != nil || len(items) == 0 {
			slog.Error(fmt.Sprintf("获取TopN代理失败: %v", err))
			return
		}
		rate := config.GlobalConfig.IpCheck.RatePerMin
		interval := time.Duration(0)
		if rate > 0 {
			interval = time.Minute / time.Duration(rate)
		}
		conc := config.GlobalConfig.IpCheck.Concurrent
		if conc <= 0 {
			conc = 3
		}
		sem := make(chan struct{}, conc)
		ticker := (&time.Ticker{})
		if interval > 0 {
			ticker = time.NewTicker(interval)
			defer ticker.Stop()
		}
		var wg sync.WaitGroup
		for idx, js := range items {
			if interval > 0 {
				if idx > 0 {
					<-ticker.C
				}
			}
			var m map[string]any
			if err := json.Unmarshal([]byte(js), &m); err != nil {
				continue
			}
			sem <- struct{}{}
			wg.Add(1)
			go func(mp map[string]any) {
				defer func(){ <-sem; wg.Done() }()
				pc := check.CreateClient(mp)
				if pc == nil { return }
				country, ip := proxyutils.GetProxyCountry(pc.Client)
				if ip == "" { return }
				risk, err := plat.CheckIPRisk(pc.Client, ip)
				if err != nil { return }
				var score int
				if len(risk) > 0 && risk[len(risk)-1] == '%' {
					fmt.Sscanf(risk, "%d%%", &score)
				}
				level := func(s int) string {
					switch {
					case s <= 10:
						return "VeryLow"
					case s <= 25:
						return "Low"
					case s <= 50:
						return "Medium"
					case s <= 75:
						return "High"
					default:
						if s > 0 { return "VeryHigh" }
						return "Unknown"
					}
				}(score)
				if storage.DB != nil {
					_, _ = storage.DB.Exec(`
						INSERT INTO ip_quality_results (
							subscription_id, ip_address, fraud_score, risk_level, is_proxy, is_vpn, is_tor, country_code
						) VALUES (NULL, ?, ?, ?, NULL, NULL, NULL, ?)
					`, ip, score, level, country)
				}
			}(m)
		}
		wg.Wait()
		slog.Info("Per-node IP质量检测完成", "count", len(items))
		return
	}

	res, err := ipcheck.Run(ctx, "")
	if err != nil {
		slog.Error(fmt.Sprintf("IP quality check failed: %v", err))
		return
	}
	if res == nil {
		slog.Warn("IP quality check returned empty result")
		return
	}

	core := ipcheck.ExtractCore(res)
	if storage.DB != nil {
		_, err := storage.DB.Exec(`
			INSERT INTO ip_quality_results (
				subscription_id, ip_address, fraud_score, risk_level, is_proxy, is_vpn, is_tor, country_code
			) VALUES (NULL, ?, ?, ?, ?, ?, ?, ?)
		`, core.IP, core.FraudScore, core.RiskLevel, core.IsProxy, core.IsVPN, core.IsTor, core.CountryCode)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to save IP quality check result: %v", err))
		} else {
			slog.Info("IP quality check completed and saved")
		}
	} else {
		slog.Warn("Database not initialized, result not saved")
	}
}

// New 创建新的应用实例
func New(version string) *App {
	configPath := flag.String("f", "", "配置文件路径")
	flag.Parse()

	return &App{
		configPath: *configPath,
		checkChan:  make(chan struct{}),
		done:       make(chan struct{}),
		version:    version,
	}
}

// Initialize 初始化应用程序
func (app *App) Initialize() error {
	// 初始化配置文件路径
	if err := app.initConfigPath(); err != nil {
		return fmt.Errorf("Failed to initialize config file path: %w", err)
	}

	// 加载配置文件
	if err := app.loadConfig(); err != nil {
		return fmt.Errorf("Failed to load config file: %w", err)
	}

	// 初始化配置文件监听
	if err := app.initConfigWatcher(); err != nil {
		return fmt.Errorf("Failed to initialize config file watcher: %w", err)
	}

	// 从配置文件中读取代理，设置代理
	if config.GlobalConfig.Proxy != "" {
		os.Setenv("HTTP_PROXY", config.GlobalConfig.Proxy)
		os.Setenv("HTTPS_PROXY", config.GlobalConfig.Proxy)
	}

	app.interval = func() int {
		if config.GlobalConfig.CheckInterval <= 0 {
			return 1
		}
		return config.GlobalConfig.CheckInterval
	}()

	if config.GlobalConfig.ListenPort != "" {
		if err := app.initHttpServer(); err != nil {
			return fmt.Errorf("Failed to initialize HTTP server: %w", err)
		}
	}

	if config.GlobalConfig.SubStorePort != "" {
		go assets.RunSubStoreService()
		// 求等吗得，日志会按预期顺序输出
		time.Sleep(500 * time.Millisecond)
	}

	// 启动内存监控
	monitor.StartMemoryMonitor()

	// 初始化数据库
	if err := storage.Init(""); err != nil {
		return fmt.Errorf("Failed to initialize database: %w", err)
	}
	if err := storage.Migrate(); err != nil {
		return fmt.Errorf("Database migration failed: %w", err)
	}

	// 初始化IP质量检测cron（每月执行一次）
	if config.GlobalConfig.IpCheck.Enabled {
		if err := app.initIPCron(); err != nil {
			return fmt.Errorf("Failed to initialize IP quality check task: %w", err)
		}
	}

	// 设置信号处理器
	utils.SetupSignalHandler(&check.ForceClose)
	return nil
}

// Run 运行应用程序主循环
func (app *App) Run() {
	defer func() {
		app.watcher.Close()
		if app.ticker != nil {
			app.ticker.Stop()
		}
		if app.cron != nil {
			app.cron.Stop()
		}
		if app.ipCron != nil {
			app.ipCron.Stop()
		}
		_ = storage.Close()
	}()

	// 设置初始定时器模式
	app.setTimer()

	// 仅在cron表达式为空时，首次启动立即执行检测
	if config.GlobalConfig.CronExpression != "" {
		slog.Warn("Using cron expression, skipping initial check")
	} else {
		app.triggerCheck()
	}

	// 在主循环中处理手动触发
	for range app.checkChan {
		go app.triggerCheck()
	}
}

// setTimer 根据配置设置定时器
func (app *App) setTimer() {
	// 停止现有定时器
	if app.ticker != nil {
		// 应该先发送停止信号，防止被=nil后panic
		close(app.done)                // 发送停止信号
		app.done = make(chan struct{}) // 创建新通道
		app.ticker.Stop()
		app.ticker = nil
	}

	// 停止现有cron
	if app.cron != nil {
		app.cron.Stop()
		app.cron = nil
	}

	// 检查是否设置了cron表达式
	if config.GlobalConfig.CronExpression != "" {
		slog.Info(fmt.Sprintf("Using cron expression: %s", config.GlobalConfig.CronExpression))
		app.cron = cron.New()
		_, err := app.cron.AddFunc(config.GlobalConfig.CronExpression, func() {
			app.triggerCheck()
		})
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to parse cron expression '%s': %v, falling back to interval timer",
				config.GlobalConfig.CronExpression, err))
			// 使用间隔时间
			app.useIntervalTimer()
		} else {
			app.cron.Start()
		}
	} else {
		// 使用间隔时间
		app.useIntervalTimer()
	}
}

// useIntervalTimer 使用间隔时间模式运行
func (app *App) useIntervalTimer() {
	// 初始化定时器
	app.ticker = time.NewTicker(time.Duration(app.interval) * time.Minute)
	done := app.done
	// 启动一个goroutine监听定时器事件
	go func() {
		for {
			select {
			case <-app.ticker.C:
				app.triggerCheck()
			case <-done:
				return // 收到停止信号，退出goroutine
			}
		}
	}()
}

// TriggerCheck 供外部调用的触发检测方法
func (app *App) TriggerCheck() {
	select {
	case app.checkChan <- struct{}{}:
		slog.Info("Manually triggered check")
	default:
		slog.Warn("Check is already in progress, skipping this trigger")
	}
}

// triggerCheck 内部检测方法
func (app *App) triggerCheck() {
	// 如果已经在检测中，直接返回
	if !app.checking.CompareAndSwap(false, true) {
		slog.Warn("Check is already in progress, skipping this check")
		return
	}
	defer app.checking.Store(false)

	if err := app.checkProxies(); err != nil {
		slog.Error(fmt.Sprintf("Failed to check proxies: %v", err))
		os.Exit(1)
	}

	// 检测完成后显示下次检查时间
	if app.ticker != nil {
		// 使用间隔时间模式
		app.ticker.Reset(time.Duration(app.interval) * time.Minute)
		nextCheck := time.Now().Add(time.Duration(app.interval) * time.Minute)
		slog.Info(fmt.Sprintf("Next check time: %s", nextCheck.Format("2006-01-02 15:04:05")))
	} else if app.cron != nil {
		// 使用cron模式
		entries := app.cron.Entries()
		if len(entries) > 0 {
			nextTime := entries[0].Next
			slog.Info(fmt.Sprintf("Next check time: %s", nextTime.Format("2006-01-02 15:04:05")))
		}
	}
	debug.FreeOSMemory()
}

// checkProxies 执行代理检测
func (app *App) checkProxies() error {
	slog.Info("Preparing to check proxies", "progress display", config.GlobalConfig.PrintProgress)

	results, err := check.Check()
	if err != nil {
		return fmt.Errorf("Failed to check proxies: %w", err)
	}
	// 将成功的节点添加到全局中，暂时内存保存
	if config.GlobalConfig.KeepSuccessProxies {
		for _, result := range results {
			if result.Proxy != nil {
				config.GlobalProxies = append(config.GlobalProxies, result.Proxy)
			}
		}
	}

	// 入库速度测试结果（简版，无订阅ID关联）
	for _, r := range results {
		var ip sql.NullString
		if r.IP != "" {
			ip = sql.NullString{String: r.IP, Valid: true}
		}
		var pjs sql.NullString
		if b, err := json.Marshal(r.Proxy); err == nil {
			pjs = sql.NullString{String: string(b), Valid: true}
		}
		_ = storage.SaveSpeedResult(context.Background(), sql.NullInt64{}, fmt.Sprint(r.Proxy["name"]), sql.NullInt64{}, float64(r.SpeedKBps), sql.NullFloat64{}, ip, pjs)
	}

	slog.Info("检测完成")
	save.SaveConfig(results)
	utils.SendNotify(len(results))
	utils.UpdateSubs()

	// 执行回调脚本
	utils.ExecuteCallback(len(results))

	return nil
}

func TempLog() string {
	return filepath.Join(os.TempDir(), "subcheck.log")
}
