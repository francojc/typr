package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	// Word/Quote Mode
	Words  string `yaml:"words"`
	Quotes string `yaml:"quotes"`

	// Test Parameters
	N     int `yaml:"n"`
	G     int `yaml:"g"`
	Start int `yaml:"start"`
	W     int `yaml:"w"`
	T     int `yaml:"t"`

	// Display Options
	Theme       string `yaml:"theme"`
	ShowWpm     bool   `yaml:"showwpm"`
	NoTheme     bool   `yaml:"notheme"`
	BlockCursor bool   `yaml:"blockcursor"`
	Bold        bool   `yaml:"bold"`

	// Behavior Options
	NoSkip         bool `yaml:"noskip"`
	NoBackspace    bool `yaml:"nobackspace"`
	NoHighlight    bool `yaml:"nohighlight"`
	Highlight1     bool `yaml:"highlight1"`
	Highlight2     bool `yaml:"highlight2"`
	Raw            bool `yaml:"raw"`
	Multi          bool `yaml:"multi"`

	// Output Options
	Csv      bool   `yaml:"csv"`
	CsvDir   string `yaml:"csvdir"`
	Json     bool   `yaml:"json"`
	OneShot  bool   `yaml:"oneshot"`
	NoReport bool   `yaml:"noreport"`
}

var YAML_CONFIG_FILE string

func getDefaultConfig() AppConfig {
	return AppConfig{
		Words:       "1000en",
		Quotes:      "",
		N:           50,
		G:           1,
		Start:       -1,
		W:           80,
		T:           -1,
		Theme:       "default",
		ShowWpm:     false,
		NoTheme:     false,
		BlockCursor: false,
		Bold:        false,
		NoSkip:      false,
		NoBackspace: false,
		NoHighlight: false,
		Highlight1:  false,
		Highlight2:  false,
		Raw:         false,
		Multi:       false,
		Csv:         false,
		CsvDir:      "",
		Json:        false,
		OneShot:     false,
		NoReport:    false,
	}
}

func createDefaultConfigFile(configPath string) error {
	cfg := getDefaultConfig()

	// Marshal config to YAML
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add header comments to the generated YAML
	header := `# typr - Typing Test Configuration
#
# This file contains default settings for typr.
# Command-line flags override these settings.

`
	fullContent := header + string(data)

	// Write to file
	err = os.WriteFile(configPath, []byte(fullContent), 0600)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func loadConfig(configPath string) (*AppConfig, error) {
	cfg := getDefaultConfig()

	// Try to read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, return defaults
			return &cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal YAML into config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

func ensureConfigExists(configPath string) error {
	// Check if config file exists
	_, err := os.Stat(configPath)
	if err == nil {
		// File exists
		return nil
	}

	if !os.IsNotExist(err) {
		// Some other error occurred
		return fmt.Errorf("failed to check config file: %w", err)
	}

	// File doesn't exist, create it with defaults
	return createDefaultConfigFile(configPath)
}
