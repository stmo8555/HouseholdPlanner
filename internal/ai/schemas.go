package ai

type ExtractedIngredient struct {
	Product string `json:"name" jsonschema_description:"product name only. No brand, amount, store, or preparation words."`
	Amount  string `json:"amount" jsonschema_description:"Explicit amount only. Empty string if not written. Never guess."`
	Note    string `json:"note" jsonschema_description:"Everything else written about the item, such as brand, store, or a preparation or quality remark. Copy the words as written. Empty string if there is nothing. Never guess."`
}

type ExtractedIngredientList struct {
	List []ExtractedIngredient `json:"extractedIngredientList" jsonschema_description:"Extracted ingredient items."`
}
