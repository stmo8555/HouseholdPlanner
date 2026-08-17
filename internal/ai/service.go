package ai

import (
	"context"
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
				- Missing amount or note = "".
				- Product is the grocery name only.
				- Note holds anything else written about the item (brand, store, quality or preparation remarks), copied as written.
				- The text is most of the time given in swedish. Don't translate it to english.

				Text:
				` + text

	return SendStructuredRequest[ExtractedIngredientList](s.client, ctx, prompt, "ExtractedIngredientList")
}
