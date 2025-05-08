package mediatr

import (
	"context"
	"testing"
)

func Benchmark_Send(b *testing.B) {
	handler := &RequestTestHandler{}
	ctx := context.Background()
	request := &RequestTest{Data: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear and re-register for each iteration to get consistent results
		ClearRequestRegistrations()
		errRegister := RegisterRequestHandler[*RequestTest, *ResponseTest](handler)
		if errRegister != nil {
			b.Fatal(errRegister)
		}

		_, err := Send[*RequestTest, *ResponseTest](ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Publish(b *testing.B) {
	handler1 := &NotificationTestHandler{}
	handler2 := &NotificationTestHandler4{}
	notification := &NotificationTest{Data: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear and re-register for each iteration
		ClearNotificationRegistrations()
		errRegister := RegisterNotificationHandlers[*NotificationTest](handler1, handler2)
		if errRegister != nil {
			b.Fatal(errRegister)
		}

		err := Publish[*NotificationTest](context.Background(), notification)
		if err != nil {
			b.Fatal(err)
		}
	}
}
