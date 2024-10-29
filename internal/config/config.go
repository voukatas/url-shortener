package config

import (
	"encoding/json"
	"github.com/voukatas/url-shortener/internal/model"
	"io"
	"os"
)

func LoadConfig(filename string) (*model.Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config model.Config
	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
