module github.com/Cyberax/slog-tidbits/otel

go 1.22

require (
	github.com/Cyberax/slog-tidbits v0.0.0-20210909123456-123456789012
	go.opentelemetry.io/otel/trace v1.27.0
)

require go.opentelemetry.io/otel v1.27.0 // indirect

replace github.com/Cyberax/slog-tidbits => ..
