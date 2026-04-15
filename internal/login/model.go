package login

type Session struct {
	UserID      int
	HouseholdID *int
}

type User struct {
	ID int
	Uname string
	Hash string
}
