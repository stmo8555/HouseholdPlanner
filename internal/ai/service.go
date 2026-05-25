package ai

import (
	"context"
	"fmt"
	"time"
)

type Service struct {
	client *Client
}

func NewService(client *Client) *Service {
	if client == nil {
		panic("nil ai client")
	}

	return &Service{
		client: client,
	}
}

func (s *Service) ExtractIngredients(ctx context.Context, text string) (ExtractedIngredientList, error) {
	prompt := `Extract groceries from text.
				Rules:
				- Do not guess or infer.
				- Missing amount, brand, or store = "".
				- Product is the grocery name only.
				- The text is most of the time given in swedish. Don't translate it to english.

				Text:
				` + text

	return SendStructuredRequest[ExtractedIngredientList](s.client, ctx, prompt, "ExtractedIngredientList")
}

func (s *Service) ExtractTodo(ctx context.Context, text string) (ExtractedTodoList, error) {
	promptTmpl := `Extract Todos from text.
				Rules:
				- Do not guess or infer.
				- Look at json schema to see extra rules.
				- The text is most of the time given in swedish. Don't translate it to english.
				Information:
				- Todays date: %v
				- Week day: %v

				Text:
				` + text

	today := time.Now()
	prompt := fmt.Sprintf(promptTmpl, today, today.Weekday().String())

	return SendStructuredRequest[ExtractedTodoList](s.client, ctx, prompt, "ExtractedTodoList")
}
