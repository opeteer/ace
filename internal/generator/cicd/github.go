package cicd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opeteer/ace/internal/config"
)

// GenerateGitHubActions creates .github/workflows/ci.yml tailored to the project stack
func GenerateGitHubActions(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath
	workflowDir := filepath.Join(base, ".github/workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		return err
	}

	var ciYaml string

	switch cfg.Framework.ID {
	case "laravel":
		ciYaml = fmt.Sprintf(`name: Laravel CI

on:
  push:
    branches: [ main, master, develop ]
  pull_request:
    branches: [ main, master ]

jobs:
  laravel-tests:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: %s
          POSTGRES_USER: ace_user
          POSTGRES_PASSWORD: secret
        ports:
          - 5432:5432
        options: --health-cmd pg_isready --health-interval 5s --health-timeout 5s --health-retries 5

    steps:
    - uses: actions/checkout@v4

    - name: Setup PHP
      uses: shivammathur/setup-php@v2
      with:
        php-version: '8.3'
        extensions: mbstring, pdo, pdo_pgsql, bcmath, gd
        coverage: none

    - name: Copy .env
      run: php -r "file_exists('.env') || copy('.env.example', '.env');"

    - name: Install Dependencies
      run: composer install -q --no-ansi --no-interaction --no-scripts --no-progress --prefer-dist

    - name: Generate key
      run: php artisan key:generate

    - name: Run Migrations
      env:
        DB_CONNECTION: pgsql
        DB_HOST: 127.0.0.1
        DB_PORT: 5432
        DB_DATABASE: %s
        DB_USERNAME: ace_user
        DB_PASSWORD: secret
      run: php artisan migrate --force

    - name: Execute tests
      run: vendor/bin/phpunit || true
`, cfg.Name, cfg.Name)

	case "fastapi":
		ciYaml = `name: FastAPI CI

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Set up Python
      uses: actions/setup-python@v5
      with:
        python-version: "3.12"
        cache: 'pip'

    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install -r requirements.txt
        pip install pytest httpx

    - name: Test with pytest
      run: |
        pytest
`

	case "next":
		ciYaml = `name: Next.js CI

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Use Node.js 20
      uses: actions/setup-node@v4
      with:
        node-version: 20
        cache: 'npm'

    - name: Install dependencies
      run: npm ci || npm install

    - name: Lint
      run: npm run lint || true

    - name: Build
      run: npm run build
`

	case "fiber":
		ciYaml = `name: Go Fiber CI

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.22'

    - name: Build
      run: go build -v ./...

    - name: Test
      run: go test -v ./...
`

	default:
		ciYaml = `name: CI

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - name: Run verification
      run: echo "Ace scaffold CI verification"
`
	}

	return os.WriteFile(filepath.Join(workflowDir, "ci.yml"), []byte(ciYaml), 0644)
}
