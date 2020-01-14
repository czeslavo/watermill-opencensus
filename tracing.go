// OpenCensus tracing for Watermill
package opencensus

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

const spanContextKey = "opencensus_span_context"

/*
TracingMiddleware is a Watermill middleware providing OpenCensus tracing.

It creates a span with a name derived from the handler. The span is ended after message is handled.
It tries to extract an existing span context from the incoming message and use it as a parent -
if there's no such it creates a new one.

The span context is set as the message's context so message handlers' code can start children spans out of it.

	spanCtx := trace.FromContext(message.Context())

Depending on a result of the message handling, the span's status is set.

All messages procuded by the handler have span context set so it gets propagated further.
Span context is serialized to binary format and is transported in messages' metadata.
*/
func TracingMiddleware(h message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) (producedMessages []*message.Message, err error) {
		var span *trace.Span
		ctx := msg.Context()

		parentSpanContext, ok := GetSpanContext(msg)
		if ok {
			ctx, span = trace.StartSpanWithRemoteParent(ctx, message.HandlerNameFromCtx(ctx), parentSpanContext)

			span.AddLink(trace.Link{
				TraceID:    parentSpanContext.TraceID,
				SpanID:     parentSpanContext.SpanID,
				Type:       trace.LinkTypeParent,
				Attributes: nil,
			})
		} else {
			ctx, span = trace.StartSpan(ctx, message.HandlerNameFromCtx(ctx))
		}

		defer func() {
			for _, producedMessage := range producedMessages {
				SetSpanContext(span.SpanContext(), producedMessage)
			}
		}()

		defer func() {
			if err == nil {
				span.SetStatus(trace.Status{
					Code:    trace.StatusCodeOK,
					Message: "OK",
				})
			} else {
				span.SetStatus(trace.Status{
					Code:    trace.StatusCodeUnknown,
					Message: err.Error(),
				})
				// some exporters don't handle status (i.e. Stackdriver) therefore we report error attribute too
				span.AddAttributes(trace.StringAttribute("error", err.Error()))
			}
			span.End()
		}()

		msg.SetContext(ctx)
		return h(msg)
	}
}

// SetSpanContext serialize trace.SpanContext to binary format and sets it in a message's metadata.
func SetSpanContext(sc trace.SpanContext, msg *message.Message) {
	binarySc := string(propagation.Binary(sc))
	msg.Metadata.Set(spanContextKey, binarySc)
}

// GetSpanContext gets and deserialize trace.SpanContext from a message's metadata.
func GetSpanContext(message *message.Message) (sc trace.SpanContext, ok bool) {
	binarySc := []byte(message.Metadata.Get(spanContextKey))
	return propagation.FromBinary(binarySc)
}

/*
NewPublisherDecorator creates `message.Publisher` decorator which propagates span context in the published message's metadata.

Note: Please keep in mind that the span context needs to be passed as message's context, otherwise no action will be taken.
*/
func NewPublisherDecorator(logger watermill.LoggerAdapter) func(pub message.Publisher) message.Publisher {
	return func(pub message.Publisher) message.Publisher {
		return &publisherDecorator{pub, logger}
	}
}

type publisherDecorator struct {
	message.Publisher

	logger watermill.LoggerAdapter
}

func (d *publisherDecorator) Publish(topic string, messages ...*message.Message) error {
	for i := range messages {
		msg := messages[i]
		span := trace.FromContext(msg.Context())
		if span == nil {
			d.logger.Debug("Span context nil, cannot propagate", watermill.LogFields{"topic": topic})
		} else {
			SetSpanContext(span.SpanContext(), msg)
		}
	}

	return d.Publisher.Publish(topic, messages...)
}
