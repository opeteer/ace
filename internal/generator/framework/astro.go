package framework

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opeteer/ace/internal/config"
)

// ScaffoldAstro generates a complete Astro islands content/web framework project
func ScaffoldAstro(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	dirs := []string{
		"src/pages/api",
		"src/components",
		"src/layouts",
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
  "type": "module",
  "version": "0.1.0",
  "scripts": {
    "dev": "astro dev",
    "start": "astro dev",
    "build": "astro check && astro build",
    "preview": "astro preview",
    "astro": "astro"
  },
  "dependencies": {
    "astro": "^4.10.0"
  },
  "devDependencies": {
    "@astrojs/check": "^0.7.0",
    "typescript": "^5.4.5"
  }
}
`, cfg.Name),

		"astro.config.mjs": `// @ts-check
import { defineConfig } from 'astro/config';

// https://astro.build/config
export default defineConfig({});
`,

		"tsconfig.json": `{
  "extends": "astro/tsconfigs/strict"
}
`,

		"src/pages/index.astro": fmt.Sprintf(`---
const appName = '%s';
---

<html lang="en">
	<head>
		<meta charset="utf-8" />
		<link rel="icon" type="image/svg+xml" href="/favicon.svg" />
		<meta name="viewport" content="width=device-width" />
		<title>{appName}</title>
	</head>
	<body style="font-family: system-ui, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; border: 1px solid #ddd; border-radius: 8px;">
		<h1>{appName}</h1>
		<p>Framework: <strong>Astro (Islands Architecture)</strong></p>
		<p>Status: <strong style="color: #7d56f4;">Online</strong></p>
		<p>Scaffolded by Ace CLI</p>
		<p>API Endpoint: <a href="/api/health">/api/health</a></p>
	</body>
</html>
`, cfg.Name),

		"src/pages/api/health.ts": `import type { APIRoute } from 'astro';

export const GET: APIRoute = async () => {
  return new Response(
    JSON.stringify({
      status: 'healthy',
      framework: 'Astro',
      timestamp: new Date().toISOString()
    }),
    {
      status: 200,
      headers: {
        'Content-Type': 'application/json'
      }
    }
  );
};
`,

		".env.example": fmt.Sprintf(`PORT=%d
NODE_ENV=development
`, cfg.Framework.DefaultPort),

		".gitignore": `dist/
node_modules/
.astro/
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
