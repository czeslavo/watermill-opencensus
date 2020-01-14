package opencensus_test

import (
	"context"
	"testing"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/trace"

	opencensus "github.com/czeslavo/watermill-opencensus"
)

func TestTracingMiddleware(t *testing.T) {
	msg := message.NewMessage(watermill.NewULID(), nil)
	ctx, parentSpan := trace.StartSpan(context.Background(), "parent")
	msg.SetContext(ctx)

	tracingMiddleware := opencensus.TracingMiddleware(func(m *message.Message) ([]*message.Message, error) {
		spanCtx := trace.FromContext(m.Context())
		require.NotNil(t, spanCtx)
		require.Equal(t, parentSpan.SpanContext().TraceID, spanCtx.SpanContext().TraceID, "span context should have the same trace id as the parent's one")
		require.NotEqual(t, parentSpan.SpanContext().SpanID, spanCtx.SpanContext().SpanID, "span context should have different span id than parent")

		return []*message.Message{sampleMsg()}, nil
	})

	producedMsgs, err := tracingMiddleware(msg)
	require.NoError(t, err)
	require.NotEmpty(t, producedMsgs)

	for _, producedMsg := range producedMsgs {
		spanContext, ok := opencensus.GetSpanContext(producedMsg)
		require.True(t, ok)
		assert.Equal(t, parentSpan.SpanContext().TraceID, spanContext.TraceID, "produced msg should have same trace id as the parent's one")
		assert.NotEqual(t, parentSpan.SpanContext().SpanID, spanContext.SpanID, "produced msg should have span ID different than parent")
	}
}

func TestTracingMiddleware_no_parent_span(t *testing.T) {
	tracingMiddleware := opencensus.TracingMiddleware(func(m *message.Message) ([]*message.Message, error) {
		require.NotNil(t, trace.FromContext(m.Context()), "new span context should be propagated when there's no parent")
		return nil, nil
	})
	msg := message.NewMessage(watermill.NewULID(), nil)

	_, err := tracingMiddleware(msg)
	require.NoError(t, err)
}

func sampleMsg() *message.Message {
	return message.NewMessage(watermill.NewULID(), nil)
}

func TestPublisherDecorator_propagate_span_context(t *testing.T) {
	logger := watermill.NewStdLogger(true, false)
	pubsub := gochannel.NewGoChannel(gochannel.Config{}, logger)
	publisher := opencensus.PublisherDecorator(pubsub, logger)

	msgCh, err := pubsub.Subscribe(context.Background(), "topic")
	require.NoError(t, err)

	msg := message.NewMessage(watermill.NewULID(), nil)

	// create a span and set the message's context to the span's one
	ctx, _ := trace.StartSpan(context.Background(), "span_name")
	msg.SetContext(ctx)

	err = publisher.Publish("topic", msg)
	require.NoError(t, err)

	// wait for the message with propagated span context
	receivedMsg := <-msgCh
	defer receivedMsg.Ack()
	sc, ok := opencensus.GetSpanContext(receivedMsg)
	require.True(t, ok)
	require.NotNil(t, sc)
}

func TestPublisherDecorator_no_span_context_in_msg(t *testing.T) {
	logger := watermill.NewStdLogger(true, false)
	pubsub := gochannel.NewGoChannel(gochannel.Config{}, logger)
	publisher := opencensus.PublisherDecorator(pubsub, logger)

	msgCh, err := pubsub.Subscribe(context.Background(), "topic")
	require.NoError(t, err)

	msg := message.NewMessage(watermill.NewULID(), nil)
	err = publisher.Publish("topic", msg)
	require.NoError(t, err)

	receivedMsg := <-msgCh
	defer receivedMsg.Ack()
	_, ok := opencensus.GetSpanContext(receivedMsg)
	require.False(t, ok)
}
