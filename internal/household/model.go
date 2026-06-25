package household

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

type SettingsView struct {
	Household Household
	Members   []Member
	IsOwner   bool
}


