package framework

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opeteer/ace/internal/config"
)

// ScaffoldReact generates a modern client-side React 18 application powered by Vite & TypeScript
func ScaffoldReact(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	dirs := []string{
		"src/components",
		"src/assets",
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
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc -b && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "react": "^18.3.1",
    "react-dom": "^18.3.1"
  },
  "devDependencies": {
    "@types/react": "^18.3.3",
    "@types/react-dom": "^18.3.0",
    "@vitejs/plugin-react": "^4.3.0",
    "typescript": "^5.2.2",
    "vite": "^5.2.11"
  }
}
`, cfg.Name),

		"vite.config.ts": `import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
});
`,

		"tsconfig.json": `{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true
  },
  "include": ["src"]
}
`,

		"index.html": fmt.Sprintf(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>%s</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
`, cfg.Name),

		"src/main.tsx": `import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App.tsx';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
`,

		"src/App.tsx": fmt.Sprintf(`import { useState } from 'react';

function App() {
  const [count, setCount] = useState(0);

  return (
    <div style={{ fontFamily: 'system-ui, sans-serif', maxWidth: 600, margin: '50px auto', padding: 20, border: '1px solid #ddd', borderRadius: 8 }}>
      <h1>%s</h1>
      <p>Framework: <strong>React 18 + Vite (TypeScript)</strong></p>
      <p>Status: <strong style={{ color: '#00d7d7' }}>Online</strong></p>
      <p>Scaffolded with Ace CLI</p>
      <button 
        style={{ padding: '8px 16px', background: '#7d56f4', color: '#fff', border: 'none', borderRadius: 4, cursor: 'pointer' }}
        onClick={() => setCount((c) => c + 1)}
      >
        Count: {count}
      </button>
    </div>
  );
}

export default App;
`, cfg.Name),

		".env.example": fmt.Sprintf(`VITE_APP_TITLE=%s
`, cfg.Name),

		".gitignore": `node_modules/
dist/
dist-ssr/
*.local
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

// ScaffoldVue generates a modern client-side Vue 3 application powered by Vite & TypeScript
func ScaffoldVue(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	dirs := []string{
		"src/components",
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
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc -b && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "vue": "^3.4.29"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.0.5",
    "typescript": "^5.2.2",
    "vite": "^5.3.1",
    "vue-tsc": "^2.0.21"
  }
}
`, cfg.Name),

		"vite.config.ts": `import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
});
`,

		"tsconfig.json": `{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "module": "ESNext",
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "preserve",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true
  },
  "include": ["src/**/*.ts", "src/**/*.d.ts", "src/**/*.tsx", "src/**/*.vue"]
}
`,

		"index.html": fmt.Sprintf(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>%s</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/main.ts"></script>
  </body>
</html>
`, cfg.Name),

		"src/main.ts": `import { createApp } from 'vue';
import App from './App.vue';

createApp(App).mount('#app');
`,

		"src/App.vue": fmt.Sprintf(`<template>
  <div style="font-family: system-ui, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; border: 1px solid #ddd; border-radius: 8px;">
    <h1>{{ title }}</h1>
    <p>Framework: <strong>Vue 3 + Vite (TypeScript)</strong></p>
    <p>Status: <strong style="color: #00ff87;">Online</strong></p>
    <p>Scaffolded with Ace CLI</p>
    <button 
      style="padding: 8px 16px; background: #42b883; color: #fff; border: none; border-radius: 4px; cursor: pointer;"
      @click="count++"
    >
      Count: {{ count }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const title = '%s';
const count = ref(0);
</script>
`, cfg.Name),

		".env.example": fmt.Sprintf(`VITE_APP_TITLE=%s
`, cfg.Name),

		".gitignore": `node_modules/
dist/
dist-ssr/
*.local
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
