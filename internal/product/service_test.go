package product

import "testing"

func TestCategoryFor(t *testing.T) {
	lookup := map[string]string{
		"fish":         "Meat & fish",
		"fish fingers": "Frozen",
		"frozen":       "Frozen",
		"greek yogurt": "Dairy",
		"gröna ärtor":  "Fruit & veg",
		"frysta":       "Frozen",
	}
	normalized := make(map[string]string, len(lookup))
	for key, category := range lookup {
		normalized[normalizeLookupKey(key)] = category
	}

	service := &Service{FoodCategories: normalized}
	tests := []struct {
		name string
		want string
	}{
		{name: "Findus fish fingers", want: "Frozen"},
		{name: "Arla Greek Yogurt 1kg", want: "Dairy"},
		{name: "frysta gröna ärtor", want: "Frozen"},
		{name: "unknown product", want: "Other"},
	}

	for _, tt := range tests {
		if got := service.categoryFor(tt.name); got != tt.want {
			t.Fatalf("categoryFor(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
