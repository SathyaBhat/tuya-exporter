package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// AllowedCodes is the fixed set of data point codes collected from all devices.
var AllowedCodes = []string{"va_temperature", "va_humidity", "battery_percentage"}

type Config struct {
	Tuya         TuyaConfig
	Influx       InfluxConfig
	PollInterval time.Duration
}

type TuyaConfig struct {
	BaseURL   string
	AccessID  string
	AccessKey string
	DeviceIDs []string
}

type InfluxConfig struct {
	Host        string
	Token       string
	Database    string
	Measurement string
	Skip        bool
}

func Load() (*Config, error) {
	baseURL := getEnv("TUYA_BASE_URL", "https://openapi.tuyaus.com")
	accessID := os.Getenv("TUYA_ACCESS_ID")
	accessKey := os.Getenv("TUYA_ACCESS_KEY")
	deviceIDs := os.Getenv("TUYA_DEVICE_IDS")
	influxHost := os.Getenv("INFLUXDB_HOST")
	influxToken := os.Getenv("INFLUXDB_TOKEN")
	influxDB := os.Getenv("INFLUXDB_DATABASE")
	measurement := getEnv("INFLUXDB_MEASUREMENT", "temperature")
	pollStr := getEnv("POLL_INTERVAL", "5m")
	skipInflux := os.Getenv("SKIP_INFLUX") == "true"

	if accessID == "" {
		return nil, fmt.Errorf("TUYA_ACCESS_ID is required")
	}
	if accessKey == "" {
		return nil, fmt.Errorf("TUYA_ACCESS_KEY is required")
	}
	if deviceIDs == "" {
		return nil, fmt.Errorf("TUYA_DEVICE_IDS is required")
	}
	if !skipInflux {
		if influxHost == "" {
			return nil, fmt.Errorf("INFLUXDB_HOST is required")
		}
		if influxToken == "" {
			return nil, fmt.Errorf("INFLUXDB_TOKEN is required")
		}
		if influxDB == "" {
			return nil, fmt.Errorf("INFLUXDB_DATABASE is required")
		}
	}

	pollInterval, err := time.ParseDuration(pollStr)
	if err != nil {
		return nil, fmt.Errorf("invalid POLL_INTERVAL %q: %w", pollStr, err)
	}

	ids := strings.Split(deviceIDs, ",")
	for i := range ids {
		ids[i] = strings.TrimSpace(ids[i])
	}

	return &Config{
		Tuya: TuyaConfig{
			BaseURL:   baseURL,
			AccessID:  accessID,
			AccessKey: accessKey,
			DeviceIDs: ids,
		},
		Influx: InfluxConfig{
			Host:        influxHost,
			Token:       influxToken,
			Database:    influxDB,
			Measurement: measurement,
			Skip:        skipInflux,
		},
		PollInterval: pollInterval,
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
