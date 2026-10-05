package framework

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/opeteer/ace/internal/config"
)

// ScaffoldLaravel generates a standard, production-ready Laravel 11 project structure
func ScaffoldLaravel(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	dirs := []string{
		"app/Http/Controllers",
		"app/Models",
		"app/Providers",
		"bootstrap",
		"config",
		"database/factories",
		"database/migrations",
		"database/seeders",
		"public",
		"resources/views",
		"routes",
		"storage/app",
		"storage/framework/cache/data",
		"storage/framework/sessions",
		"storage/framework/views",
		"storage/logs",
		"bootstrap/cache",
		"docker",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(base, d), 0755); err != nil {
			return err
		}
	}

	appKey := generateLaravelKey()

	// Files map
	files := map[string]string{
		"composer.json": fmt.Sprintf(`{
    "name": "ace/%s",
    "type": "project",
    "description": "Scaffolded with Ace CLI",
    "require": {
        "php": "^8.2",
        "laravel/framework": "^11.0",
        "laravel/tinker": "^2.9"
    },
    "require-dev": {
        "fakerphp/faker": "^1.23",
        "mockery/mockery": "^1.6",
        "nunomaduro/collision": "^8.0",
        "phpunit/phpunit": "^11.0"
    },
    "autoload": {
        "psr-4": {
            "App\\": "app/",
            "Database\\Factories\\": "database/factories/",
            "Database\\Seeders\\": "database/seeders/"
        }
    },
    "scripts": {
        "post-autoload-dump": [
            "Illuminate\\Foundation\\ComposerScripts::postAutoloadDump"
        ]
    },
    "config": {
        "optimize-autoloader": true,
        "preferred-install": "dist",
        "sort-packages": true,
        "audit": {
            "block": false
        }
    },
    "minimum-stability": "stable",
    "prefer-stable": true
}
`, cfg.Name),

		"artisan": `#!/usr/bin/env php
<?php

define('LARAVEL_START', microtime(true));

if (file_exists($maintenance = __DIR__.'/storage/framework/maintenance.php')) {
    require $maintenance;
}

require __DIR__.'/vendor/autoload.php';

$app = require_once __DIR__.'/bootstrap/app.php';

$status = $app->handleCommand(new Symfony\Component\Console\Input\ArgvInput);

exit($status);
`,

		"bootstrap/app.php": `<?php

use Illuminate\Foundation\Application;
use Illuminate\Foundation\Configuration\Exceptions;
use Illuminate\Foundation\Configuration\Middleware;

return Application::configure(basePath: dirname(__DIR__))
    ->withRouting(
        web: __DIR__.'/../routes/web.php',
        api: __DIR__.'/../routes/api.php',
        commands: __DIR__.'/../routes/console.php',
        health: '/up',
    )
    ->withMiddleware(function (Middleware $middleware) {
        //
    })
    ->withExceptions(function (Exceptions $exceptions) {
        //
    })->create();
`,

		"bootstrap/providers.php": `<?php

return [
    App\Providers\AppServiceProvider::class,
];
`,

		"app/Providers/AppServiceProvider.php": `<?php

namespace App\Providers;

use Illuminate\Support\ServiceProvider;

class AppServiceProvider extends ServiceProvider
{
    public function register(): void
    {
        //
    }

    public function boot(): void
    {
        //
    }
}
`,

		"app/Http/Controllers/Controller.php": `<?php

namespace App\Http\Controllers;

abstract class Controller
{
    //
}
`,

		"app/Models/User.php": `<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Foundation\Auth\User as Authenticatable;

class User extends Authenticatable
{
    use HasFactory;

    protected $fillable = [
        'name',
        'email',
        'password',
    ];

    protected $hidden = [
        'password',
        'remember_token',
    ];
}
`,

		"routes/web.php": `<?php

use Illuminate\Support\Facades\Route;

Route::get('/', function () {
    return response()->json([
        'app' => config('app.name'),
        'status' => 'online',
        'scaffolded_by' => 'Ace CLI',
    ]);
});
`,

		"routes/api.php": `<?php

use Illuminate\Support\Facades\Route;

Route::get('/v1/health', function () {
    return response()->json([
        'status' => 'healthy',
        'timestamp' => now()->toIso8601String(),
    ]);
});
`,

		"routes/console.php": `<?php

use Illuminate\Support\Facades\Artisan;

Artisan::command('inspire', function () {
    $this->comment('Simplicity is the soul of efficiency.');
})->purpose('Display an inspiring quote');
`,

		"public/index.php": `<?php

use Illuminate\Http\Request;

define('LARAVEL_START', microtime(true));

if (file_exists($maintenance = __DIR__.'/../storage/framework/maintenance.php')) {
    require $maintenance;
}

require __DIR__.'/../vendor/autoload.php';

(require_once __DIR__.'/../bootstrap/app.php')
    ->handleRequest(Request::capture());
`,

		"docker/entrypoint.sh": `#!/bin/sh
set -e

# If vendor/autoload.php is missing, install dependencies inside container
if [ ! -f "/var/www/html/vendor/autoload.php" ]; then
    echo ">> Ace: Installing Composer dependencies inside container..."
    composer install --no-interaction --prefer-dist --optimize-autoloader
fi

# Ensure storage and cache directories exist and are writable
mkdir -p storage/framework/cache/data storage/framework/sessions storage/framework/views storage/logs bootstrap/cache
chown -R www-data:www-data storage bootstrap/cache
chmod -R 775 storage bootstrap/cache

exec "$@"
`,

		"config/app.php": fmt.Sprintf(`<?php

return [
    'name' => env('APP_NAME', '%s'),
    'env' => env('APP_ENV', 'production'),
    'debug' => (bool) env('APP_DEBUG', false),
    'url' => env('APP_URL', 'http://localhost:8000'),
    'timezone' => 'UTC',
    'locale' => 'en',
    'fallback_locale' => 'en',
    'cipher' => 'AES-256-CBC',
    'key' => env('APP_KEY', '%s'),
];
`, cfg.Name, appKey),

		".gitignore": `/node_modules
/public/hot
/public/storage
/storage/*.key
/vendor
.env
.env.backup
.env.production
.phpunit.result.cache
Homestead.json
Homestead.yaml
auth.json
npm-debug.log
yarn-error.log
/.fleet
/.idea
/.vscode
`,

		".env.example": fmt.Sprintf(`APP_NAME=%s
APP_ENV=local
APP_KEY=%s
APP_DEBUG=true
APP_TIMEZONE=UTC
APP_URL=http://localhost:8000

APP_LOCALE=en
APP_FALLBACK_LOCALE=en

LOG_CHANNEL=stack
LOG_STACK=single
LOG_LEVEL=debug
`, cfg.Name, appKey),

		"storage/app/.gitkeep":                    "",
		"storage/framework/cache/.gitkeep":        "",
		"storage/framework/cache/data/.gitkeep":   "",
		"storage/framework/sessions/.gitkeep":     "",
		"storage/framework/views/.gitkeep":        "",
		"storage/logs/.gitkeep":                    "",
		"bootstrap/cache/.gitkeep":                "",
	}

	for relPath, content := range files {
		targetFile := filepath.Join(base, relPath)
		_ = os.MkdirAll(filepath.Dir(targetFile), 0755)
		if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
			return err
		}
	}

	// Make artisan and entrypoint executable
	_ = os.Chmod(filepath.Join(base, "artisan"), 0755)
	_ = os.Chmod(filepath.Join(base, "docker/entrypoint.sh"), 0755)

	// Create initial .env from .env.example
	envExample, err := os.ReadFile(filepath.Join(base, ".env.example"))
	if err == nil {
		_ = os.WriteFile(filepath.Join(base, ".env"), envExample, 0644)
	}

	return nil
}

func generateLaravelKey() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return "base64:" + base64.StdEncoding.EncodeToString(b)
}

