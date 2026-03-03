package oss

import (
	"QA-System/internal/global/config"

	"github.com/zjutjh/WeJH-SDK/cube"
)

func getConfig() cube.Config {
	return cube.Config{
		BaseURL:    config.Config.GetString("cube.baseUrl"),
		APIKey:     config.Config.GetString("cube.apiKey"),
		BucketName: config.Config.GetString("cube.bucketName"),
	}
}
