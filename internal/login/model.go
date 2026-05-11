package login

import (
	"time"
)

type Session struct {
	ID          string
	UserID      int
	HouseholdID *int
	ExpiresAt   time.Time
}

type User struct {
	ID    int
	Uname string
	Hash  string
}
