package exporter

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/SathyaBhat/tuya-exporter/internal/config"
	"github.com/SathyaBhat/tuya-exporter/internal/influx"
	"github.com/SathyaBhat/tuya-exporter/internal/tuya"
)

type Exporter struct {
	tuyaClient   *tuya.Client
	influxWriter *influx.Writer
	config       *config.Config
	deviceNames  map[string]string // device_id -> name
}

func New(cfg *config.Config) (*Exporter, error) {
	tuyaClient := tuya.NewClient(
		cfg.Tuya.BaseURL,
		cfg.Tuya.AccessID,
		cfg.Tuya.AccessKey,
	)

	var influxWriter *influx.Writer
	if !cfg.Influx.Skip {
		w, err := influx.New(
			cfg.Influx.Host,
			cfg.Influx.Token,
			cfg.Influx.Database,
			cfg.Influx.Measurement,
		)
		if err != nil {
			return nil, fmt.Errorf("create influx writer: %w", err)
		}
		influxWriter = w
	}

	return &Exporter{
		tuyaClient:   tuyaClient,
		influxWriter: influxWriter,
		config:       cfg,
		deviceNames:  make(map[string]string),
	}, nil
}

func (e *Exporter) Run(ctx context.Context) error {
	// Fetch device names once before starting the poll loop
	e.fetchDeviceNames(ctx)

	ticker := time.NewTicker(e.config.PollInterval)
	defer ticker.Stop()

	slog.Info("exporter started", "poll_interval", e.config.PollInterval)

	// Run first poll immediately
	if err := e.poll(ctx); err != nil {
		slog.Error("initial poll failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := e.poll(ctx); err != nil {
				slog.Error("poll failed", "error", err)
			}
		}
	}
}

func (e *Exporter) fetchDeviceNames(ctx context.Context) {
	for _, deviceID := range e.config.Tuya.DeviceIDs {
		info, err := e.tuyaClient.GetDeviceInfo(ctx, deviceID)
		if err != nil {
			slog.Warn("failed to get device name, will use device_id", "device_id", deviceID, "error", err)
			continue
		}
		e.deviceNames[deviceID] = info.Name
		slog.Info("resolved device name", "device_id", deviceID, "name", info.Name)
	}
}

func (e *Exporter) poll(ctx context.Context) error {
	slog.Debug("polling devices")

	allowed := make(map[string]struct{}, len(config.AllowedCodes))
	for _, c := range config.AllowedCodes {
		allowed[c] = struct{}{}
	}

	var points []influx.DataPoint
	timestamp := time.Now()

	for _, deviceID := range e.config.Tuya.DeviceIDs {
		status, err := e.tuyaClient.GetDeviceStatus(ctx, deviceID)
		if err != nil {
			slog.Error("failed to get device status", "device_id", deviceID, "error", err)
			continue
		}

		for _, item := range status {
			if _, ok := allowed[item.Code]; !ok {
				continue
			}
			points = append(points, influx.DataPoint{
				DeviceID:   deviceID,
				DeviceName: e.deviceNames[deviceID],
				Code:       item.Code,
				Value:      item.Value,
				Timestamp:  timestamp,
			})
			slog.Debug("collected data point", "device_id", deviceID, "code", item.Code, "value", item.Value)
		}
	}

	if len(points) > 0 {
		if e.influxWriter == nil {
			for _, p := range points {
				slog.Info("skip influx: data point", "device_id", p.DeviceID, "device_name", p.DeviceName, "code", p.Code, "value", p.Value, "timestamp", p.Timestamp)
			}
		} else {
			if err := e.influxWriter.WritePoints(ctx, points); err != nil {
				return fmt.Errorf("write points to influx: %w", err)
			}
			slog.Info("wrote points to influxdb", "count", len(points))
		}
	}

	return nil
}

func (e *Exporter) Close() error {
	if e.influxWriter == nil {
		return nil
	}
	return e.influxWriter.Close()
}
