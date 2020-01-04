module main.go

go 1.13

require (
	contrib.go.opencensus.io/exporter/jaeger v0.2.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.8 // indirect
	github.com/czeslavo/watermill-opencensus v0.0.0-00010101000000-000000000000 // indirect
)

replace github.com/czeslavo/watermill-opencensus => ../
