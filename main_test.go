package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/atotto/clipboard"
)

const testConfig = `
url: "http://localhost:8080"
source_lang: "auto"
target_lang: "ZH"
`

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, configFileName)

	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}

	if cfg.URL != "http://localhost:8080" {
		t.Errorf("Expected URL 'http://localhost:8080', got '%s'", cfg.URL)
	}
	if cfg.SourceLang != "auto" {
		t.Errorf("Expected source_lang 'auto', got '%s'", cfg.SourceLang)
	}
	if cfg.TargetLang != "ZH" {
		t.Errorf("Expected target_lang 'ZH', got '%s'", cfg.TargetLang)
	}
}

func TestTranslateText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req TranslationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Text != "hello" {
			t.Errorf("Expected text 'hello', got '%s'", req.Text)
		}
		if req.SourceLang != "auto" {
			t.Errorf("Expected source_lang 'auto', got '%s'", req.SourceLang)
		}
		if req.TargetLang != "ZH" {
			t.Errorf("Expected target_lang 'ZH', got '%s'", req.TargetLang)
		}

		response := TranslationResponse{
			Code: 200,
			Data: "你好",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	result, err := translateText("hello", "auto", "ZH", server.URL)
	if err != nil {
		t.Fatalf("translateText failed: %v", err)
	}
	if result != "你好" {
		t.Errorf("Expected '你好', got '%s'", result)
	}
}

func TestClipboardIntegration(t *testing.T) {
	testText := "clipboard test"
	if err := clipboard.WriteAll(testText); err != nil {
		t.Fatalf("clipboard.WriteAll failed: %v", err)
	}

	result, err := clipboard.ReadAll()
	if err != nil {
		t.Fatalf("clipboard.ReadAll failed: %v", err)
	}
	if result != testText {
		t.Errorf("Expected '%s', got '%s'", testText, result)
	}
}

func TestEndToEnd(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := TranslationResponse{
			Code: 200,
			Data: "模拟翻译结果",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "TestEndToEnd")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Manual cleanup
	configPath := filepath.Join(tempDir, configFileName)
	configContent := strings.Replace(testConfig, "http://localhost:8080", server.URL, 1)
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Build binary to avoid module downloads
	cmd := exec.Command("go", "build", "-o", filepath.Join(tempDir, "deeplx-cli"))
	if err := cmd.Run(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Execute built binary
	cmd = exec.Command(filepath.Join(tempDir, "deeplx-cli"),
		"-text", "Test text",
		"-url", server.URL,
		"-s", "auto",
		"-t", "ZH")

	// Set temp home directory
	cmd.Env = append(os.Environ(), "HOME="+tempDir)

	// Capture output
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// Execute command
	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, out.String())
	}

	// Verify output - check last line only
	output := strings.TrimSpace(out.String())
	lines := strings.Split(output, "\n")
	lastLine := ""
	if len(lines) > 0 {
		lastLine = strings.TrimSpace(lines[len(lines)-1])
	}

	if lastLine != "模拟翻译结果" {
		t.Errorf("Expected last output line '模拟翻译结果', got '%s'", lastLine)
	}

	// Verify clipboard
	clipContent, err := clipboard.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read clipboard: %v", err)
	}
	if clipContent != "模拟翻译结果" {
		t.Errorf("Expected clipboard content '模拟翻译结果', got '%s'", clipContent)
	}
}

