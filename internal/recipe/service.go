package recipe

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type Service struct {
	repo                IRepo
	productService      IProductService
	ingredientExtractor IIngredientExtractor
}

func NewService(repo IRepo, product IProductService, ingredient IIngredientExtractor) *Service {
	if repo == nil || product == nil || ingredient == nil {
		panic("service not initialized")
	}

	return &Service{
		repo:                repo,
		productService:      product,
		ingredientExtractor: ingredient,
	}
}

func (s *Service) List(ctx context.Context, hid int) ([]Recipe, map[int][]RecipeIngredient, error) {
	return s.repo.List(ctx, hid)
}

func (s *Service) Add(ctx context.Context, hid int, link string) error {
	if !strings.HasPrefix(link, "http") {
		return errors.New("Recipe: Not an URL")
	}

	resp, err := http.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var doc *html.Node
	doc, err = html.Parse(resp.Body)

	if err != nil {
		return err
	}

	title := findTitle(doc)

	recipe := Recipe{
		Link:        link,
		HouseholdID: hid,
		Title:       title,
		ImgURL:      findImg(doc, title),
	}

	var recipeID int
	recipeID, err = s.repo.AddRecipe(ctx, hid, recipe)

	if err != nil {
		panic(err)
	}

	ingredients, err := s.ingredientExtractor.FromRecipeURL(ctx, recipe.Link)

	if err != nil {
		panic(err)
	}

	recipeIngredients := make([]RecipeIngredient, 0, len(ingredients))

	for _, i := range ingredients {
		id, err := s.productService.GetID(ctx, i.Product)

		if err != nil {
			panic(err)
		}

		i.ProductID = id

		ri := RecipeIngredient{RecipeID: recipeID, Ingredient: i}
		recipeIngredients = append(recipeIngredients, ri)
	}

	return s.repo.AddIngredients(ctx, recipeIngredients)
}

func (s *Service) Ingredients(ctx context.Context, recipeId, hid int) ([]RecipeIngredient, error) {
	return s.repo.Ingredients(ctx, recipeId, hid)
}

func findTitle(n *html.Node) string {
	if n == nil {
		return ""
	}

	if h1 := findFirst(n, "h1"); h1 != nil {
		return textContent(h1)
	}

	if title := findFirst(n, "title"); title != nil {
		return textContent(title)
	}

	return ""
}

func textContent(n *html.Node) string {
	if n == nil {
		return ""
	}

	if n.Type == html.TextNode {
		return strings.TrimSpace(n.Data)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text := textContent(c)
		if text != "" {
			return text
		}
	}

	return ""
}

func findFirst(n *html.Node, tag string) *html.Node {
	if n == nil {
		return nil
	}

	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findFirst(c, tag); found != nil {
			return found
		}
	}

	return nil
}

func findImg(doc *html.Node, title string) string {
	src := ""

	var findImg func(n *html.Node)

	findImg = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, v := range n.Attr {
				if v.Key == "alt" && strings.Contains(v.Val, title) {
					if n.Parent.Type == html.ElementNode && n.Parent.Data == "picture" {
						p := n.Parent
						for c := p.FirstChild; c != nil; c = c.NextSibling {
							if c.Type == html.ElementNode && c.Data == "source" {
								for _, v := range c.Attr {
									if v.Key == "srcset" {
										src = findLargestImageFromSrcSet(v.Val)
										return
									}
								}
							}
						}
					}

					for _, v := range n.Attr {
						if v.Key == "src" {
							src = v.Val
							return
						}
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findImg(c)
		}
	}

	findImg(doc)

	if src == "" {
		src = FindLooseImage(doc, title)
	}

	return src
}

var widthRegex = regexp.MustCompile(`\s(\d+)w`)

func findLargestImageFromSrcSet(srcsetStr string) string {

	srcset := strings.Split(srcsetStr, ",")
	for i, v := range srcset {
		srcset[i] = strings.TrimSpace(v)
		if !strings.HasPrefix(srcset[i], "http") {
			return srcsetStr
		}
	}

	maxWidth := 0
	src := ""

	for _, v := range srcset {
		matches := widthRegex.FindStringSubmatch(v)
		if matches != nil {
			width, err := strconv.Atoi(matches[1])
			if err != nil {
				panic(err)
			}

			if width > maxWidth {
				maxWidth = width
				src = strings.Fields(v)[0]
			}
		}
	}

	return src
}

func FindLooseImage(doc *html.Node, title string) string {
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return ""
	}

	rendered := buf.String()

	urlRe := regexp.MustCompile(`https?://[^"'<> ]+\.(jpg|jpeg|webp|png)(\?[^"'<> ]*)?`)
	urls := urlRe.FindAllString(rendered, -1)

	titleWords := strings.Fields(normalizeForSearch(title))

	for _, raw := range urls {
		cleanURL := html.UnescapeString(raw)

		searchURL := normalizeForSearch(cleanURL)

		ok := true
		for _, word := range titleWords {
			if !strings.Contains(searchURL, word) {
				ok = false
				break
			}
		}

		if ok {
			return cleanURL
		}
	}

	return ""
}

func normalizeForSearch(s string) string {
	s = strings.ToLower(s)

	replacer := strings.NewReplacer(
		"å", "a",
		"ä", "a",
		"ö", "o",
		"é", "e",
		"è", "e",
		"ü", "u",
	)

	return replacer.Replace(s)
}

func printNode(n *html.Node) {
	var buf bytes.Buffer
	html.Render(&buf, n)
	fmt.Println(buf.String())
}

