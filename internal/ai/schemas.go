package ai

type ExtractedIngredient struct {
	Product string `json:"name" jsonschema_description:"product name only. No brand, amount, store, or preparation words."`
	Amount  string `json:"amount" jsonschema_description:"Explicit amount only. Empty string if not written. Never guess."`
	Brand   string `json:"brand" jsonschema_description:"Explicit brand only. Empty string if not written. Never guess."`
}

type ExtractedIngredientList struct {
	List []ExtractedIngredient `json:"extractedIngredientList" jsonschema_description:"Extracted ingredient items."`
}

type ExtractedTodo struct {
	Task      string `json:"task" jsonschema_description:"task only. No Date, repeats or frequency."`
	Due       string `json:"due" jsonschema_description:"Only the due date. Not required"`
	Repeat    string `json:"repeat" jsonschema_description:"Should be never if no repeat. Otherwise daily, weekly, monthly or yearly. Only valid if due date is given"`
	Frequency int    `json:"frequency" jsonschema_description:"How often the task should be done. Example: Every 2 weeks. Here is '2' the frequency"`
}

type ExtractedTodoList struct {
	List []ExtractedTodo `json:"extractedTodoList" jsonschema_description:"Extracted todos."`
}
