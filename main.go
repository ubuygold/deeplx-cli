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
	// 获取用户主目录
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("failed to get current user home directory: %v", err)
	}
	configPath := filepath.Join(currentUser.HomeDir, configFileName)

	// 加载配置文件
	cfg := &Config{}
	// 检查配置文件是否存在，如果不存在则生成默认配置文件
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
		cfg = defaultConfig // 使用默认配置作为初始配置
	} else if err != nil {
		log.Fatalf("failed to check config file %s: %v", configPath, err)
	}

	// 加载配置文件 (无论是新生成的还是已存在的)
	loadedConfig, err := loadConfig(configPath)
	if err != nil {
		log.Printf("Warning: Failed to load or parse config file %s, using default values or command-line arguments: %v", configPath, err)
	} else {
		cfg = loadedConfig
	}

	// 定义命令行参数
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

	// 处理版本标志
	if versionFlag {
		fmt.Printf("deeplx-cli version %s\n", version)
		os.Exit(0)
	}

	// 根据优先级确定最终参数
	finalURL := defaultDeepLXAPI
	if cfg.URL != "" {
		finalURL = cfg.URL
	}
	if urlArg != "" {
		finalURL = urlArg
	}

	finalSourceLang := cfg.SourceLang
	// Command-line arguments (long or shorthand) take precedence over config
	if sourceLangArg != "" {
		finalSourceLang = sourceLangArg
	} else if sourceLangShortArg != "" { // Check shorthand if long form is not provided
		finalSourceLang = sourceLangShortArg
	} else if finalSourceLang == "" {
		finalSourceLang = "auto" // 默认源语言
	}

	finalTargetLang := cfg.TargetLang
	// Command-line arguments (long or shorthand) take precedence over config
	if targetLangArg != "" {
		finalTargetLang = targetLangArg
	} else if targetLangShortArg != "" { // Check shorthand if long form is not provided
		finalTargetLang = targetLangShortArg
	} else if finalTargetLang == "" {
		finalTargetLang = "EN" // 默认目标语言
	}

	// 获取要翻译的文本
	var inputText string
	if textArg != "" {
		inputText = textArg
	} else if len(flag.Args()) > 0 {
		// 合并所有非标志参数作为翻译文本
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
		inputText = trimSuffix(inputText, "\n") // 移除末尾多余的换行符
	}

	if inputText == "" {
		log.Fatalf("No text provided for translation. Please use the -text flag or provide text via standard input.")
	}

	// 调用翻译函数
	translatedText, err := translateText(inputText, finalSourceLang, finalTargetLang, finalURL)
	if err != nil {
		log.Fatalf("Translation failed: %v", err)
	}

	fmt.Printf("%s\n", translatedText)
}

// trimSuffix removes the specified suffix from the end of a string
func trimSuffix(s, suffix string) string {
	if len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix {
		return s[:len(s)-len(suffix)]
	}
	return s
}
