package user

import "time"

// User represents a user in the application's database.
// The ID corresponds to the ID from the authentication provider.
type User struct {
	ID        string    `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
