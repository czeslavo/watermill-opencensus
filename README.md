# watermill-opencensus
[![](https://godoc.org/github.com/czeslavo/watermill-opencensus?status.svg)](http://godoc.org/github.com/czeslavo/watermill-opencensus)
OpenCensus tracing for Watermill.

# Usage
```go
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}

	router.AddMiddleware(
		opencensus.TracingMiddleware,
		middleware.RandomFail(0.1),
	)
```