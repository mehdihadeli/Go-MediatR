package mediatr

import (
	"context"
	"github.com/goccy/go-reflect"
	"testing"
)

func Benchmark_Send(b *testing.B) {
	// because benchmark method will run multiple times, we need to reset the request handler registry before each run.
	requestHandlersRegistrations = make(map[reflect.Type]interface{})

	handler := &RequestTestHandler{}
	errRegister := RegisterRequestHandler[*RequestTest, *ResponseTest](handler)
	if errRegister != nil {
		b.Error(errRegister)
	}

	b.ResetTimer()
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, err := Send[*RequestTest, *ResponseTest](ctx, &RequestTest{Data: "test"})
		if err != nil {
			b.Error(err)
		}
	}
}

func Benchmark_Publish(b *testing.B) {
	// because benchmark method will run multiple times, we need to reset the notification handlers registry before each run.
	notificationHandlersRegistrations = make(map[reflect.Type][]interface{})

	handler := &NotificationTestHandler{}
	handler2 := &NotificationTestHandler4{}

	errRegister := RegisterNotificationHandlers[*NotificationTest](handler, handler2)
	if errRegister != nil {
		b.Error(errRegister)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := Publish[*NotificationTest](context.Background(), &NotificationTest{Data: "test"})
		if err != nil {
			b.Error(err)
		}
	}
}
