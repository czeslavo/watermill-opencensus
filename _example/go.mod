module main.go

go 1.13

require (
	contrib.go.opencensus.io/exporter/jaeger v0.2.0
	contrib.go.opencensus.io/exporter/stackdriver v0.12.8
	github.com/ThreeDotsLabs/watermill v1.1.0
	github.com/czeslavo/watermill-opencensus v0.0.0-00010101000000-000000000000
	go.opencensus.io v0.22.2
)

replace github.com/czeslavo/watermill-opencensus => ../
