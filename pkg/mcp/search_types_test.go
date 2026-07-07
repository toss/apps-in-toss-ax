package mcp

import "testing"

func floatPtr(v float64) *float64 {
	return &v
}

func TestSearchInputSearchOptions_DefaultLimit(t *testing.T) {
	opts := SearchInput{Query: "결제"}.searchOptions()

	if opts.Limit != 10 {
		t.Errorf("Expected default limit 10, got %d", opts.Limit)
	}
	if opts.Boosts.Title != nil || opts.Boosts.Description != nil ||
		opts.Boosts.Content != nil || opts.Boosts.Category != nil {
		t.Errorf("Expected nil boost overrides, got %+v", opts.Boosts)
	}
}

func TestSearchInputSearchOptions_BoostPassthrough(t *testing.T) {
	input := SearchInput{
		Query:            "결제",
		Limit:            5,
		TitleBoost:       floatPtr(2.0),
		DescriptionBoost: floatPtr(0.5),
		ContentBoost:     floatPtr(3.0),
		CategoryBoost:    floatPtr(0),
	}

	opts := input.searchOptions()

	if opts.Limit != 5 {
		t.Errorf("Expected limit 5, got %d", opts.Limit)
	}
	if opts.Boosts.Title == nil || *opts.Boosts.Title != 2.0 {
		t.Errorf("Expected title boost 2.0, got %v", opts.Boosts.Title)
	}
	if opts.Boosts.Description == nil || *opts.Boosts.Description != 0.5 {
		t.Errorf("Expected description boost 0.5, got %v", opts.Boosts.Description)
	}
	if opts.Boosts.Content == nil || *opts.Boosts.Content != 3.0 {
		t.Errorf("Expected content boost 3.0, got %v", opts.Boosts.Content)
	}
	if opts.Boosts.Category == nil || *opts.Boosts.Category != 0 {
		t.Errorf("Expected category boost 0, got %v", opts.Boosts.Category)
	}
}
