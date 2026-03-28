# Development Guide

## Project Structure

```
tuya-exporter/
├── cmd/
│   └── exporter/          # Main application entry point
│       └── main.go
├── internal/
│   ├── config/            # Configuration management
│   │   └── config.go
│   ├── exporter/          # Main exporter logic
│   │   └── exporter.go
│   ├── influx/            # InfluxDB client wrapper
│   │   └── writer.go
│   └── tuya/              # Tuya API client
│       ├── client.go
│       ├── token.go
│       └── types.go
├── Dockerfile             # Container image definition
├── docker-compose.yml     # Docker Compose setup for testing
├── go.mod / go.sum        # Go dependencies
└── README.md / QUICKSTART.md / DEVELOPMENT.md
```

## Building

```bash
# Build the binary
go build -o tuya-exporter ./cmd/exporter

# Build with specific output
GOOS=linux GOARCH=amd64 go build -o tuya-exporter-linux ./cmd/exporter
```

## Testing

```bash
# Run tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestFunctionName ./internal/package
```

## Code Style

This project follows Go conventions:
- Use `gofmt` for formatting
- Run `go vet` for static analysis
- Use meaningful variable/function names
- Add comments for exported functions

```bash
gofmt -w .
go vet ./...
```

## Adding Features

### Adding a New Configuration Option

1. Add field to relevant struct in `internal/config/config.go`
2. Load from environment variable in `Load()` function
3. Use in your feature code
4. Update `.env.example` with the new variable

### Adding a New Sensor Type

1. The exporter automatically handles all sensor codes from Tuya API
2. Data points are stored with their code as a tag
3. Query by code in InfluxDB

### Modifying Tuya API Interaction

1. Update functions in `internal/tuya/client.go`
2. Add any new response types to `internal/tuya/types.go`
3. Test with real Tuya device

### Modifying InfluxDB Write Format

Edit the `formatLineProtocol()` function in `internal/influx/writer.go` to change how data is formatted before sending to InfluxDB.

## Dependencies

Key dependencies:
- `github.com/InfluxCommunity/influxdb3-go/v2` - InfluxDB v3 client
- Standard library for HTTP, JSON, crypto operations

Update dependencies:
```bash
go get -u ./...
go mod tidy
```

## Docker Development

Build the image:
```bash
docker build -t tuya-exporter:dev .
```

Test with docker-compose:
```bash
docker-compose up --build
```

View logs:
```bash
docker-compose logs -f
```

Stop services:
```bash
docker-compose down
```

## Debugging

Enable debug logging by modifying `main.go`:
```go
// Change log level to debug
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
```

Or set at runtime (if implemented):
```bash
LOG_LEVEL=debug ./tuya-exporter
```

## Performance Considerations

1. **Token Caching**: Tokens are cached and only refreshed when near expiry
2. **Batch Writes**: Multiple device readings are batched into a single write to InfluxDB
3. **Connection Pooling**: HTTP client reuses connections
4. **Polling Interval**: Default 30s balances freshness vs API rate limits

## Common Issues During Development

### Import Issues
Make sure to run `go mod tidy` after adding dependencies.

### Type Errors in InfluxDB Client
The client requires pointers for mutable operations. Check method signatures carefully.

### Tuya API Errors
- Code 1010/1011: Token expired (handled automatically)
- Code 2001: Invalid parameters (check device ID)
- Code 1000: System error (retry)

## Contribution Workflow

1. Create a feature branch: `git checkout -b feature/my-feature`
2. Make changes and test locally
3. Run: `go fmt ./...`, `go vet ./...`
4. Commit with clear messages
5. Push and create a pull request

## Release Process

1. Update version in code if needed
2. Tag the commit: `git tag v1.0.0`
3. Push tag: `git push origin v1.0.0`
4. GitHub Actions automatically builds and pushes Docker image

## License

MIT
