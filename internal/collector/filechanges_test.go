package collector

import "testing"

func TestGetLanguage(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"test.go", "go"},
		{"test.js", "javascript"},
		{"test.ts", "typescript"},
		{"test.tsx", "typescript"},
		{"test.jsx", "javascript"},
		{"test.py", "python"},
		{"test.rb", "ruby"},
		{"test.java", "java"},
		{"test.rs", "rust"},
		{"test.c", "c"},
		{"test.cpp", "cpp"},
		{"test.h", "c"},
		{"test.hpp", "cpp"},
		{"test.css", "css"},
		{"test.scss", "scss"},
		{"test.html", "html"},
		{"test.md", "markdown"},
		{"test.sql", "sql"},
		{"test.yaml", "yaml"},
		{"test.yml", "yaml"},
		{"test.json", "json"},
		{"test.xml", "xml"},
		{"test.toml", "toml"},
		{"test.swift", "swift"},
		{"test.php", "php"},
		{"test.dart", "dart"},
		{"test.unknown", ""},
		{"test", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := getLanguage(tt.path)
			if result != tt.expected {
				t.Errorf("getLanguage(%q) = %q; want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsCodeFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"test.go", true},
		{"test.js", true},
		{"test.ts", true},
		{"test.py", true},
		{"test.txt", false},
		{"test.jpg", false},
		{"test.png", false},
		{"test.pdf", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isCodeFile(tt.path)
			if result != tt.expected {
				t.Errorf("isCodeFile(%q) = %v; want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsBlacklistedDir(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"node_modules", true},
		{"vendor", true},
		{"dist", true},
		{"build", true},
		{"target", true},
		{".git", true},
		{".vscode", true},
		{".idea", true},
		{"Library", true},
		{"src", false},
		{"cmd", false},
		{"internal", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isBlacklistedDir(tt.path)
			if result != tt.expected {
				t.Errorf("isBlacklistedDir(%q) = %v; want %v", tt.path, result, tt.expected)
			}
		})
	}
}
