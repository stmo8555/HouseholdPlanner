package ingredient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/stmo8555/HouseholdPlanner/internal/ai"
	"github.com/stmo8555/HouseholdPlanner/internal/product"
	"golang.org/x/net/html"
)

type Extractor struct {
	AIService *ai.Service
}

func CreateExtractor(ai *ai.Service) *Extractor {
	if ai == nil {
		panic(errors.New("Extractor not initialized"))
	}

	return &Extractor{AIService: ai}
}

func (e *Extractor) FromText(ctx context.Context, text string) ([]Ingredient, error) {
	aiIngredients, err := e.AIService.ExtractIngredients(ctx, text)

	if err != nil {
		return nil, err
	}

	ingredients := make([]Ingredient, 0, len(aiIngredients.List))

	for _, v := range aiIngredients.List {
		ingredients = append(ingredients, Ingredient{
			Product: product.Product{
				Name:  v.Product,
				Brand: v.Brand,
			},
			Amount: v.Amount,
		})
	}

	return ingredients, nil
}

func (e *Extractor) FromRecipeURL(ctx context.Context, url string) ([]Ingredient, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	tags := findTags(resp.Body)

	aiIngredients, err := e.AIService.ExtractIngredients(ctx, tags)
	ingredients := make([]Ingredient, 0, len(aiIngredients.List))

	for _, v := range aiIngredients.List {
		ingredients = append(ingredients, Ingredient{
			Product: product.Product{
				Name:  v.Product,
				Brand: v.Brand,
			},
			Amount: v.Amount,
		})
	}

	return ingredients, nil
}

func findTags(r io.Reader) string {
	doc, err := html.Parse(r)

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

	return text.String()
}
