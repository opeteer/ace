package framework

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opeteer/ace/internal/config"
)

// ScaffoldExpress generates a production-ready TypeScript Express.js API
func ScaffoldExpress(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	dirs := []string{
		"src/controllers",
		"src/routes",
		"src/middleware",
		"src/services",
		"tests",
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
  "description": "Scaffolded with Ace CLI",
  "main": "dist/index.js",
  "scripts": {
    "dev": "tsx watch src/index.ts",
    "build": "tsc",
    "start": "node dist/index.js",
    "test": "vitest run"
  },
  "dependencies": {
    "cors": "^2.8.5",
    "dotenv": "^16.4.5",
    "express": "^4.19.2",
    "helmet": "^7.1.0",
    "zod": "^3.23.8"
  },
  "devDependencies": {
    "@types/cors": "^2.8.17",
    "@types/express": "^4.17.21",
    "@types/node": "^20.14.0",
    "tsx": "^4.11.0",
    "typescript": "^5.4.5",
    "vitest": "^1.6.0"
  }
}
`, cfg.Name),

		"tsconfig.json": `{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "esModuleInterop": true,
    "strict": true,
    "skipLibCheck": true,
    "outDir": "./dist",
    "rootDir": "./src"
  },
  "include": ["src/**/*"]
}
`,

		"src/index.ts": fmt.Sprintf(`import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import dotenv from 'dotenv';
import { healthRouter } from './routes/health.js';

dotenv.config();

const app = express();
const port = process.env.PORT || %d;

app.use(helmet());
app.use(cors());
app.use(express.json());

app.use('/api', healthRouter);

app.get('/', (req, res) => {
  res.json({
    app: '%s',
    framework: 'Express.js (TypeScript)',
    status: 'online',
    scaffolded_by: 'Ace CLI',
  });
});

app.listen(port, () => {
  console.log('>> Server is running at http://localhost:' + port);
});
`, cfg.Framework.DefaultPort, cfg.Name),

		"src/routes/health.ts": `import { Router } from 'express';

export const healthRouter = Router();

healthRouter.get('/health', (req, res) => {
  res.json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
  });
});
`,

		".env.example": fmt.Sprintf(`PORT=%d
NODE_ENV=development
`, cfg.Framework.DefaultPort),

		".gitignore": `node_modules/
dist/
.env
.env.backup
.DS_Store
coverage/
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
