package framework

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opeteer/ace/internal/config"
)

// ScaffoldGeneric generates a standard baseline folder structure and README for any framework
func ScaffoldGeneric(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	dirs := []string{
		"src",
		"tests",
		"docs",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(base, d), 0755); err != nil {
			return err
		}
	}

	readme := fmt.Sprintf("# %s\n\nThis project was scaffolded using [Ace CLI](https://github.com/opeteer/ace).\n\n## Stack Specifications\n\n- **Framework**: %s (%s)\n- **Category**: %s\n- **Primary Language**: %s\n\n## Getting Started\n\n1. Check configuration in `.env`\n2. Start the development server according to the %s ecosystem conventions.\n",
		cfg.Name, cfg.Framework.Name, cfg.Framework.ID, cfg.Framework.Category, cfg.Framework.Language, cfg.Framework.Name)

	if err := os.WriteFile(filepath.Join(base, "README.md"), []byte(readme), 0644); err != nil {
		return err
	}

	envExample := fmt.Sprintf("APP_NAME=%s\nAPP_ENV=development\nPORT=%d\n", cfg.Name, cfg.Framework.DefaultPort)

	_ = os.WriteFile(filepath.Join(base, ".env.example"), []byte(envExample), 0644)
	_ = os.WriteFile(filepath.Join(base, ".env"), []byte(envExample), 0644)
	_ = os.WriteFile(filepath.Join(base, ".gitignore"), []byte(".env\nnode_modules/\nvendor/\nbuild/\ndist/\n"), 0644)

	return nil
}
