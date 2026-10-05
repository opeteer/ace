package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opeteer/ace/internal/config"
)

// ScaffoldSpringBoot generates a standard Spring Boot 3.3 Java enterprise application
func ScaffoldSpringBoot(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	pkgPath := "src/main/java/com/ace/app"
	dirs := []string{
		pkgPath + "/controller",
		"src/main/resources",
		"src/test/java/com/ace/app",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(base, d), 0755); err != nil {
			return err
		}
	}

	artifactId := strings.ReplaceAll(strings.ToLower(cfg.Name), "_", "-")

	files := map[string]string{
		"pom.xml": fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
	xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
	<modelVersion>4.0.0</modelVersion>
	<parent>
		<groupId>org.springframework.boot</groupId>
		<artifactId>spring-boot-starter-parent</artifactId>
		<version>3.3.0</version>
		<relativePath/>
	</parent>
	<groupId>com.ace</groupId>
	<artifactId>%s</artifactId>
	<version>0.1.0-SNAPSHOT</version>
	<name>%s</name>
	<description>Scaffolded with Ace CLI</description>
	<properties>
		<java.version>17</java.version>
	</properties>
	<dependencies>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-web</artifactId>
		</dependency>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-actuator</artifactId>
		</dependency>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-test</artifactId>
			<scope>test</scope>
		</dependency>
	</dependencies>
	<build>
		<plugins>
			<plugin>
				<groupId>org.springframework.boot</groupId>
				<artifactId>spring-boot-maven-plugin</artifactId>
			</plugin>
		</plugins>
	</build>
</project>
`, artifactId, cfg.Name),

		filepath.Join(pkgPath, "Application.java"): `package com.ace.app;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class Application {
	public static void main(String[] args) {
		SpringApplication.run(Application.class, args);
	}
}
`,

		filepath.Join(pkgPath, "controller/HealthController.java"): fmt.Sprintf(`package com.ace.app.controller;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;
import java.time.Instant;
import java.util.Map;

@RestController
public class HealthController {

    @GetMapping("/")
    public Map<String, Object> index() {
        return Map.of(
            "app", "%s",
            "framework", "Spring Boot 3",
            "status", "online",
            "scaffolded_by", "Ace CLI"
        );
    }

    @GetMapping("/api/health")
    public Map<String, Object> health() {
        return Map.of(
            "status", "healthy",
            "timestamp", Instant.now().toString()
        );
    }
}
`, cfg.Name),

		"src/main/resources/application.properties": fmt.Sprintf(`server.port=%d
spring.application.name=%s
management.endpoints.web.exposure.include=health,info
`, cfg.Framework.DefaultPort, cfg.Name),

		".env.example": fmt.Sprintf(`SERVER_PORT=%d
SPRING_PROFILES_ACTIVE=dev
`, cfg.Framework.DefaultPort),

		".gitignore": `target/
!.mvn/wrapper/maven-wrapper.jar
!**/src/main/**/target/
!**/src/test/**/target/
.idea
.DS_Store
*.iml
.env
.env.backup
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
