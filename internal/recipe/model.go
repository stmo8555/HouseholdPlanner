package recipe

type Recipe struct {
	Id           int
	Title        string `json:"title"`
	ImgURL      string `json:"img_url"`
	Link         string `json:"link"`

	Household_id int
}
