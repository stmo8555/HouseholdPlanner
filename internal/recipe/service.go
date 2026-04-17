package recipe

import (
	"context"
	"errors"
	"golang.org/x/net/html"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

type Service struct {
	Repo *Repo
}

func (s *Service) List(ctx context.Context, hid int) ([]Recipe, error) {
	return s.Repo.List(ctx, hid)
}
func (s *Service) Add(c context.Context, hid int, link string) error {
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

	var recipe Recipe

	recipe.Link = link
	recipe.Household_id = hid

	var findTitle func(n *html.Node)
	findTitle = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "h1" {
				var findtext func(n *html.Node) string
				findtext = func(n *html.Node) string {
					if n.Type == html.TextNode {
						return strings.TrimSpace(n.Data)
					}
					for c := n.FirstChild; c != nil; c = c.NextSibling {
						return findtext(c)
					}

					return ""
				}

				recipe.Title = findtext(n.FirstChild)
				return
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findTitle(c)
		}
	}
	findTitle(doc)

	titleComponents := strings.Split(recipe.Title, " ")
	matchFactor := len(titleComponents)

	if matchFactor < 1 {
		return errors.New("Title not found")
	}

	type item struct {
		src    string
		points int
	}

	img_tags := []item{}

	var findImg func(n *html.Node)
	findImg = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "img" {
				src := ""
				points := 0
				for _, v := range n.Attr {
					switch v.Key {
					case "src":
						src = v.Val
					case "alt":
					default:
						continue
					}

					for _, tc := range titleComponents {
						match, err := regexp.MatchString("(?i)"+tc, v.Val)

						if err != nil {
							panic(err)
						}

						if match {
							points += 1
						}
					}

					match, err := regexp.MatchString("^(?i)https", src)

					if err != nil {
						panic(err)
					}

					if match && points >= matchFactor {
						img_tags = append(img_tags, item{src, points})
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findImg(c)
		}
	}

	findImg(doc)

	if len(img_tags) < 1 {
		return errors.New("Failed to find recipe image")
	}

	sort.Slice(img_tags, func(i, j int) bool {
		return img_tags[i].points < img_tags[j].points
	})

	recipe.Img_url = img_tags[len(img_tags)-1].src

	return s.Repo.Add(c, hid, recipe)
}
