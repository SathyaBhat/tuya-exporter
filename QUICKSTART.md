# Tuya Exporter - Quick Start Guide

## Overview

This exporter fetches temperature and humidity data from Tuya IoT devices and sends it to InfluxDB v3 for storage and visualization.

## Prerequisites

1. **Tuya Cloud Account** with API access
   - Sign up at https://iot.tuya.com/
   - Create a cloud project with Open API permissions
   - Generate API credentials (Access ID and Key)
   
2. **Tuya Device IDs**
   - Get device IDs from the Tuya Cloud console
   
3. **InfluxDB v3 Instance** (running separately)
   - Can be self-hosted or use InfluxDB Cloud

## Step 1: Get Tuya API Credentials

1. Log into [Tuya IoT Platform](https://iot.tuya.com/)
2. Go to Cloud → Projects
3. Create or select a project
4. In Project Overview, find your credentials:
   - **Access ID** (Client ID)
   - **Access Key** (Client Secret)
5. Find your **Device IDs** in the Devices section

## Step 2: Set Up Configuration

Copy `.env.example` to `.env` and fill in your values:

```bash
cp .env.example .env
```

Edit `.env`:
```bash
TUYA_ACCESS_ID=your_access_id
TUYA_ACCESS_KEY=your_access_key
TUYA_DEVICE_IDS=device1,device2
INFLUXDB_HOST=http://your-influxdb-host:8086
INFLUXDB_TOKEN=your_influxdb_token
INFLUXDB_DATABASE=tuya_data
```

## Step 3: Run the Exporter

### Option A: Docker Compose (with InfluxDB)

This starts both the exporter and InfluxDB:

```bash
docker-compose up -d
```

Check logs:
```bash
docker-compose logs -f tuya-exporter
```

### Option B: Docker (with external InfluxDB)

```bash
docker build -t tuya-exporter:latest .
docker run \
  -e TUYA_ACCESS_ID=your_access_id \
  -e TUYA_ACCESS_KEY=your_access_key \
  -e TUYA_DEVICE_IDS=device_id \
  -e INFLUXDB_HOST=http://influxdb-host:8086 \
  -e INFLUXDB_TOKEN=your_token \
  -e INFLUXDB_DATABASE=tuya_data \
  tuya-exporter:latest
```

### Option C: Run Locally

```bash
go run ./cmd/exporter
```

## Step 4: Verify Data in InfluxDB

Once running, data flows into InfluxDB. Query it with InfluxQL:

```sql
SELECT "value" FROM "tuya_env" 
WHERE "code" = 'temp_current' 
ORDER BY time DESC LIMIT 10
```

Or with SQL:
```sql
SELECT * FROM tuya_env 
WHERE code = 'temp_current' 
ORDER BY time DESC LIMIT 10
```

## Data Format

Each sensor reading is stored as:

- **Measurement**: `tuya_env` (configurable)
- **Tags**:
  - `device_id`: Device identifier
  - `code`: Sensor type (e.g., `temp_current`, `humidity_value`)
- **Fields**:
  - `value`: The sensor reading (numeric)
- **Timestamp**: Automatically set to collection time

### Common Tuya Sensor Codes

| Code | Meaning | Example Value |
|------|---------|---------------|
| `temp_current` | Temperature | 23.5 |
| `humidity_value` | Humidity | 45 |
| `battery_percentage` | Battery level | 85 |

## Troubleshooting

### "TUYA_ACCESS_ID is required"
Check that your `.env` file is properly set and all variables are exported.

### Token refresh errors
- Verify your Tuya Access Key is correct
- Check the API endpoint matches your region

### Connection to InfluxDB fails
- Verify `INFLUXDB_HOST` is accessible
- Check `INFLUXDB_TOKEN` has write permissions to the database
- Ensure the database exists in InfluxDB

### No data appearing in InfluxDB
- Check exporter logs: `docker-compose logs tuya-exporter`
- Verify device IDs are correct
- Ensure devices are online in Tuya Cloud
- Check that POLL_INTERVAL allows time for data collection

## Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `TUYA_BASE_URL` | No | `https://openapi.tuyaus.com` | Tuya API endpoint |
| `TUYA_ACCESS_ID` | Yes | — | Tuya API Access ID |
| `TUYA_ACCESS_KEY` | Yes | — | Tuya API Access Key |
| `TUYA_DEVICE_IDS` | Yes | — | Comma-separated device IDs |
| `INFLUXDB_HOST` | Yes | — | InfluxDB server URL |
| `INFLUXDB_TOKEN` | Yes | — | InfluxDB API token |
| `INFLUXDB_DATABASE` | Yes | — | InfluxDB database name |
| `INFLUXDB_MEASUREMENT` | No | `tuya_env` | InfluxDB measurement name |
| `POLL_INTERVAL` | No | `30s` | Polling frequency |

## Monitoring with Grafana

Once data is in InfluxDB, create a Grafana dashboard:

1. Add InfluxDB as a data source
2. Create a panel with query:
   ```
   SELECT "value" FROM "tuya_env" 
   WHERE "device_id" = 'your_device_id' 
   AND "code" = 'temp_current'
   ```
3. Visualize temperature trends

## Production Deployment

For production, consider:

- Running in Kubernetes with proper resource limits
- Using secrets management for credentials
- Setting up monitoring/alerts
- Configuring log aggregation
- Using a persistent volume for any local state

## Support

For issues:
1. Check the logs
2. Verify credentials and configuration
3. Test Tuya API access manually
4. Check InfluxDB connectivity

## License

MIT
