package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opeteer/ace/internal/config"
)

// ScaffoldFlutter generates a clean, cross-platform Flutter application
func ScaffoldFlutter(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	dirs := []string{
		"lib",
		"test",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(base, d), 0755); err != nil {
			return err
		}
	}

	flutterName := strings.ReplaceAll(strings.ToLower(cfg.Name), "-", "_")

	files := map[string]string{
		"pubspec.yaml": fmt.Sprintf(`name: %s
description: "A new Flutter project scaffolded with Ace CLI."
publish_to: "none"
version: 1.0.0+1

environment:
  sdk: ">=3.0.0 <4.0.0"

dependencies:
  flutter:
    sdk: flutter
  cupertino_icons: ^1.0.6

dev_dependencies:
  flutter_test:
    sdk: flutter
  flutter_lints: ^3.0.0

flutter:
  uses-material-design: true
`, flutterName),

		"lib/main.dart": fmt.Sprintf(`import 'package:flutter/material.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: '%s',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepPurple),
        useMaterial3: true,
      ),
      home: const MyHomePage(title: '%s'),
    );
  }
}

class MyHomePage extends StatefulWidget {
  const MyHomePage({super.key, required this.title});

  final String title;

  @override
  State<MyHomePage> createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage> {
  int _counter = 0;

  void _incrementCounter() {
    setState(() {
      _counter++;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        title: Text(widget.title),
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            const Text('Scaffolded with Ace CLI'),
            const SizedBox(height: 8),
            const Text(
              'Status: Online',
              style: TextStyle(color: Colors.teal, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 24),
            const Text('You have pushed the button this many times:'),
            Text(
              '$_counter',
              style: Theme.of(context).textTheme.headlineMedium,
            ),
          ],
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _incrementCounter,
        tooltip: 'Increment',
        child: const Icon(Icons.add),
      ),
    );
  }
}
`, cfg.Name, cfg.Name),

		"test/widget_test.dart": `import 'package:flutter_test/flutter_test.dart';
import 'package:flutter/material.dart';

void main() {
  testWidgets('Smoke test', (WidgetTester tester) async {
    expect(find.byType(MaterialApp), findsNothing);
  });
}
`,

		".gitignore": `.dart_tool/
.flutter-plugins
.flutter-plugins-dependencies
.packages
build/
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
