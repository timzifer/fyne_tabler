package main

import "testing"

func TestParseAliases(t *testing.T) {
	got, err := parseAliases([]byte(`{"outline": {"circle-0": "circle-number-0", "seeding": "seedling"}, "filled": {"seeding": "plant"}}`))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"circle-0":        "circle-number-0",
		"circle-0-filled": "circle-number-0-filled",
		"seeding":         "seedling",
		"seeding-filled":  "plant-filled",
	}
	if len(got) != len(want) {
		t.Fatalf("aliases = %v", got)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("aliases[%q] = %q, want %q", k, got[k], v)
		}
	}
}
