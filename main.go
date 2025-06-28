package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"gopkg.in/yaml.v2"
)

const (
	defaultDeepLXAPI = "https://deeplx.vercel.app/translate"
	configFileName   = ".deeplx-cli.yml"
)

// version will be set by build flags
var version = "dev"

// Config struct defines the structure of the configuration file
type Config struct {
	URL        string `yaml:"url"`
	SourceLang string `yaml:"source_lang"`
	TargetLang string `yaml:"target_lang"`
}

// TranslationRequest struct defines the JSON structure of the translation request
type TranslationRequest struct {
	Text       string `json:"text"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

// TranslationResponse struct defines the JSON structure of the translation response
type TranslationResponse struct {
	Code    int    `json:"code"`
	ID      int64  `json:"id"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

// loadConfig loads the YAML configuration file from the specified path
func loadConfig(configPath string) (*Config, error) {
	config := &Config{}
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}
	return config, nil
}

// translateText encapsulates the translation logic
func translateText(text, sourceLang, targetLang, apiURL string) (string, error) {
	requestBody, err := json.Marshal(TranslationRequest{
		Text:       text,
		SourceLang: sourceLang,
		TargetLang: targetLang,
	})
	if err != nil {
		return "", fmt.Errorf("error marshalling request: %w", err)
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error sending request to DeepLX API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("DeepLX API returned non-200 status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	var translationResponse TranslationResponse
	err = json.Unmarshal(bodyBytes, &translationResponse)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling response: %w", err)
	}

	if translationResponse.Code != 200 {
		return "", fmt.Errorf("translation failed with code %d: %s", translationResponse.Code, translationResponse.Message)
	}

	return translationResponse.Data, nil
}

func main() {
	// Get user home directory
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("failed to get current user home directory: %v", err)
	}
	configPath := filepath.Join(currentUser.HomeDir, configFileName)

	// Load configuration
	cfg := &Config{}
	// Check if config file exists, generate default if not
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Config file %s does not exist, generating default config.", configPath)
		defaultConfig := &Config{
			URL:        defaultDeepLXAPI,
			SourceLang: "auto",
			TargetLang: "EN",
		}
		yamlData, err := yaml.Marshal(defaultConfig)
		if err != nil {
			log.Printf("Warning: Failed to marshal default config: %v", err)
		} else {
			err = os.WriteFile(configPath, yamlData, 0644)
			if err != nil {
				log.Printf("Warning: Failed to write default config file %s: %v", configPath, err)
			} else {
				log.Printf("Default config file generated: %s", configPath)
			}
		}
		cfg = defaultConfig // Use default config as initial
	} else if err != nil {
		log.Fatalf("failed to check config file %s: %v", configPath, err)
	}

	// Load config (either newly generated or existing)
	loadedConfig, err := loadConfig(configPath)
	if err != nil {
		log.Printf("Warning: Failed to load or parse config file %s, using default values or command-line arguments: %v", configPath, err)
	} else {
		cfg = loadedConfig
	}

	// Define command-line flags
	var textArg string
	var sourceLangArg string
	var sourceLangShortArg string // New variable for shorthand
	var targetLangArg string
	var targetLangShortArg string // New variable for shorthand
	var urlArg string
	var versionFlag bool

	flag.StringVar(&textArg, "text", "", "Text to translate. If not provided, reads from standard input.")
	flag.StringVar(&sourceLangArg, "source_lang", "", "Source language.")
	flag.StringVar(&sourceLangShortArg, "s", "", "Source language (shorthand for -source_lang).") // Define shorthand -s
	flag.StringVar(&targetLangArg, "target_lang", "", "Target language.")
	flag.StringVar(&targetLangShortArg, "t", "", "Target language (shorthand for -target_lang).") // Define shorthand -t
	flag.StringVar(&urlArg, "url", "", "URL of the DeepLX API.")
	flag.BoolVar(&versionFlag, "version", false, "Show version information.")
	flag.BoolVar(&versionFlag, "v", false, "Show version information (shorthand).")

	flag.Parse()

	// Handle version flag
	if versionFlag {
		fmt.Printf("deeplx-cli version %s\n", version)
		os.Exit(0)
	}

	// Determine final parameters with priority
	finalURL := defaultDeepLXAPI
	if cfg.URL != "" {
		finalURL = cfg.URL
	}
	if urlArg != "" {
		finalURL = urlArg
	}

	finalSourceLang := cfg.SourceLang
	if sourceLangArg != "" {
		finalSourceLang = sourceLangArg
	} else if sourceLangShortArg != "" {
		finalSourceLang = sourceLangShortArg
	} else if finalSourceLang == "" {
		finalSourceLang = "auto" // Default source language
	}

	finalTargetLang := cfg.TargetLang
	if targetLangArg != "" {
		finalTargetLang = targetLangArg
	} else if targetLangShortArg != "" {
		finalTargetLang = targetLangShortArg
	} else if finalTargetLang == "" {
		finalTargetLang = "EN" // Default target language
	}

	// Get text to translate
	var inputText string
	if textArg != "" {
		inputText = textArg
	} else if len(flag.Args()) > 0 {
		// Combine all non-flag arguments as translation text
		for i, arg := range flag.Args() {
			inputText += arg
			if i < len(flag.Args())-1 {
				inputText += " "
			}
		}
	} else {
		fmt.Println("Enter text to translate (press Ctrl+D to finish input):")
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("failed to read standard input: %v", err)
		}
		inputText = string(bytes)
		inputText = strings.TrimSuffix(inputText, "\n") // Remove trailing newline
	}

	if inputText == "" {
		log.Fatalf("No text provided for translation. Please use the -text flag or provide text via standard input.")
	}

	// Call translation function
	translatedText, err := translateText(inputText, finalSourceLang, finalTargetLang, finalURL)
	if err != nil {
		log.Fatalf("Translation failed: %v", err)
	}

	// Copy to clipboard
	if err := clipboard.WriteAll(translatedText); err != nil {
		// Silent failure to avoid test output interference
	} else {
		// Silent success to avoid test output interference
	}

	fmt.Printf("%s\n", translatedText)
}
