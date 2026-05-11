package recipe

import (
	"context"
	"golang.org/x/net/html"
	"net/http"
	"testing"
)

var recipes = [...]string{
	"https://www.koket.se/ugnsbakad-lax-med-vitloks-och-ortsas",
	"https://www.mathem.se/se/recipes/6873-mari-bergman-chicken-shawarma-med-ortig-vitlokssas/",
	"https://www.arla.se/recept/pannkaka/",
	"https://www.ica.se/recept/flaskfilegryta-med-champinjoner-724256/",
	"https://recept.se/recept/spetskal-teriyaki-med-stekt-schnitzel",
}

var titles = [...]string{
	"Ugnsbakad lax med vitlöks- och örtsås",
	"Chicken shawarma med örtig vitlökssås",
	"Pannkakor",
	"Fläskfilégryta med champinjoner",
	"Spetskål teriyaki med stekt schnitzel",
}

var images = [...]string{
	"https://img.koket.se/standard-mega/ugnsbakad-lax-med-vitloks-och-ortsas-bong.png.webp",
	"https://images.mathem.se/prod/recipes/093d8b00-ac91-4481-96d5-ea5dab856f84.png?fit=bounds&format=auto&optimize=medium&quality=75&width=2000&s=0x5ee146571f8073f9e75a0c90442360087e1b9c72",
	"https://images.arla.com/recordid/CAF6A3FD-D0CB-4979-B54A4866FC4EBDD3/pannkaka.jpg?width=1269&height=715&mode=crop&crop=(0,134,0,-14)&format=webp",
	"https://assets.icanet.se/e_sharpen:80,q_auto,dpr_1.25,w_718,h_718,c_lfill/imagevaultfiles/id_186632/cf_259/flaskfilégryta_med_champinjoner.jpg",
	"https://images.recept.se/images/recipes/spetskal-teriyaki-med-stekt-schnitzel_33650.jpg?fit=crop&crop=focalpoint&auto=format&fp-x=0.44464801176872&fp-y=0.5031578774166&fp-z=1.0647443851965&w=1200&h=1200",
}

func TestFindTitle(t *testing.T) {
	for i, url := range recipes {
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("http error: %v", err)
		}

		doc, err := html.Parse(resp.Body)
		resp.Body.Close()

		if err != nil {
			t.Fatalf("parse error: %v", err)
		}

		title := findTitle(doc)

		if title != titles[i] {
			t.Errorf("got %q, want %q", title, titles[i])
		}
	}
}

func TestFindImage(t *testing.T) {
	for i, url := range recipes {
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("http error: %v", err)
		}

		doc, err := html.Parse(resp.Body)
		resp.Body.Close()

		if err != nil {
			t.Fatalf("parse error: %v", err)
		}

		img := findImg(doc, titles[i])

		if img != images[i] {
			t.Errorf("got %q, want %q", img, images[i])
		}
	}
}

type mockDB struct {
	result []Recipe
}

func (m *mockDB) List(ctx context.Context, hid int) ([]Recipe, map[int][]RecipeIngredient, error) {
	panic("Not implemented for test")
}

func (m *mockDB) AddRecipe(ctx context.Context, hid int, recipe Recipe) (int, error) {
	m.result = append(m.result, recipe)
	return 0, nil
}

func (m *mockDB) AddIngredients(ctx context.Context, recipeIngredients []RecipeIngredient) error {
	return nil
}

func (m *mockDB) Ingredients(ctx context.Context, recipeID, hid int) ([]RecipeIngredient, error) {
	return nil, nil
}

func TestAddRecipe(t *testing.T) {
	var mockDB mockDB
	mockDB.result = make([]Recipe, 0)
	service := Service{repo: &mockDB}

	for i, url := range recipes {
		service.Add(context.Background(), i, url)
	}

	for i, v := range mockDB.result {
		if v.Link != recipes[i] {
			t.Errorf("got %q, want %q", v.Link, recipes[i])
		}

		if v.Title != titles[i] {
			t.Errorf("got %q, want %q", v.Title, titles[i])
		}

		if v.ImgURL != images[i] {
			t.Errorf("got %q, want %q", v.ImgURL, images[i])
		}
	}
}
