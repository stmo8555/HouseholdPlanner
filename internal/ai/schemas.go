package ai

type ExtractedIngredient struct {
	Product string `json:"name" jsonschema_description:"Product name only. No brand, amount, unit or counting word, store, or preparation words."`
	Amount  string `json:"amount" jsonschema_description:"The written amount including its unit or counting word, e.g. \"3 blad\", \"3 klyftor\", \"0,5 kruka\", \"1 msk\", \"1,25 kg\". Never split the number from its unit. Empty string if no amount is written. Never guess."`
	Note    string `json:"note" jsonschema_description:"Everything else written about the item, such as brand, store, or a preparation or quality remark. Copy the words as written. Empty string if there is nothing. Never guess."`
}

type ExtractedIngredientList struct {
	List []ExtractedIngredient `json:"extractedIngredientList" jsonschema_description:"Extracted ingredient items."`
}
