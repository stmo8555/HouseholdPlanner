package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

type Client struct {
	API openai.Client
}

func NewClient() *Client {
	return &Client{API: openai.NewClient()}
}

func SendStructuredRequest[T any](c *Client, ctx context.Context, prompt string, name string) (T, error) {
	var result T

	start := time.Now()
	response, err := c.API.Responses.New(ctx, responses.ResponseNewParams{
		Model: openai.ChatModelGPT5_6Luna,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(prompt),
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigParamOfJSONSchema(
				name,
				GenerateSchema[T](),
			),
		},
		Reasoning: openai.ReasoningParam{
			Effort: openai.ReasoningEffortNone,
		},
		MaxOutputTokens: openai.Int(4000),
	})

	elapsed := time.Since(start)

	if err != nil {
		return result, err
	}

	fmt.Printf("ai: %s took %s (in=%d out=%d reasoning=%d)\n",
		response.Model,
		elapsed.Round(time.Millisecond),
		response.Usage.InputTokens,
		response.Usage.OutputTokens,
		response.Usage.OutputTokensDetails.ReasoningTokens)

	if response.Status == responses.ResponseStatusIncomplete {
		return result, fmt.Errorf("ai response incomplete (%s), no usable output", response.IncompleteDetails.Reason)
	}

	fmt.Println("response from ai: " + response.OutputText())

	if err := json.Unmarshal([]byte(response.OutputText()), &result); err != nil {
		return result, err
	}

	return result, nil
}

func GenerateSchema[T any]() map[string]any {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)

	data, _ := json.Marshal(schema)
	var result map[string]any
	json.Unmarshal(data, &result)
	return result
}
