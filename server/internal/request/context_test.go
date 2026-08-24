package request

import (
	"context"
	"testing"
)

func TestRequestIDContextRoundTrip(t *testing.T) {
	original := context.Background()
	const requestID = "req-123"

	withRequestID := WithRequestID(original, requestID)

	if got := GetRequestID(withRequestID); got != requestID {
		t.Errorf("GetRequestID() = %q, want %q", got, requestID)
	}
	if got := GetRequestID(original); got != "" {
		t.Errorf("original context request ID = %q, want empty string", got)
	}
}

func TestGetRequestIDForUnsetAndNilContexts(t *testing.T) {
	t.Run("unset context returns an empty string", func(t *testing.T) {
		if got := GetRequestID(context.Background()); got != "" {
			t.Errorf("GetRequestID() = %q, want empty string", got)
		}
	})

	t.Run("nil context panics", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("GetRequestID(nil) did not panic")
			}
		}()

		GetRequestID(nil)
	})
}
