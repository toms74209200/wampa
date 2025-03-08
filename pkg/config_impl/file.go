// Package config_impl implements OS dependent configuration loading
package config_impl

import (
	"fmt"
	"os"

	"github.com/toms74209200/wampa/pkg/config"
)

// LoadFromFile loads configuration from a file
func LoadFromFile(path string) (*config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("設定ファイル %s が見つかりません", path)
		}
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}

	return config.Parse(data)
}
