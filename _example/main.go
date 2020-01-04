package main

import (
	"context"
	"flag"
	"os"
	"time"

	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"go.opencensus.io/trace"

	opencensus "github.com/czeslavo/watermill-opencensus"
)

var (
	logger          = watermill.NewStdLogger(true, false)
	withStackdriver = flag.Bool("stackdriver", false, "Use Stackdriver exporter, GCP_PROJECT_ID needs to be set")
)

func init() {
	flag.Parse()
}

func main() {
	trace.RegisterExporter(jaegerExporter())
	if *withStackdriver {
		trace.RegisterExporter(stackdriverExporter())
	}

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}

	router.AddMiddleware(
		opencensus.TracingMiddleware,
		middleware.RandomFail(0.1),
	)
	pubsub := gochannel.NewGoChannel(gochannel.Config{}, logger)

	router.AddHandler("message_one_handler", "message_one", pubsub, "message_two", pubsub, func(msg *message.Message) ([]*message.Message, error) {
		logger.Debug("Received message one", nil)
		<-time.After(time.Second)

		span := trace.FromContext(msg.Context())
		if span == nil {
			return nil, nil
		}
		span.AddAttributes(trace.StringAttribute("time", time.Now().String()))

		msgTwo := message.NewMessage(watermill.NewULID(), nil)
		return []*message.Message{msgTwo}, nil
	})

	router.AddNoPublisherHandler("message_two_handler", "message_two", pubsub, func(msg *message.Message) error {
		logger.Debug("Received message two", nil)
		<-time.After(time.Second)

		span := trace.FromContext(msg.Context())
		if span == nil {
			logger.Debug("no span context in message two", watermill.LogFields{"ctx": msg.Context()})
			return nil
		}

		span.AddAttributes(trace.StringAttribute("time", time.Now().String()))

		return nil
	})

	go publishMessages(pubsub)

	if err := router.Run(context.Background()); err != nil {
		panic(err)
	}
}

func stackdriverExporter() *stackdriver.Exporter {
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID:         os.Getenv("GCP_PROJECT_ID"),
		MetricPrefix:      "tracing-demo",
		ReportingInterval: 60 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	return exporter
}

func jaegerExporter() *jaeger.Exporter {
	exporter, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint:     "localhost:6831",
		CollectorEndpoint: "http://localhost:14268/api/traces",
		ServiceName:       "tracing-demo",
	})
	if err != nil {
		panic(err)
	}

	return exporter
}

func publishMessages(pub message.Publisher) {
	for range time.Tick(2 * time.Second) {
		msg := message.NewMessage(watermill.NewULID(), nil)

		if err := pub.Publish("message_one", msg); err != nil {
			panic(err)
		}
	}
}
