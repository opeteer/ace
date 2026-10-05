package framework

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opeteer/ace/internal/config"
)

// ScaffoldNuxt generates a complete Nuxt 3 fullstack Vue application
func ScaffoldNuxt(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	dirs := []string{
		"pages",
		"components",
		"server/api",
		"public",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(base, d), 0755); err != nil {
			return err
		}
	}

	files := map[string]string{
		"package.json": fmt.Sprintf(`{
  "name": "%s",
  "private": true,
  "type": "module",
  "scripts": {
    "build": "nuxt build",
    "dev": "nuxt dev",
    "generate": "nuxt generate",
    "preview": "nuxt preview",
    "postinstall": "nuxt prepare"
  },
  "dependencies": {
    "nuxt": "^3.12.0",
    "vue": "^3.4.27",
    "vue-router": "^4.3.2"
  }
}
`, cfg.Name),

		"nuxt.config.ts": `// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-04-03',
  devtools: { enabled: true }
})
`,

		"app.vue": fmt.Sprintf(`<template>
  <div style="font-family: system-ui, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; border: 1px solid #ddd; border-radius: 8px;">
    <h1>{{ title }}</h1>
    <p>Status: <strong style="color: #00d7d7;">Online</strong></p>
    <p>Scaffolded with Ace CLI</p>
    <p>API Endpoint: <a href="/api/health">/api/health</a></p>
  </div>
</template>

<script setup>
const title = '%s'
</script>
`, cfg.Name),

		"server/api/health.ts": `export default defineEventHandler((event) => {
  return {
    status: 'healthy',
    framework: 'Nuxt 3',
    timestamp: new Date().toISOString()
  }
})
`,

		".env.example": fmt.Sprintf(`PORT=%d
NODE_ENV=development
`, cfg.Framework.DefaultPort),

		".gitignore": `.nuxt/
.output/
dist/
node_modules/
.env
.env.backup
.DS_Store
`,
	}

	for relPath, content := range files {
		targetFile := filepath.Join(base, relPath)
		_ = os.MkdirAll(filepath.Dir(targetFile), 0755)
		if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
			return err
		}
	}

	envExample, err := os.ReadFile(filepath.Join(base, ".env.example"))
	if err == nil {
		_ = os.WriteFile(filepath.Join(base, ".env"), envExample, 0644)
	}

	return nil
}
