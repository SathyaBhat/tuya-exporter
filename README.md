# Tuya Exporter

A Go-based exporter for Tuya temperature and humidity sensors to InfluxDB v3.

## Features

- Fetches temperature and humidity data from Tuya devices via their API
- Exports metrics to InfluxDB v3
- Configurable polling interval
- Automatic token refresh
- Docker containerized deployment
- Structured logging with JSON output

## Prerequisites

- Tuya Cloud API credentials (Access ID and Access Key)
- List of Tuya device IDs to monitor
- InfluxDB v3 instance running (or use docker-compose)

## Environment Variables

### Tuya Configuration
- `TUYA_BASE_URL` - Tuya API base URL (default: `https://openapi.tuyaus.com`)
- `TUYA_ACCESS_ID` - Tuya API Access ID (required)
- `TUYA_ACCESS_KEY` - Tuya API Access Key (required)
- `TUYA_DEVICE_IDS` - Comma-separated list of device IDs (required)

### InfluxDB Configuration
- `INFLUXDB_HOST` - InfluxDB host URL (required, e.g., `http://localhost:8086`)
- `INFLUXDB_TOKEN` - InfluxDB API token (required)
- `INFLUXDB_DATABASE` - InfluxDB database name (required)
- `INFLUXDB_MEASUREMENT` - InfluxDB measurement name (default: `tuya_env`)

### Exporter Configuration
- `POLL_INTERVAL` - Polling interval (default: `30s`, e.g., `1m`, `5s`)

## Quick Start with Docker Compose

1. Create a `.env` file with your configuration:
```bash
TUYA_ACCESS_ID=your_access_id
TUYA_ACCESS_KEY=your_access_key
TUYA_DEVICE_IDS=device_id_1,device_id_2
INFLUXDB_TOKEN=your_influx_token
INFLUXDB_DATABASE=tuya_data
INFLUXDB_ADMIN_USER=admin
INFLUXDB_ADMIN_PASSWORD=adminpassword
```

2. Start the services:
```bash
docker-compose up -d
```

3. View logs:
```bash
docker-compose logs -f tuya-exporter
```

## Building from Source

```bash
go mod download
go build -o tuya-exporter ./cmd/exporter
```

## Running the Exporter

```bash
./tuya-exporter
```

## Docker Build

```bash
docker build -t tuya-exporter:latest .
docker run \
  -e TUYA_ACCESS_ID=your_access_id \
  -e TUYA_ACCESS_KEY=your_access_key \
  -e TUYA_DEVICE_IDS=device_id \
  -e INFLUXDB_HOST=http://influxdb:8086 \
  -e INFLUXDB_TOKEN=your_token \
  -e INFLUXDB_DATABASE=tuya_data \
  tuya-exporter:latest
```

## Data Points

Each device status code is written as a separate data point with:
- **Measurement**: Configured measurement name (default: `tuya_env`)
- **Tags**:
  - `device_id`: The Tuya device ID
  - `code`: The status code (e.g., `temp_current`, `humidity_value`)
- **Fields**:
  - `value`: The numeric value from the sensor
- **Timestamp**: The time when the data was collected

### Example InfluxQL Query

```sql
SELECT "value" FROM "tuya_env" 
WHERE "code" = 'temp_current' 
AND time > now() - 1h
```

## Troubleshooting

### Token Refresh Issues
If you see errors like "code=1010", the token has expired. The exporter automatically refreshes tokens, but ensure your credentials are correct.

### Connection Issues
- Verify `INFLUXDB_HOST` is accessible and InfluxDB is running
- Check `INFLUXDB_TOKEN` and `INFLUXDB_DATABASE` are correct

### Device Not Found
- Verify device IDs are correct in `TUYA_DEVICE_IDS`
- Check device credentials and permissions in Tuya Cloud

## License

MIT
