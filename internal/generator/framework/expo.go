package framework

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opeteer/ace/internal/config"
)

// ScaffoldExpo generates a React Native with Expo universal mobile project
func ScaffoldExpo(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	dirs := []string{
		"assets",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(base, d), 0755); err != nil {
			return err
		}
	}

	files := map[string]string{
		"package.json": fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "main": "expo/AppEntry.js",
  "scripts": {
    "start": "expo start",
    "android": "expo start --android",
    "ios": "expo start --ios",
    "web": "expo start --web"
  },
  "dependencies": {
    "expo": "~51.0.0",
    "expo-status-bar": "~1.12.1",
    "react": "18.2.0",
    "react-native": "0.74.1"
  },
  "devDependencies": {
    "@babel/core": "^7.20.0",
    "@types/react": "~18.2.45",
    "typescript": "~5.3.3"
  },
  "private": true
}
`, cfg.Name),

		"app.json": fmt.Sprintf(`{
  "expo": {
    "name": "%s",
    "slug": "%s",
    "version": "1.0.0",
    "orientation": "portrait",
    "userInterfaceStyle": "light",
    "splash": {
      "resizeMode": "contain",
      "backgroundColor": "#ffffff"
    },
    "ios": {
      "supportsTablet": true
    },
    "android": {
      "adaptiveIcon": {
        "backgroundColor": "#ffffff"
      }
    },
    "web": {
      "bundler": "metro"
    }
  }
}
`, cfg.Name, cfg.Name),

		"App.tsx": fmt.Sprintf(`import { StatusBar } from 'expo-status-bar';
import { StyleSheet, Text, View } from 'react-native';

export default function App() {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>%s</Text>
      <Text style={styles.subtitle}>React Native with Expo</Text>
      <Text style={styles.status}>Status: Online</Text>
      <Text style={styles.credit}>Scaffolded with Ace CLI</Text>
      <StatusBar style="auto" />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#fff',
    alignItems: 'center',
    justifyContent: 'center',
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 16,
    color: '#666',
    marginBottom: 16,
  },
  status: {
    fontSize: 14,
    color: '#00d7d7',
    fontWeight: '600',
    marginBottom: 4,
  },
  credit: {
    fontSize: 12,
    color: '#999',
  },
});
`, cfg.Name),

		"babel.config.js": `module.exports = function(api) {
  api.cache(true);
  return {
    presets: ['babel-preset-expo'],
  };
};
`,

		"tsconfig.json": `{
  "extends": "expo/tsconfig.base"
}
`,

		".gitignore": `.expo/
dist/
web-build/
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

	return nil
}
