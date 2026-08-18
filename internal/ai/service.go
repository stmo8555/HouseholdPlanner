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
				- Only extract things that are actually food, drink, or
				  household shopping items. Leave out anything else entirely:
				  names of people, greetings, dates, headings, notes to self,
				  serving suggestions ("Serveringsförslag"), and any word that
				  is not a real product -- including profanity, insults, crude
				  or joke words, and nonsense typed to test the feature.
				- If nothing in the text is a grocery, return an empty list.
				- Missing amount or note = "".
				- Product is the grocery name only.
				- Note holds anything else written about the item (brand, store, quality or preparation remarks), copied as written.
				- The text is most of the time given in swedish. Don't translate it to english.
				- Swedish counting words (blad, klyftor, kruka, stjälk, knippe,
				  burk, paket, förp) belong to the amount, never to the product
				  name. Keep Swedish spelling exactly as written. Examples:
				    "3 blad salladskål"          -> name "salladskål", amount "3 blad", note ""
				    "3 klyftor vitlök, hackad"   -> name "vitlök", amount "3 klyftor", note "hackad"
				    "0,5 kruka färsk persilja"  -> name "persilja", amount "0,5 kruka", note "färsk"

				Text:
				` + text

	return SendStructuredRequest[ExtractedIngredientList](s.client, ctx, prompt, "ExtractedIngredientList")
}
