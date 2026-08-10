package ingredient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	neturl "net/url"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/stmo8555/HouseholdPlanner/internal/ai"
	"github.com/stmo8555/HouseholdPlanner/internal/product"
	"golang.org/x/net/html"
)

type Extractor struct {
	AIService *ai.Service
}

func NewExtractor(ai *ai.Service) *Extractor {
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

// safeHTTPClient blocks Server-Side Request Forgery: the dialer's Control hook
// runs for every connection — including redirects — after DNS resolution but
// before the socket connects, so it rejects private, loopback, and link-local
// targets even if a public hostname rebinds or redirects to an internal IP.
var safeHTTPClient = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 10 * time.Second,
			Control: func(network, address string, _ syscall.RawConn) error {
				host, _, err := net.SplitHostPort(address)
				if err != nil {
					return err
				}
				ip := net.ParseIP(host)
				if ip == nil {
					return fmt.Errorf("ssrf: could not resolve %q to an IP", address)
				}
				if isBlockedIP(ip) {
					return fmt.Errorf("ssrf: refusing to connect to internal address %s", ip)
				}
				return nil
			},
		}).DialContext,
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("ssrf: too many redirects")
		}
		if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
			return fmt.Errorf("ssrf: refusing redirect to scheme %q", req.URL.Scheme)
		}
		return nil
	},
}

func isBlockedIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() ||
		ip.IsUnspecified()
}

func safeGet(ctx context.Context, rawURL string) (*http.Response, error) {
	parsed, err := neturl.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("unsupported url scheme %q (only http/https allowed)", parsed.Scheme)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, err
	}
	return safeHTTPClient.Do(req)
}

func (e *Extractor) FromRecipeURL(ctx context.Context, url string) ([]Ingredient, error) {
	resp, err := safeGet(ctx, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ingredientText := extractIngredientText(resp.Body)

	aiIngredients, err := e.AIService.ExtractIngredients(ctx, ingredientText)
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

func extractIngredientText(r io.Reader) string {
	doc, err := html.Parse(r)

	if err != nil {
		panic(err)
	}

	data := jsonLD(doc)

	return data
	if data != "" {
		return data
	}

	ingredientNode := findIngredientNode(doc)
	if ingredientNode == nil {
		return ""
	}

	return extractNodeText(ingredientNode)
}

var (
	specialCharacters = regexp.MustCompile(`[^\w\såäöÅÄÖé\.,]+`)
	ingredient        = regexp.MustCompile(`^[iI]ngredien(ser|ts)$`)
	units             = regexp.MustCompile(`^([mcd]?l|tm[sk]|k?g|krm)$`)
	numbers           = regexp.MustCompile(`^\d+$`)
	containIngredient = regexp.MustCompile(`[iI]ngredien(ser|ts)`)
)

func jsonLD(doc *html.Node) string {

	var node *html.Node = nil

	var processNode func(n *html.Node)
	processNode = func(n *html.Node) {
		if n.Data == "script" {
			for _, attr := range n.Attr {
				if attr.Key == "type" && attr.Val == "application/ld+json" {
					node = n
					fmt.Println("FOUND!!!!!!!!!")
					return
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			processNode(c)
		}
	}
	processNode(doc)

	if node == nil {
		return ""
	}

	pattern := regexp.MustCompile(`"recipeIngredient"\s*:\s*(\[[^]]*\])`)

	match := pattern.FindStringSubmatch(node.FirstChild.Data)

	if len(match) < 2 {
		panic(fmt.Errorf("recipeIngredient not found"))
	}

	rawIngredients := match[1]
	var ingredients []string
	if err := json.Unmarshal([]byte(rawIngredients), &ingredients); err != nil {
		return ""
	}
	return strings.Join(ingredients, "\n")
}

func printNodeTree(node *html.Node) {
	var output strings.Builder

	var walk func(n *html.Node, depth int)
	walk = func(n *html.Node, depth int) {
		if n == nil {
			return
		}

		indent := strings.Repeat("  ", depth)
		switch n.Type {
		case html.DocumentNode:
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				walk(child, depth)
			}
			return
		case html.ElementNode:
			output.WriteString(indent)
			output.WriteByte('<')
			output.WriteString(n.Data)
			for _, attr := range n.Attr {
				fmt.Fprintf(&output, " %s=%q", attr.Key, attr.Val)
			}
			output.WriteString(">\n")
		case html.TextNode:
			text := strings.TrimSpace(n.Data)
			if text != "" {
				fmt.Fprintf(&output, "%s%q\n", indent, text)
			}
			return
		case html.CommentNode:
			fmt.Fprintf(&output, "%s<!-- %s -->\n", indent, strings.TrimSpace(n.Data))
			return
		case html.DoctypeNode:
			fmt.Fprintf(&output, "%s<!doctype %s>\n", indent, n.Data)
			return
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child, depth+1)
		}
		if n.Type == html.ElementNode {
			fmt.Fprintf(&output, "%s</%s>\n", indent, n.Data)
		}
	}

	walk(node, 0)
	fmt.Print(output.String())
}

func findIngredientNode(doc *html.Node) *html.Node {
	potentialCandidates := make([]*html.Node, 0)

	var processNode func(n *html.Node, pattern *regexp.Regexp)
	processNode = func(n *html.Node, pattern *regexp.Regexp) {
		if n.Type == html.TextNode {
			if pattern.MatchString(strings.TrimSpace(n.Data)) {
				daddyTag := regexp.MustCompile("div|script")
				tag := n.Parent
				for tag != nil && daddyTag.MatchString(tag.Data) {
					fmt.Println("loooping....")
					tag = tag.Parent
				}
				if tag == nil {
					return
				}
				i := 0
				for ; tag != nil; tag = tag.NextSibling {
					fmt.Println("-----" + strconv.Itoa(i) + "-----")
					printNodeTree(tag)
					potentialCandidates = append(potentialCandidates, tag)
					fmt.Println("-----------")
					i++
				}

				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			processNode(c, pattern)
		}
	}
	ingredientHeader := regexp.MustCompile(`^Ingredien(ser|ts)$`)
	ingredientLoose := regexp.MustCompile(`[iI]ngredien(ser|ts)`)
	processNode(doc, ingredientHeader)

	fmt.Printf("numb of potential: %v\n", len(potentialCandidates))
	if len(potentialCandidates) == 0 {
		processNode(doc, ingredientLoose)
		if len(potentialCandidates) == 0 {
			return nil
		}
	}

	score_li := 1
	score_tr := 1
	score_attr := 1
	score_ingredient := 1
	score_regularText := 2
	score_units := 5
	score_numbers := 1

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
	return potentialCandidates[maxIndex]
}

func extractNodeText(node *html.Node) string {
	var text strings.Builder

	var findText func(n *html.Node)
	findText = func(n *html.Node) {
		if n.Type == html.TextNode {
			line := strings.TrimSpace(n.Data)
			if line != "" {
				text.WriteString(line)
				text.WriteByte(' ')
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findText(c)
		}
	}
	findText(node)

	fmt.Printf("Bytes: %v\n", text.Len())
	return text.String()
}
