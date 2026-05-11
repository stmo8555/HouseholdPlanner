package ai

type ExtractedIngredient struct {
	Product string `json:"name" jsonschema_description:"product name only. No brand, amount, store, or preparation words."`
	Amount  string `json:"amount" jsonschema_description:"Explicit amount only. Empty string if not written. Never guess."`
	Brand   string `json:"brand" jsonschema_description:"Explicit brand only. Empty string if not written. Never guess."`
	Store   string `json:"store" jsonschema_description:"Explicit store only. Empty string if not written. Never guess."`
}

type ExtractedIngredientList struct {
	List []ExtractedIngredient `json:"extractedIngredientList" jsonschema_description:"Extracted ingredient items."`
}
