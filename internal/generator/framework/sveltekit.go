package framework

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opeteer/ace/internal/config"
)

// ScaffoldSvelteKit generates a complete SvelteKit fullstack application
func ScaffoldSvelteKit(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	dirs := []string{
		"src/routes/api/health",
		"src/lib",
		"static",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(base, d), 0755); err != nil {
			return err
		}
	}

	files := map[string]string{
		"package.json": fmt.Sprintf(`{
  "name": "%s",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite dev",
    "build": "vite build",
    "preview": "vite preview",
    "check": "svelte-kit sync && svelte-check --tsconfig ./tsconfig.json",
    "check:watch": "svelte-kit sync && svelte-check --tsconfig ./tsconfig.json --watch"
  },
  "devDependencies": {
    "@sveltejs/adapter-auto": "^3.0.0",
    "@sveltejs/kit": "^2.0.0",
    "@sveltejs/vite-plugin-svelte": "^3.0.0",
    "svelte": "^4.2.7",
    "svelte-check": "^3.6.0",
    "typescript": "^5.0.0",
    "vite": "^5.0.3"
  }
}
`, cfg.Name),

		"svelte.config.js": `import adapter from '@sveltejs/adapter-auto';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),
  kit: {
    adapter: adapter()
  }
};

export default config;
`,

		"vite.config.ts": `import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [sveltekit()]
});
`,

		"tsconfig.json": `{
  "extends": "./.svelte-kit/tsconfig.json",
  "compilerOptions": {
    "allowJs": true,
    "checkJs": true,
    "esModuleInterop": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "skipLibCheck": true,
    "sourceMap": true,
    "strict": true,
    "moduleResolution": "bundler"
  }
}
`,

		"src/app.html": `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    %sveltekit.head%
  </head>
  <body data-sveltekit-preload-data="hover">
    <div style="display: contents">%sveltekit.body%</div>
  </body>
</html>
`,

		"src/routes/+page.svelte": fmt.Sprintf(`<script lang="ts">
  let appName = '%s';
</script>

<main style="font-family: system-ui, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; border: 1px solid #ddd; border-radius: 8px;">
  <h1>{appName}</h1>
  <p>Framework: <strong>SvelteKit</strong></p>
  <p>Status: <strong style="color: #00ff87;">Online</strong></p>
  <p>Scaffolded by Ace CLI</p>
  <p>API Endpoint: <a href="/api/health">/api/health</a></p>
</main>
`, cfg.Name),

		"src/routes/api/health/+server.ts": `import { json } from '@sveltejs/kit';

export async function GET() {
  return json({
    status: 'healthy',
    framework: 'SvelteKit',
    timestamp: new Date().toISOString()
  });
}
`,

		".env.example": fmt.Sprintf(`PORT=%d
NODE_ENV=development
`, cfg.Framework.DefaultPort),

		".gitignore": `.svelte-kit/
build/
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
