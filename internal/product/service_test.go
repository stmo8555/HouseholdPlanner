package product

import "testing"

func TestCategoryFor(t *testing.T) {
	lookup := map[string]string{
		"fish":         "meat and fish",
		"fish fingers": "frozen",
		"frozen":       "frozen",
		"greek yogurt": "dairy",
		"gröna ärtor":  "fruit & vegetables",
		"frysta":       "frozen",
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
		{name: "Findus fish fingers", want: "frozen"},
		{name: "Arla Greek Yogurt 1kg", want: "dairy"},
		{name: "frysta gröna ärtor", want: "frozen"},
		{name: "unknown product", want: "other"},
	}

	for _, tt := range tests {
		if got := service.categoryFor(tt.name); got != tt.want {
			t.Fatalf("categoryFor(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
