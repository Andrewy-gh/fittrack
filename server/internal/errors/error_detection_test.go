package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// TestErrorDetectionWithVariousTypes verifies that error detection works
// correctly with all types of errors used in the application.
func TestErrorDetectionWithVariousTypes(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		checkUnauth    bool
		checkNotFound  bool
		checkPgError   bool
		checkSentinel  bool
		checkValidator bool
		wantMatch      bool
	}{
		{
			name:        "Direct Unauthorized error",
			err:         &Unauthorized{Resource: "exercise", UserID: "123"},
			checkUnauth: true,
			wantMatch:   true,
		},
		{
			name:        "Wrapped Unauthorized error",
			err:         fmt.Errorf("service layer: %w", &Unauthorized{Resource: "workout", UserID: "456"}),
			checkUnauth: true,
			wantMatch:   true,
		},
		{
			name:          "Direct NotFound error",
			err:           &NotFound{Resource: "exercise", ID: "789"},
			checkNotFound: true,
			wantMatch:     true,
		},
		{
			name:          "Wrapped NotFound error",
			err:           fmt.Errorf("repository layer: %w", &NotFound{Resource: "workout", ID: "101"}),
			checkNotFound: true,
			wantMatch:     true,
		},
		{
			name:         "Direct pgconn.PgError",
			err:          &pgconn.PgError{Code: "23505", Message: "unique violation"},
			checkPgError: true,
			wantMatch:    true,
		},
		{
			name:         "Wrapped pgconn.PgError",
			err:          fmt.Errorf("database error: %w", &pgconn.PgError{Code: "23503", Message: "foreign key violation"}),
			checkPgError: true,
			wantMatch:    true,
		},
		{
			name:          "Direct sentinel error (pgx.ErrNoRows)",
			err:           pgx.ErrNoRows,
			checkSentinel: true,
			wantMatch:     true,
		},
		{
			name:          "Wrapped sentinel error",
			err:           fmt.Errorf("query failed: %w", pgx.ErrNoRows),
			checkSentinel: true,
			wantMatch:     true,
		},
		{
			name:        "Unauthorized should not match NotFound",
			err:         &Unauthorized{Resource: "exercise", UserID: "123"},
			checkNotFound: true,
			wantMatch:   false,
		},
		{
			name:        "NotFound should not match Unauthorized",
			err:         &NotFound{Resource: "exercise", ID: "123"},
			checkUnauth: true,
			wantMatch:   false,
		},
		{
			name:          "Generic error should not match custom types",
			err:           errors.New("generic error"),
			checkUnauth:   true,
			checkNotFound: true,
			checkPgError:  true,
			checkSentinel: true,
			wantMatch:     false,
		},
		{
			name:          "Nil error should not match",
			err:           nil,
			checkUnauth:   true,
			checkNotFound: true,
			checkPgError:  true,
			checkSentinel: true,
			wantMatch:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var matched bool

			if tt.checkUnauth {
				var unauthErr *Unauthorized
				matched = errors.As(tt.err, &unauthErr)
			}

			if tt.checkNotFound {
				var notFoundErr *NotFound
				matched = errors.As(tt.err, &notFoundErr)
			}

			if tt.checkPgError {
				var pgErr *pgconn.PgError
				matched = errors.As(tt.err, &pgErr)
			}

			if tt.checkSentinel {
				matched = errors.Is(tt.err, pgx.ErrNoRows)
			}

			if tt.checkValidator {
				var validationErr validator.ValidationErrors
				matched = errors.As(tt.err, &validationErr)
			}

			if matched != tt.wantMatch {
				t.Errorf("Error detection = %v, want %v", matched, tt.wantMatch)
			}
		})
	}
}

// TestMultiLayerWrappedErrors tests that error detection works through
// multiple layers of wrapping.
func TestMultiLayerWrappedErrors(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantType  string
		wantMatch bool
	}{
		{
			name: "Triple wrapped Unauthorized",
			err: fmt.Errorf("handler: %w",
				fmt.Errorf("service: %w",
					fmt.Errorf("repository: %w",
						&Unauthorized{Resource: "exercise", UserID: "123"}))),
			wantType:  "Unauthorized",
			wantMatch: true,
		},
		{
			name: "Triple wrapped NotFound",
			err: fmt.Errorf("handler: %w",
				fmt.Errorf("service: %w",
					fmt.Errorf("repository: %w",
						&NotFound{Resource: "workout", ID: "456"}))),
			wantType:  "NotFound",
			wantMatch: true,
		},
		{
			name: "Triple wrapped pgconn.PgError",
			err: fmt.Errorf("handler: %w",
				fmt.Errorf("service: %w",
					fmt.Errorf("repository: %w",
						&pgconn.PgError{Code: "23505", Message: "unique violation"}))),
			wantType:  "PgError",
			wantMatch: true,
		},
		{
			name: "Triple wrapped sentinel error",
			err: fmt.Errorf("handler: %w",
				fmt.Errorf("service: %w",
					fmt.Errorf("repository: %w", pgx.ErrNoRows))),
			wantType:  "Sentinel",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var matched bool

			switch tt.wantType {
			case "Unauthorized":
				var unauthErr *Unauthorized
				matched = errors.As(tt.err, &unauthErr)
			case "NotFound":
				var notFoundErr *NotFound
				matched = errors.As(tt.err, &notFoundErr)
			case "PgError":
				var pgErr *pgconn.PgError
				matched = errors.As(tt.err, &pgErr)
			case "Sentinel":
				matched = errors.Is(tt.err, pgx.ErrNoRows)
			}

			if matched != tt.wantMatch {
				t.Errorf("Multi-layer error detection = %v, want %v", matched, tt.wantMatch)
			}
		})
	}
}

// TestErrorTypePreservationThroughWrapping verifies that wrapping errors
// preserves their type information and field values.
func TestErrorTypePreservationThroughWrapping(t *testing.T) {
	t.Run("Unauthorized preserves fields", func(t *testing.T) {
		original := &Unauthorized{Resource: "exercise", UserID: "user123"}
		wrapped := fmt.Errorf("failed operation: %w", original)

		var extracted *Unauthorized
		if !errors.As(wrapped, &extracted) {
			t.Fatal("Failed to extract Unauthorized from wrapped error")
		}

		if extracted.Resource != "exercise" {
			t.Errorf("Resource = %v, want exercise", extracted.Resource)
		}
		if extracted.UserID != "user123" {
			t.Errorf("UserID = %v, want user123", extracted.UserID)
		}
	})

	t.Run("NotFound preserves fields", func(t *testing.T) {
		original := &NotFound{Resource: "workout", ID: "789"}
		wrapped := fmt.Errorf("failed operation: %w", original)

		var extracted *NotFound
		if !errors.As(wrapped, &extracted) {
			t.Fatal("Failed to extract NotFound from wrapped error")
		}

		if extracted.Resource != "workout" {
			t.Errorf("Resource = %v, want workout", extracted.Resource)
		}
		if extracted.ID != "789" {
			t.Errorf("ID = %v, want 789", extracted.ID)
		}
	})

	t.Run("pgconn.PgError preserves fields", func(t *testing.T) {
		original := &pgconn.PgError{Code: "23505", Message: "duplicate key"}
		wrapped := fmt.Errorf("database error: %w", original)

		var extracted *pgconn.PgError
		if !errors.As(wrapped, &extracted) {
			t.Fatal("Failed to extract PgError from wrapped error")
		}

		if extracted.Code != "23505" {
			t.Errorf("Code = %v, want 23505", extracted.Code)
		}
		if extracted.Message != "duplicate key" {
			t.Errorf("Message = %v, want 'duplicate key'", extracted.Message)
		}
	})
}
