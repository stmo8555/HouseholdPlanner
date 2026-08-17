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
				Name: v.Product,
			},
			Amount: v.Amount,
			Note:   v.Note,
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

	ingredientText, err := extractIngredientText(resp.Body)
	if err != nil {
		return nil, err
	}

	aiIngredients, err := e.AIService.ExtractIngredients(ctx, ingredientText)
	if err != nil {
		return nil, err
	}
	ingredients := make([]Ingredient, 0, len(aiIngredients.List))

	for _, v := range aiIngredients.List {
		ingredients = append(ingredients, Ingredient{
			Product: product.Product{
				Name: v.Product,
			},
			Amount: v.Amount,
			Note:   v.Note,
		})
	}

	return ingredients, nil
}

func extractIngredientText(r io.Reader) (string, error) {
	doc, err := html.Parse(r)

	if err != nil {
		return "", err
	}

	return extractJSONLDIngredientText(doc)
}

var pattern = regexp.MustCompile(`"recipeIngredient"\s*:\s*(\[[^]]*\])`)

func extractJSONLDIngredientText(doc *html.Node) (string, error) {
	var ingredients []string
	var found bool
	var matchErr error

	var processNode func(n *html.Node)
	processNode = func(n *html.Node) {
		if n.Data == "script" {
			for _, attr := range n.Attr {
				if attr.Key == "type" && attr.Val == "application/ld+json" {
					if n.FirstChild == nil {
						break
					}

					match := pattern.FindStringSubmatch(n.FirstChild.Data)
					if len(match) < 2 {
						break
					}

					if err := json.Unmarshal([]byte(match[1]), &ingredients); err != nil {
						matchErr = fmt.Errorf("parse recipeIngredient: %w", err)
						break
					}

					found = true
					return
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if found {
				return
			}
			processNode(c)
		}
	}
	processNode(doc)

	if found {
		return strings.Join(ingredients, "\n"), nil
	}
	if matchErr != nil {
		return "", matchErr
	}

	return "", errors.New("recipeIngredient not found in JSON-LD")
}
