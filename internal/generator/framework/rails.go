package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opeteer/ace/internal/config"
)

// ScaffoldRails generates a standard Ruby on Rails 7 application structure
func ScaffoldRails(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	dirs := []string{
		"app/controllers",
		"app/models",
		"app/views",
		"config/environments",
		"bin",
		"db",
		"log",
		"tmp",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(base, d), 0755); err != nil {
			return err
		}
	}

	moduleName := toRailsModuleName(cfg.Name)

	files := map[string]string{
		"Gemfile": `source "https://rubygems.org"
git_source(:github) { |repo| "https://github.com/#{repo}.git" }

ruby ">= 3.0.0"

gem "rails", "~> 7.1.3"
gem "puma", ">= 5.0"
gem "bootsnap", require: false
gem "dotenv-rails", groups: [:development, :test]

group :development, :test do
  gem "debug", platforms: %i[ mri windows ]
end
`,

		"config.ru": `# This file is used by Rack-based servers to start the application.

require_relative "config/environment"

run Rails.application
Rails.application.load_server
`,

		"Rakefile": `require_relative "config/application"

Rails.application.load_tasks
`,

		"config/boot.rb": `ENV["BUNDLE_GEMFILE"] ||= File.expand_path("../Gemfile", __dir__)

require "bundler/setup" # Set up gems listed in the Gemfile.
require "bootsnap/setup" rescue LoadError # Speed up boot time
`,

		"config/environment.rb": `# Load the Rails application.
require_relative "application"

# Initialize the Rails application.
Rails.application.initialize!
`,

		"config/application.rb": fmt.Sprintf(`require_relative "boot"

require "rails"
require "active_model/railtie"
require "active_record/railtie" rescue LoadError
require "action_controller/railtie"

Bundler.require(*Rails.groups)

module %s
  class Application < Rails::Application
    config.load_defaults 7.1
    config.api_only = true
  end
end
`, moduleName),

		"config/routes.rb": `Rails.application.routes.draw do
  root to: "health#root"
  get "/api/health", to: "health#index"
end
`,

		"config/puma.rb": fmt.Sprintf(`max_threads_count = ENV.fetch("RAILS_MAX_THREADS") { 5 }
min_threads_count = ENV.fetch("RAILS_MIN_THREADS") { max_threads_count }
threads min_threads_count, max_threads_count

port ENV.fetch("PORT") { %d }
environment ENV.fetch("RAILS_ENV") { "development" }
pidfile ENV.fetch("PIDFILE") { "tmp/pids/server.pid" }

plugin :tmp_restart
`, cfg.Framework.DefaultPort),

		"app/controllers/application_controller.rb": `class ApplicationController < ActionController::API
end
`,

		"app/controllers/health_controller.rb": fmt.Sprintf(`class HealthController < ApplicationController
  def root
    render json: {
      app: "%s",
      framework: "Ruby on Rails 7",
      status: "online",
      scaffolded_by: "Ace CLI"
    }
  end

  def index
    render json: {
      status: "healthy",
      timestamp: Time.now.utc.iso8601
    }
  end
end
`, cfg.Name),

		".env.example": fmt.Sprintf(`PORT=%d
RAILS_ENV=development
RAILS_MAX_THREADS=5
`, cfg.Framework.DefaultPort),

		".gitignore": `/.bundle
/log/*
/tmp/*
!/log/.keep
!/tmp/.keep
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

func toRailsModuleName(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || r == ' ' || r == '.'
	})
	var b strings.Builder
	for _, p := range parts {
		if len(p) > 0 {
			b.WriteString(strings.ToUpper(p[:1]) + strings.ToLower(p[1:]))
		}
	}
	res := b.String()
	if res == "" {
		return "AceRailsApp"
	}
	return res
}
