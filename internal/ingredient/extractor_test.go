package ingredient

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestIsBlockedIP(t *testing.T) {
	cases := []struct {
		ip      string
		blocked bool
	}{
		{"127.0.0.1", true},           // loopback
		{"::1", true},                 // loopback v6
		{"10.1.2.3", true},            // private
		{"172.16.5.4", true},          // private
		{"192.168.0.1", true},         // private
		{"169.254.169.254", true},     // link-local / cloud metadata
		{"0.0.0.0", true},             // unspecified
		{"fe80::1", true},             // link-local v6
		{"fc00::1", true},             // unique-local v6 (private)
		{"8.8.8.8", false},            // public
		{"93.184.216.34", false},      // public (example.com)
		{"2606:2800:220:1::1", false}, // public v6
	}
	for _, c := range cases {
		ip := net.ParseIP(c.ip)
		if ip == nil {
			t.Fatalf("bad test IP %q", c.ip)
		}
		if got := isBlockedIP(ip); got != c.blocked {
			t.Errorf("isBlockedIP(%s) = %v, want %v", c.ip, got, c.blocked)
		}
	}
}

func TestSafeGet_BlocksLoopbackServer(t *testing.T) {
	// httptest binds to 127.0.0.1, which the dialer Control hook must refuse.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("secret internal data"))
	}))
	defer srv.Close()

	resp, err := safeGet(context.Background(), srv.URL)
	if err == nil {
		resp.Body.Close()
		t.Fatalf("expected safeGet to block loopback %s, but it connected", srv.URL)
	}
	if !strings.Contains(err.Error(), "ssrf") {
		t.Errorf("expected an ssrf block error, got: %v", err)
	}
}

func TestSafeGet_RejectsNonHTTPScheme(t *testing.T) {
	for _, u := range []string{"file:///etc/passwd", "gopher://127.0.0.1:70/", "ftp://example.com/x"} {
		if _, err := safeGet(context.Background(), u); err == nil {
			t.Errorf("expected scheme rejection for %q, got nil error", u)
		}
	}
}

func TestExtractIngredientText(t *testing.T) {
	// testData := []string{"arla", "ica", "koket", "mathem", "recept"}
	testData := []string{"recept"}
	for _, name := range testData {
		t.Run(name, func(t *testing.T) {
			reader, err := os.Open("testdata/" + name + ".html")
			if err != nil {
				t.Fatalf("open input fixture: %v", err)
			}
			defer reader.Close()

			result, err := extractIngredientText(reader)
			if err != nil {
				t.Fatalf("extract ingredient text: %v", err)
			}
			result = normalizeWhitespace(result)

			expected, err := os.ReadFile("testdata/" + name + ".expected.txt")
			if err != nil {
				t.Fatalf("read expected fixture: %v", err)
			}
			want := normalizeWhitespace(string(expected))

			if result != want {
				t.Errorf("extractIngredientText() = %q\n\nwant %q", result, want)
			}
		})
	}
}

func TestJSONLDContinuesUntilRecipeIngredientMatch(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(`
		<script type="application/ld+json">{"@type":"WebSite"}</script>
		<script type="application/ld+json">{"recipeIngredient":["1 egg","2 dl milk"]}</script>
	`))
	if err != nil {
		t.Fatalf("parse HTML: %v", err)
	}

	got, err := extractJSONLDIngredientText(doc)
	if err != nil {
		t.Fatalf("extract JSON-LD: %v", err)
	}
	if want := "1 egg\n2 dl milk"; got != want {
		t.Errorf("extractJSONLDIngredientText() = %q, want %q", got, want)
	}
}

func TestJSONLDReturnsErrorWhenRecipeIngredientIsMissing(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(`
		<script type="application/ld+json">{"@type":"WebSite"}</script>
		<script type="application/ld+json">{"@type":"Organization"}</script>
	`))
	if err != nil {
		t.Fatalf("parse HTML: %v", err)
	}

	if _, err := extractJSONLDIngredientText(doc); err == nil {
		t.Fatal("extractJSONLDIngredientText() error = nil, want recipeIngredient not found error")
	}
}

func normalizeWhitespace(text string) string {
	return strings.Join(strings.Fields(text), " ")
}
