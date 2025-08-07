// Package testutils provides utilities for testing
package testutils

import (
	"context"
)

// Example_testutils demonstrates how to use the test utility
func Example_testutils() {
	// This is just an example of how to use the utility
	// In a real test, you would have a database connection
	/*
		ctx := context.Background()

		// Set up the test user context
		ctx = SetTestUserContext(ctx, t, db, "test-user-id")

		// Now you can run your tests with RLS enabled
		// The database will only return data for "test-user-id"
	*/
	_ = context.Background()
}
