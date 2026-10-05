package framework

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opeteer/ace/internal/config"
)

// ScaffoldNestJS generates a standard NestJS enterprise TypeScript architecture
func ScaffoldNestJS(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	dirs := []string{
		"src",
		"test",
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
  "scripts": {
    "build": "nest build",
    "start": "nest start",
    "dev": "nest start --watch",
    "start:debug": "nest start --debug --watch",
    "start:prod": "node dist/main"
  },
  "dependencies": {
    "@nestjs/common": "^10.3.9",
    "@nestjs/core": "^10.3.9",
    "@nestjs/platform-express": "^10.3.9",
    "reflect-metadata": "^0.2.2",
    "rxjs": "^7.8.1"
  },
  "devDependencies": {
    "@nestjs/cli": "^10.3.2",
    "@nestjs/schematics": "^10.1.1",
    "@types/express": "^4.17.21",
    "@types/node": "^20.14.0",
    "typescript": "^5.4.5"
  }
}
`, cfg.Name),

		"tsconfig.json": `{
  "compilerOptions": {
    "module": "commonjs",
    "declaration": true,
    "removeComments": true,
    "emitDecoratorMetadata": true,
    "experimentalDecorators": true,
    "allowSyntheticDefaultImports": true,
    "target": "ES2021",
    "sourceMap": true,
    "outDir": "./dist",
    "baseUrl": "./",
    "incremental": true,
    "skipLibCheck": true,
    "strictNullChecks": false,
    "noImplicitAny": false,
    "strictBindCallApply": false,
    "forceConsistentCasingInFileNames": false,
    "noFallthroughCasesInSwitch": false
  }
}
`,

		"nest-cli.json": `{
  "$schema": "https://json.schemastore.org/nest-cli",
  "collection": "@nestjs/schematics",
  "sourceRoot": "src"
}
`,

		"src/main.ts": fmt.Sprintf(`import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';

async function bootstrap() {
  const app = await NestFactory.create(AppModule);
  app.enableCors();
  const port = process.env.PORT || %d;
  await app.listen(port);
  console.log('>> NestJS server is running at http://localhost:' + port);
}
bootstrap();
`, cfg.Framework.DefaultPort),

		"src/app.module.ts": `import { Module } from '@nestjs/common';
import { AppController } from './app.controller';
import { AppService } from './app.service';

@Module({
  imports: [],
  controllers: [AppController],
  providers: [AppService],
})
export class AppModule {}
`,

		"src/app.controller.ts": `import { Controller, Get } from '@nestjs/common';
import { AppService } from './app.service';

@Controller()
export class AppController {
  constructor(private readonly appService: AppService) {}

  @Get()
  getHello(): object {
    return this.appService.getInfo();
  }

  @Get('api/health')
  getHealth(): object {
    return this.appService.getHealth();
  }
}
`,

		"src/app.service.ts": fmt.Sprintf(`import { Injectable } from '@nestjs/common';

@Injectable()
export class AppService {
  getInfo(): object {
    return {
      app: '%s',
      framework: 'NestJS',
      status: 'online',
      scaffolded_by: 'Ace CLI',
    };
  }

  getHealth(): object {
    return {
      status: 'healthy',
      timestamp: new Date().toISOString(),
    };
  }
}
`, cfg.Name),

		".env.example": fmt.Sprintf(`PORT=%d
NODE_ENV=development
`, cfg.Framework.DefaultPort),

		".gitignore": `node_modules/
dist/
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
