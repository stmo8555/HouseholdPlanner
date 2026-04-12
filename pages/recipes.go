package pages

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/net/html"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

type Recipe struct {
	Id           int
	Title        string `json:"title"`
	Img_url      string `json:"img_url"`
	Link         string `json:"link"`
	Household_id int
}

func RecipesHandleFunc(c *gin.Context, conn *pgx.Conn) {
	hid, ok := c.Get("household_id")

	if !ok {
		panic("failed to get household id from context")
	}

	sql := `
        SELECT id, title, img_url, link, household_id
        FROM recipes
        WHERE household_id = $1;
    `

	rows, err := conn.Query(context.Background(), sql, hid)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var recipes []Recipe

	for rows.Next() {
		var r Recipe
		err := rows.Scan(
			&r.Id,
			&r.Title,
			&r.Img_url,
			&r.Link,
			&r.Household_id,
		)

		if err != nil {
			panic(err)
		}

		recipes = append(recipes, r)
	}

	if rows.Err() != nil {
		panic(rows.Err())
	}

	data := gin.H{
		"Title":       "Groceries",
		"CurrentPath": c.Request.URL.Path,
		"Data":        recipes,
	}
	c.HTML(http.StatusOK, "recipes.html", data)
}

func RecipesAddHandleFunc(c *gin.Context, conn *pgx.Conn) {
	link := c.PostForm("link")
	hid, ok := c.Get("household_id")
	if !ok {
		panic("failed to get household_id")
	}
	resp, err := http.Get(link)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var doc *html.Node
	doc, err = html.Parse(resp.Body)

	if err != nil {
		panic(err)
	}

	var recipe Recipe

	recipe.Link = link
	recipe.Household_id = hid.(int)

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
		panic("Title not found")
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
		panic("Did not find match for recipes images url")
	}

	sort.Slice(img_tags, func(i, j int) bool {
		return img_tags[i].points < img_tags[j].points
	})

	recipe.Img_url = img_tags[len(img_tags)-1].src

	sql := `INSERT INTO recipes 
	(title, img_url, link, household_id)
	VALUES ($1, $2, $3, $4);`

	_, err = conn.Exec(context.Background(), sql, recipe.Title, recipe.Img_url, recipe.Link, recipe.Household_id)

	if err != nil {
		panic(err)
	}

	for _, a := range img_tags {
		fmt.Printf("%v <-> %v\n", a.points, a.src)
	}

	c.Redirect(302, "/recipes")
}
