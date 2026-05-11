package ai

import (
	"context"
)

type Service struct {
	client *Client
}

func CreateService(client *Client) *Service {
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
