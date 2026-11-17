package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/timzifer/fyne_iconkit/iconupdate"
)

const (
	aliasesURL   = "https://raw.githubusercontent.com/tabler/tabler-icons/v%s/aliases.json"
	filledSuffix = "-filled"
)

// fetchAliases returns the renames Tabler publishes in aliases.json of the
// given release, keyed by old icon name (filled icons with "-filled" suffix).
// A release without aliases.json yields no renames.
func fetchAliases(ctx context.Context, version string, _ map[string][]byte, _ []string) (map[string]string, error) {
	body, err := iconupdate.Fetch(ctx, fmt.Sprintf(aliasesURL, url.PathEscape(version)))
	if errors.Is(err, iconupdate.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return parseAliases(body)
}

func parseAliases(body []byte) (map[string]string, error) {
	var raw struct {
		Outline map[string]string `json:"outline"`
		Filled  map[string]string `json:"filled"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode aliases.json: %w", err)
	}
	aliases := map[string]string{}
	for old, name := range raw.Filled {
		aliases[old+filledSuffix] = name + filledSuffix
	}
	for old, name := range raw.Outline {
		aliases[old] = name
		// A renamed outline icon takes its filled variant along.
		if _, ok := aliases[old+filledSuffix]; !ok {
			aliases[old+filledSuffix] = name + filledSuffix
		}
	}
	return aliases, nil
}
