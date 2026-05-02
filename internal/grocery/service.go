package grocery

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"golang.org/x/net/html"
	"net/http"
	"regexp"
	"strings"
)

type Service struct {
	Repo           *Repo
	FoodCategories map[string]string
}

func (s *Service) GetTopProducts(ctx context.Context, householdID int) ([]string, error) {
	return s.Repo.getTopProducts(ctx, householdID)
}

var client = openai.NewClient()
var schema = GenerateSchema[GroceriesAI]()

func ai(ctx context.Context, text string) []Grocery {
	prompt := `Extract groceries from text.
				Rules:
				- Do not guess or infer.
				- Missing amount, brand, or store = "".
				- Product is the grocery name only.

				Text:
				` + text

	response, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: openai.ChatModelGPT4_1Nano,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(prompt),
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigParamOfJSONSchema(
				"groceries",
				schema,
			),
		},

		Temperature:     openai.Float(0),
		MaxOutputTokens: openai.Int(300),
	})

	if err != nil {
		panic(err)
	}

	var gs GroceriesAI
	err = json.Unmarshal([]byte(response.OutputText()), &gs)

	if err != nil {
		panic(err)
	}

	groceries := make([]Grocery, len(gs.List))
	for i, v := range gs.List {
		groceries[i] = Grocery{
			Product: v.Product,
			Amount: v.Amount,
			Brand: v.Brand,
			Store: v.Store,
		}
	}

	return groceries
}

func (s *Service) IngredientsFromRecipe(ctx context.Context, url string) []Grocery {

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	tags := findTags(resp)

	return ai(ctx, tags)
}

func (s *Service) AddGroceries(ctx context.Context, groceries []Grocery) error {
	for i, g := range groceries {
		key := strings.ToLower(strings.TrimSpace(g.Product))

		category := "other"

		if cat, ok := s.FoodCategories[key]; ok {
			category = cat
		} else {
			for token := range strings.FieldsSeq(key) {
				if cat, ok := s.FoodCategories[token]; ok {
					category = cat
					break
				}
			}
		}

		groceries[i].Category = category
	}

	return s.Repo.AddGroceries(ctx, groceries)
}

func (s *Service) SmartAdd(ctx context.Context, text string) ([]Grocery, error) {
	return ai(ctx, text), nil
}

func (s *Service) List(ctx context.Context, sortBy, order string, householdID int) (GroceriesView, error) {
	allowedSorts := map[string]string{
		"product": "product",
		"amount":  "amount",
		"brand":   "brand",
		"store":   "store",
	}

	column, ok := allowedSorts[sortBy]
	if !ok {
		column = "product"
	}

	groceries, err := s.Repo.List(ctx, column, order, householdID)

	if err != nil {
		return GroceriesView{}, err
	}

	var sortedGroceries GroceriesView

	for _, g := range groceries {
		if g.Picked {
			sortedGroceries.Picked = append(sortedGroceries.Picked, g)
		} else {
			switch g.Category {
			case "dairy":
				sortedGroceries.Dairy = append(sortedGroceries.Dairy, g)
			case "pantry":
				sortedGroceries.Pantry = append(sortedGroceries.Pantry, g)
			case "fruit & vegetables":
				sortedGroceries.FruitAndVegetables = append(sortedGroceries.FruitAndVegetables, g)
			case "meat and fish":
				sortedGroceries.MeatAndFish = append(sortedGroceries.MeatAndFish, g)
			default:
				sortedGroceries.Other = append(sortedGroceries.Other, g)
			}
		}
	}

	return sortedGroceries, nil
}

func (s *Service) CountUnpicked(ctx context.Context, householdID int) (int, error) {
	return s.Repo.AmountOfUnpickedGroceries(ctx, householdID)
}

func (s *Service) TogglePicked(ctx context.Context, id, householdID int) error {
	return s.Repo.TogglePicked(ctx, id, householdID)
}

func (s *Service) DeletePicked(ctx context.Context, householdId int) error {
	return s.Repo.DeletePicked(ctx, householdId)
}

func findTags(resp *http.Response) string {
	doc, err := html.Parse(resp.Body)

	if err != nil {
		panic(err)
	}

	potentialCandidates := make([]*html.Node, 0)

	var processNode func(n *html.Node, pattern *regexp.Regexp)
	processNode = func(n *html.Node, pattern *regexp.Regexp) {
		if n.Type == html.TextNode {
			if pattern.MatchString(strings.TrimSpace(n.Data)) {
				parent := n.Parent
				for parent != nil && parent.Data == "div" {
					parent = parent.Parent
				}

				parent = parent.Parent
				for ; parent != nil; parent = parent.NextSibling {
					potentialCandidates = append(potentialCandidates, parent)
				}

				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			processNode(c, pattern)
		}
	}
	ingredientHeader := regexp.MustCompile(`^[iI]ngredien(ser|ts)$`)
	ingredientLoose := regexp.MustCompile(`[iI]ngredien(ser|ts)`)
	processNode(doc, ingredientHeader)

	fmt.Printf("numb of potential: %v\n", len(potentialCandidates))
	if len(potentialCandidates) == 0 {
		processNode(doc, ingredientLoose)
		if len(potentialCandidates) == 0 {
			return ""
		}
	}

	score_li := 1
	score_tr := 1
	score_attr := 1
	score_ingredient := 1
	score_regularText := 2
	score_units := 5
	score_numbers := 1

	specialCharacters := regexp.MustCompile(`[^\w\såäöÅÄÖ\.,]+`)
	ingredient := regexp.MustCompile(`^[iI]ngredien(ser|ts)$`)
	units := regexp.MustCompile(`^([mcd]?l|tm[sk]|k?g|krm)$`)
	numbers := regexp.MustCompile(`^\d+$`)
	containIngredient := regexp.MustCompile(`[iI]ngredien(ser|ts)`)

	var bestFit func(n *html.Node, index int)

	points := make([]int, len(potentialCandidates))
	bestFit = func(n *html.Node, index int) {
		if n.Type == html.TextNode {
			data := strings.TrimSpace(n.Data)

			if !specialCharacters.MatchString(data) {
				points[index] += score_regularText
			}

			if ingredient.MatchString(data) {
				points[index] += score_ingredient
			}

			if units.MatchString(data) {
				points[index] += score_units
			}

			if numbers.MatchString(data) {
				points[index] += score_numbers
			}

		}
		if n.Type == html.ElementNode {
			// fmt.Println("_____Element Node_____")
			for _, attr := range n.Attr {
				for val := range strings.SplitSeq(attr.Val, " ") {
					if containIngredient.MatchString(val) {
						points[index] += score_attr
					}
				}
			}
			if n.Data == "li" {
				points[index] += score_li
			}
			if n.Data == "tr" {
				points[index] += score_tr
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			bestFit(c, index)
		}
	}

	for i, nodes := range potentialCandidates {
		bestFit(nodes, i)
	}

	var text strings.Builder

	var findText func(n *html.Node)
	findText = func(n *html.Node) {
		if n.Type == html.TextNode {
			line := specialCharacters.ReplaceAllString(strings.TrimSpace(n.Data), "")
			text.WriteString(line + " ")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findText(c)
		}
	}

	maxIndex := 0
	maxValue := points[0]

	for i, v := range points {
		if v > maxValue {
			maxValue = v
			maxIndex = i
		}
		fmt.Printf("Points: %v %v\n", i, v)
	}

	fmt.Printf("Choosing: %v\n", maxIndex)
	findText(potentialCandidates[maxIndex])

	fmt.Printf("Bytes: %v", text.Len())

	fmt.Println("--------------------------------------")
	var findJSON func(n *html.Node)
	findJSON = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" {
			for _, attr := range n.Attr {
				if attr.Key == "type" && attr.Val == "application/ld+json" {
					fmt.Println(n.FirstChild.Data)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findJSON(c)
		}
	}
	fmt.Println("--------------------------------------")

	return text.String()
}

func (s *Service) Edit(ctx context.Context, groceries []Grocery, householdId int) error {
	for i := range len(groceries) {
		groceries[i].Picked = false
		groceries[i].HouseholdID = householdId
	}

	return s.Repo.Edit(ctx, groceries)
}

// Structured Outputs uses a subset of JSON schema
// These flags are necessary to comply with the subset
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
