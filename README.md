# watermill-opencensus
[![](https://godoc.org/github.com/czeslavo/watermill-opencensus?status.svg)](http://godoc.org/github.com/czeslavo/watermill-opencensus)
OpenCensus tracing for Watermill.

# Usage
```go
import (
    "contrib.go.opencensus.io/exporter/jaeger"
    "github.com/ThreeDotsLabs/watermill/message"
    "go.opencensus.io/trace"
    opencensus "github.com/czeslavo/watermill-opencensus"
)

...

// setup trace exporter (for example - jaeger)
exporter, _ := jaeger.NewExporter(jaeger.Options{
    AgentEndpoint:     "localhost:6831",
    CollectorEndpoint: "http://localhost:14268/api/traces",
    ServiceName:       "tracing-demo",
})

// create watermill router
router, _ := message.NewRouter(message.RouterConfig{}, logger)

// add opencensus middleware 
router.AddMiddleware(
    opencensus.TracingMiddleware,
)
```