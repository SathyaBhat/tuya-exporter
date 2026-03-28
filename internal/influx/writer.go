package influx

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
)

type Writer struct {
	client      *influxdb3.Client
	database    string
	measurement string
}

func New(url, token, database, measurement string) (*Writer, error) {
	client, err := influxdb3.New(influxdb3.ClientConfig{
		Host:     url,
		Token:    token,
		Database: database,
	})
	if err != nil {
		return nil, fmt.Errorf("create influxdb client: %w", err)
	}

	return &Writer{
		client:      client,
		database:    database,
		measurement: measurement,
	}, nil
}

type DataPoint struct {
	DeviceID   string
	DeviceName string
	Code       string
	Value      interface{}
	Timestamp  time.Time
}

func (w *Writer) WritePoints(ctx context.Context, points []DataPoint) error {
	if len(points) == 0 {
		return nil
	}

	pts := make([]*influxdb3.Point, 0, len(points))
	for _, dp := range points {
		tags := map[string]string{
			"device_id": dp.DeviceID,
			"code":      dp.Code,
		}
		if dp.DeviceName != "" {
			tags["device_name"] = dp.DeviceName
		}
		p := influxdb3.NewPoint(w.measurement,
			tags,
			map[string]interface{}{
				"value": dp.Value,
			},
			dp.Timestamp,
		)
		pts = append(pts, p)
	}

	return w.client.WritePoints(ctx, pts)
}

func (w *Writer) Close() error {
	err := w.client.Close()
	if err != nil {
		slog.Error("error closing influxdb client", "error", err)
	}
	return err
}
