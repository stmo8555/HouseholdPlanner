package home

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"github.com/stmo8555/HouseholdPlanner/internal/groceries"
	"github.com/stmo8555/HouseholdPlanner/internal/recipes"
	"github.com/stmo8555/HouseholdPlanner/internal/todos"
)

type Service struct {
}

var (
	templatePrompt = `You are a structured data extraction engine.

Your task is to read a natural language input (primarily Swedish, but sometimes English) and convert it into a JSON array of typed objects.

OUTPUT FORMAT (STRICT):
Return ONLY valid JSON.
No markdown.
No explanations.
Always return a JSON array.

Each object must follow:

{
  "type": "grocery" | "todo" | "recipe",
  "data": { ... }
}

GENERAL RULES:
1. Do NOT invent, infer, or group information.
2. Extract only what is explicitly present in the input.
3. Be conservative: if unsure, choose grocery or todo, NEVER guess recipe.
4. Ingredient lists are NOT recipes.
5. Do NOT create structure from lists.
6. Do NOT deduplicate unless explicitly repeated.

TYPE RULES:

-----------------------
GROCERY
-----------------------
Use for all food items, ingredients, and shopping items.

Fields:
- product: item name (normalize lightly if needed, e.g. "3 korvar" → "korv")
- amount: quantity if present
- brand: optional
- store: optional
- picked: always false unless explicitly true

-----------------------
TODO
-----------------------
Use for actions, intentions, or reminders.

Fields:
- title: short task text (preserve language)

-----------------------
RECIPE (LINK ONLY)
-----------------------
THIS TYPE IS STRICTLY RESTRICTED.

Rules:
- ONLY output a recipe if a URL is explicitly present in the input.
- If no URL exists → DO NOT output recipe at all.
- Never infer or create recipe names from ingredients or text.
- Recipe is ONLY a container for links.

Fields:
- link: must be the exact URL from input (no modification)
- title: "" unless explicitly stated near the link
- img_url: ""

CLASSIFICATION RULES:
- "need to", "I should", tasks → todo
- shopping items → grocery
- ingredient lists → grocery ONLY
- URLs → recipe ONLY if present

LANGUAGE:
Preserve original language (Swedish/English mix allowed).

Convert the following input:
%s`
)

func (s *Service) AI(ctx context.Context, question string) Content {

	client := openai.NewClient()

	prompt := fmt.Sprintf(templatePrompt, question)
	ai_resp, ai_err := client.Responses.New(ctx, responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(prompt)},
		Model: openai.ChatModelGPT4_1Nano,
	})

	if ai_err != nil {
		panic(ai_err)
	}

	var envelopes []Envelope

	json.Unmarshal([]byte(ai_resp.OutputText()), &envelopes)

	gs := make([]groceries.Grocery, 0)
	rs := make([]recipes.Recipe, 0)
	ts := make([]todos.Todo, 0)

	for _, e := range envelopes {
		switch e.Type {
		case "grocery":
			var g groceries.Grocery
			json.Unmarshal(e.Data, &g)
			gs = append(gs, g)
		case "todo":
			var t todos.Todo
			json.Unmarshal(e.Data, &t)
			ts = append(ts, t)
		case "recipe":
			var r recipes.Recipe
			json.Unmarshal(e.Data, &r)
			rs = append(rs, r)
		}
	}

	return Content{Groceries: gs, Todos: ts, Recipes: rs}
}
