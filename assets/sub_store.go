package assets

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/twj0/subcheck/save/method"
)

//go:embed sub-store.bundle.js.zst
var EmbeddedSubStore []byte

//go:embed ACL4SSR_Online_Full.yaml.zst
var EmbeddedOverrideYaml []byte

//go:embed clash_template.yaml
var EmbeddedClashTemplate []byte

func EnsureDefaultOutputFiles() error {
	saver, err := method.NewLocalSaver()
	if err != nil {
		return err
	}
	if !filepath.IsAbs(saver.OutputPath) {
		// 处理用户写相对路径的问题
		saver.OutputPath = filepath.Join(saver.BasePath, saver.OutputPath)
	}

	if err := os.MkdirAll(saver.OutputPath, 0755); err != nil {
		return fmt.Errorf("create output dir failed: %w", err)
	}

	tplPath := filepath.Join(saver.OutputPath, "clash_template.yaml")
	if _, err := os.Stat(tplPath); err == nil {
		slog.Debug("default template already exists", "file", tplPath)
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat clash_template.yaml failed: %w", err)
	}

	if err := os.WriteFile(tplPath, EmbeddedClashTemplate, 0644); err != nil {
		return fmt.Errorf("write clash_template.yaml failed: %w", err)
	}
	slog.Debug("default template created", "file", tplPath)
	return nil
}
