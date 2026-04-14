package recipes

type Recipe struct {
	Id           int
	Title        string `json:"title"`
	Img_url      string `json:"img_url"`
	Link         string `json:"link"`
	Household_id int
}
