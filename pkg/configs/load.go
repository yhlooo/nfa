package configs

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadConfig 加载配置
func LoadConfig(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, err
	}

	cfg := Config{}
	if err := json.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config from json error: %w", err)
	}

	return cfg, nil
}
