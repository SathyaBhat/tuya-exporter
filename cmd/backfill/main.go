package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/SathyaBhat/tuya-exporter/internal/config"
	"github.com/SathyaBhat/tuya-exporter/internal/influx"
	"github.com/SathyaBhat/tuya-exporter/internal/tuya"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	var (
		fromStr = flag.String("from", "", "Start time (RFC3339, e.g. 2026-03-01T00:00:00Z)")
		toStr   = flag.String("to", "", "End time (RFC3339, defaults to now)")
		dryRun  = flag.Bool("dry-run", false, "Print data points without writing to InfluxDB")
	)
	flag.Parse()

	if *fromStr == "" {
		slog.Error("--from is required")
		flag.Usage()
		os.Exit(1)
	}

	fromTime, err := time.Parse(time.RFC3339, *fromStr)
	if err != nil {
		slog.Error("invalid --from time", "error", err)
		os.Exit(1)
	}

	toTime := time.Now()
	if *toStr != "" {
		toTime, err = time.Parse(time.RFC3339, *toStr)
		if err != nil {
			slog.Error("invalid --to time", "error", err)
			os.Exit(1)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	tuyaClient := tuya.NewClient(cfg.Tuya.BaseURL, cfg.Tuya.AccessID, cfg.Tuya.AccessKey)

	var influxWriter *influx.Writer
	if !*dryRun {
		influxWriter, err = influx.New(cfg.Influx.Host, cfg.Influx.Token, cfg.Influx.Database, cfg.Influx.Measurement)
		if err != nil {
			slog.Error("failed to create influx writer", "error", err)
			os.Exit(1)
		}
		defer influxWriter.Close()
	}

	ctx := context.Background()
	startMs := fromTime.UnixMilli()
	endMs := toTime.UnixMilli()

	// Resolve device names upfront
	deviceNames := make(map[string]string)
	for _, deviceID := range cfg.Tuya.DeviceIDs {
		info, err := tuyaClient.GetDeviceInfo(ctx, deviceID)
		if err != nil {
			slog.Warn("failed to get device name, will use device_id", "device_id", deviceID, "error", err)
			continue
		}
		deviceNames[deviceID] = info.Name
		slog.Info("resolved device name", "device_id", deviceID, "name", info.Name)
	}

	slog.Info("starting backfill",
		"from", fromTime.Format(time.RFC3339),
		"to", toTime.Format(time.RFC3339),
		"devices", len(cfg.Tuya.DeviceIDs),
		"codes", config.AllowedCodes,
		"dry_run", *dryRun,
	)

	totalPoints := 0

	for _, deviceID := range cfg.Tuya.DeviceIDs {
		slog.Info("backfilling device", "device_id", deviceID, "name", deviceNames[deviceID])

		points, err := fetchAllLogs(ctx, tuyaClient, deviceID, deviceNames[deviceID], startMs, endMs)
		if err != nil {
			slog.Error("failed to fetch logs", "device_id", deviceID, "error", err)
			continue
		}

		slog.Info("fetched log entries", "device_id", deviceID, "count", len(points))

		if *dryRun {
			for _, p := range points {
				slog.Info("dry-run: data point",
					"device_id", p.DeviceID,
					"device_name", p.DeviceName,
					"code", p.Code,
					"value", p.Value,
					"timestamp", p.Timestamp,
				)
			}
		} else if len(points) > 0 {
			if err := influxWriter.WritePoints(ctx, points); err != nil {
				slog.Error("failed to write points", "device_id", deviceID, "error", err)
				continue
			}
			slog.Info("wrote points to influxdb", "device_id", deviceID, "count", len(points))
		}

		totalPoints += len(points)
	}

	slog.Info("backfill complete", "total_points", totalPoints)
}

func fetchAllLogs(ctx context.Context, client *tuya.Client, deviceID, deviceName string, startMs, endMs int64) ([]influx.DataPoint, error) {
	var allPoints []influx.DataPoint
	nextRowKey := ""

	for {
		result, err := client.GetDeviceLogs(ctx, deviceID, startMs, endMs, config.AllowedCodes, nextRowKey)
		if err != nil {
			return nil, fmt.Errorf("get device logs: %w", err)
		}

		for _, log := range result.Logs {
			if log.Code == "" || log.Value == nil {
				continue
			}
			allPoints = append(allPoints, influx.DataPoint{
				DeviceID:   deviceID,
				DeviceName: deviceName,
				Code:       log.Code,
				Value:      log.Value,
				Timestamp:  time.UnixMilli(log.EventTime),
			})
		}

		if !result.HasNext || result.NextRowKey == "" {
			break
		}
		nextRowKey = result.NextRowKey

		// Be polite to the API
		time.Sleep(200 * time.Millisecond)
	}

	return allPoints, nil
}
