package core

import (
	"path/filepath"
	"strings"
)

// knownTestSuffixes covers common test file patterns across languages.
var knownTestSuffixes = []string{
	"_test.go",  // Go
	".test.ts",  // TypeScript (jest/vitest)
	".test.js",  // JavaScript (jest/vitest)
	".test.tsx", // React TypeScript
	".test.jsx", // React JavaScript
	".spec.ts",  // TypeScript (jasmine/karma)
	".spec.js",  // JavaScript (jasmine/karma)
	"_test.py",  // Python (pytest)
	"_spec.rb",  // Ruby (RSpec)
	"_test.rs",  // Rust
	"Test.java", // Java (JUnit)
	"Tests.cs",  // C# (NUnit/xUnit)
}

// IsTestFile checks if a file path matches any known test file pattern.
func IsTestFile(path string) bool {
	base := filepath.Base(path)
	for _, suffix := range knownTestSuffixes {
		if strings.HasSuffix(base, suffix) {
			return true
		}
	}
	// Python prefix convention: test_*.py
	if strings.HasPrefix(base, "test_") && strings.HasSuffix(base, ".py") {
		return true
	}
	return false
}

// StripTestSuffix removes the test suffix/prefix from a filename, returning the base name.
// E.g., "auth_test.go" -> "auth", "test_auth.py" -> "auth", "AuthTest.java" -> "Auth"
func StripTestSuffix(filename string) string {
	for _, suffix := range knownTestSuffixes {
		if strings.HasSuffix(filename, suffix) {
			return strings.TrimSuffix(filename, suffix)
		}
	}
	// Python prefix: test_auth.py -> auth
	if strings.HasPrefix(filename, "test_") && strings.HasSuffix(filename, ".py") {
		name := strings.TrimPrefix(filename, "test_")
		return strings.TrimSuffix(name, ".py")
	}
	return filename
}

// camelToKebab converts CamelCase/snake_case to kebab-case for feature IDs.
// "UserService" -> "user-service", "auth_handler" -> "auth-handler", "config" -> "config"
func camelToKebab(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('-')
			}
			b.WriteByte(byte(r - 'A' + 'a'))
		} else if r == '_' {
			b.WriteByte('-')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// DefaultTestPatterns returns file glob patterns appropriate for a detected test runner.
func DefaultTestPatterns(runner string) []string {
	switch {
	case strings.Contains(runner, "vitest") || strings.Contains(runner, "jest"):
		return []string{"**/*.test.ts", "**/*.test.js", "**/*.test.tsx", "**/*.test.jsx"}
	case strings.Contains(runner, "mocha") || strings.Contains(runner, "jasmine"):
		return []string{"**/*.spec.ts", "**/*.spec.js"}
	case strings.Contains(runner, "pytest"):
		return []string{"**/test_*.py", "**/*_test.py"}
	case strings.Contains(runner, "go test"):
		return []string{"**/*_test.go"}
	case strings.Contains(runner, "rspec"):
		return []string{"**/*_spec.rb"}
	case strings.Contains(runner, "cargo"):
		return []string{"**/*_test.rs"}
	case strings.Contains(runner, "mvn") || strings.Contains(runner, "gradle"):
		return []string{"**/*Test.java"}
	case strings.Contains(runner, "dotnet"):
		return []string{"**/*Tests.cs"}
	default:
		return []string{"**/*_test.*", "**/*.test.*", "**/*.spec.*"}
	}
}
