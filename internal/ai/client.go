package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

type Client struct {
	API openai.Client
}

func CreateClient() *Client {
	return &Client{API: openai.NewClient()}
}

func SendStructuredRequest[T any](c *Client, ctx context.Context, prompt string, name string) (T, error) {
	var result T

	response, err := c.API.Responses.New(ctx, responses.ResponseNewParams{
		Model: openai.ChatModelGPT4_1Nano,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(prompt),
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigParamOfJSONSchema(
				name,
				GenerateSchema[T](),
			),
		},
		Temperature:     openai.Float(0),
		MaxOutputTokens: openai.Int(1000),
	})

	if err != nil {
		return result, err
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
