package login

import (
	"time"
)

type Session struct {
	ID          string
	User        User
	HouseholdID *int
	ExpiresAt   time.Time
}

type User struct {
	ID       int
	Uname    string
	Hash     string
	LastSeen time.Time
}
