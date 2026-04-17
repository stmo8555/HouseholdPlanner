package grocery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"golang.org/x/net/html"
)

type Service struct {
	Repo *Repo
}

func (s *Service) GetTopProducts(ctx context.Context, householdID int) ([]string, error) {
	return s.Repo.getTopProducts(ctx, householdID)
}

func ai(ctx context.Context, link string) string {

	resp, err := http.Get(link)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	start := time.Now()

	tags := findTags(resp)

	fmt.Println("took:", time.Since(start))

	// fmt.Println("Parsed: " + tags)

	client := openai.NewClient()

	templatePrompt := `Extract ingredients from the text.

Return ONLY raw JSON (no markdown, no code blocks):
[
	{ "Product": "ingredient name", "Amount": "amount", "Brand": "brand"}
]

Rules:
- Amount = only numbers + units (e.g. "2 dl", "1 msk", "efter smak")
- Product = only the ingredient name
- Brand = only the products brand
- If there is not suffiecent info, return empty json list []
- Remove preparation words (hackad, skivad, riven, etc.)
- Remove text after commas
- No explanations
- Do NOT wrap output in ANYTHING ELSE THAN []

Text: 
%s`

	question := fmt.Sprintf(templatePrompt, tags)
	ai_resp, ai_err := client.Responses.New(ctx, responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(question)},
		Model: openai.ChatModelGPT4_1Nano,
	})

	if ai_err != nil {
		panic(ai_err)
	}

	return ai_resp.OutputText()
}

func (s *Service) IngredientsFromRecipe(ctx context.Context, url string) []Grocery {
	received := ai(ctx, url)
	// 	received := `[
	//   { "Product": "olivolja", "Amount": "till stekning" },
	//   { "Product": "vitlök", "Amount": "3 klyftor" },
	//   { "Product": "smör", "Amount": "till stekning" },
	//   { "Product": "färska salsicciakorvar", "Amount": "4" },
	//   { "Product": "vispgrädde", "Amount": "2 dl" },
	//   { "Product": "passerade tomater", "Amount": "3 dl" },
	//   { "Product": "tomatpuré", "Amount": "2 msk" },
	//   { "Product": "vitt vin", "Amount": "1-2 dl" },
	//   { "Product": "oregano", "Amount": "några kvistar" },
	//   { "Product": "timjankvist", "Amount": "några kvistar" },
	//   { "Product": "pasta (av favoritsort)", "Amount": "4 port" },
	//   { "Product": "salt", "Amount": "efter smak" },
	//   { "Product": "svartpeppar", "Amount": "nymalen" },
	//   { "Product": "persilja", "Amount": "några kvistar" }
	// ]`

	var gs []Grocery
	err := json.Unmarshal([]byte(received), &gs)

	if err != nil {
		panic(err)
	}

	return gs
}

func (s *Service) AddGroceries(ctx context.Context, groceries []Grocery) error {
	return s.Repo.AddGroceries(ctx, groceries)
}

func (s *Service) List(ctx context.Context, householdID int) (Groceries, error) {
	groceries, err := s.Repo.List(ctx, householdID)

	if err != nil {
		return Groceries{}, err
	}

	var sortedGroceries Groceries

	for _, g := range groceries {
		if g.Picked {
			sortedGroceries.Picked = append(sortedGroceries.Picked, g)
		} else {
			sortedGroceries.NotPicked = append(sortedGroceries.NotPicked, g)
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
