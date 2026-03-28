package influx

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
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
		v, err := toFloat(dp.Value)
		if err != nil {
			slog.Warn("skipping data point with non-numeric value", "device_id", dp.DeviceID, "code", dp.Code, "value", dp.Value)
			continue
		}
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
				"value": v,
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

// toFloat converts a value returned by the Tuya API to float64.
// The status endpoint returns numeric types; the logs endpoint returns strings.
func toFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("unsupported type %T", v)
	}
}
