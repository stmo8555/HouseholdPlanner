package household

import "time"

type Household struct {
	ID   int
	Name string
	Code string
}

type Member struct {
	ID      int
	Name    string
	IsOwner bool
	IsYou   bool
}

type Invite struct {
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time

	Link         string
	Masked       string
	Created      string
	Expires      string
	ExpiringSoon bool
}

type SettingsView struct {
	Household Household
	Members   []Member
	Invites   []Invite
	IsOwner   bool
}


