package ipcheck

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/beck-8/subs-check/config"
	"github.com/beck-8/subs-check/utils"
)

type Result map[string]any

func Run(ctx context.Context, proxy string) (Result, error) {
	cfg := config.GlobalConfig.IpCheck
	if cfg.ScriptPath == "" {
		return nil, errors.New("ip-check.script-path is empty")
	}

	scriptPath := cfg.ScriptPath
	if !filepath.IsAbs(scriptPath) {
		execDir := utils.GetExecutablePath()
		primary := filepath.Join(execDir, scriptPath)
		if _, err := os.Stat(primary); err == nil {
			scriptPath = primary
		} else {
			fallback := filepath.Join(execDir, "..", scriptPath)
			if _, err2 := os.Stat(fallback); err2 == nil {
				scriptPath = fallback
			} else {
				scriptPath = primary
			}
		}
	}

	if _, err := os.Stat(scriptPath); err != nil {
		return nil, fmt.Errorf("ip.sh not found: %w", err)
	}

	args := []string{scriptPath, "-E", "-j", "-n", "-f"}
	if proxy != "" {
		args = append(args, "-x", proxy)
	}

	cmdName := "bash"
	if _, err := exec.LookPath(cmdName); err != nil {
		return nil, fmt.Errorf("bash not found in PATH; please install Git Bash or provide a full bash path: %w", err)
	}

	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Dir = filepath.Dir(scriptPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if deadline, ok := ctx.Deadline(); !ok {
		timeout := time.Duration(cfg.Timeout) * time.Second
		if timeout <= 0 {
			timeout = 300 * time.Second
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
		cmd = exec.CommandContext(ctx, cmdName, args...)
		cmd.Dir = filepath.Dir(scriptPath)
	}

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ip.sh failed: %w; stderr: %s", err, strings.TrimSpace(stderr.String()))
	}

	out := strings.TrimSpace(stdout.String())
	if out == "" {
		return nil, errors.New("empty output from ip.sh")
	}

	var res Result
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		idx := strings.Index(out, "{")
		if idx >= 0 {
			trim := out[idx:]
			if err2 := json.Unmarshal([]byte(trim), &res); err2 == nil {
				return res, nil
			}
		}
		return nil, fmt.Errorf("parse json failed: %w; raw: %s", err, out)
	}
	return res, nil
}

func RunDefault(ctx context.Context) (Result, error) {
	return Run(ctx, "")
}

type Core struct {
	IP          string
	FraudScore  int
	RiskLevel   string
	IsProxy     bool
	IsVPN       bool
	IsTor       bool
	CountryCode string
}

func ExtractCore(res Result) Core {
	var core Core
	getMap := func(v any) map[string]any {
		if m, ok := v.(map[string]any); ok {
			return m
		}
		return map[string]any{}
	}
	getArr := func(m map[string]any, key string) []any {
		if a, ok := m[key].([]any); ok {
			return a
		}
		return nil
	}
	getStr := func(v any) string {
		if s, ok := v.(string); ok {
			return s
		}
		return ""
	}
	getBool := func(v any) bool {
		if b, ok := v.(bool); ok {
			return b
		}
		if s, ok := v.(string); ok {
			switch strings.ToLower(s) {
			case "true", "1", "yes":
				return true
			}
		}
		return false
	}

	if head := getArr(res, "Head"); len(head) > 0 {
		m := getMap(head[0])
		core.IP = getStr(m["IP"])
	}

	if info := getArr(res, "Info"); len(info) > 0 {
		m := getMap(info[0])
		if region, ok := m["Region"]; ok {
			rm := getMap(region)
			core.CountryCode = getStr(rm["Code"])
		}
	}

	if score := getArr(res, "Score"); len(score) > 0 {
		m := getMap(score[0])
		pick := func(keys ...string) string {
			for _, k := range keys {
				if v, ok := m[k]; ok {
					s := getStr(v)
					if s != "" && strings.ToLower(s) != "null" {
						return s
					}
				}
			}
			return ""
		}
		s := pick("SCAMALYTICS", "IPQS", "IP2LOCATION", "ipapi", "DBIP")
		if s != "" {
			if n, err := strconv.Atoi(s); err == nil {
				core.FraudScore = n
			}
		}
		switch {
		case core.FraudScore <= 10:
			core.RiskLevel = "VeryLow"
		case core.FraudScore <= 25:
			core.RiskLevel = "Low"
		case core.FraudScore <= 50:
			core.RiskLevel = "Medium"
		case core.FraudScore <= 75:
			core.RiskLevel = "High"
		default:
			if core.FraudScore > 0 {
				core.RiskLevel = "VeryHigh"
			} else {
				core.RiskLevel = "Unknown"
			}
		}
	}

	if factor := getArr(res, "Factor"); len(factor) > 0 {
		fm := getMap(factor[0])
		anyTrue := func(section string) bool {
			if v, ok := fm[section]; ok {
				m := getMap(v)
				for _, vv := range m {
					if getBool(vv) {
						return true
					}
				}
			}
			return false
		}
		core.IsProxy = anyTrue("Proxy")
		core.IsVPN = anyTrue("VPN")
		core.IsTor = anyTrue("Tor")
	}
	return core
}
