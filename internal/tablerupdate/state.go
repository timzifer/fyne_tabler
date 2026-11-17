package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

// state is persisted in tabler.json. It records which Tabler releases were
// imported and the deprecated aliases of renamed icons.
type state struct {
	// Version is the Tabler Icons release currently embedded.
	Version string `json:"version"`
	// Releases lists all imported Tabler releases, oldest first.
	Releases []string `json:"releases"`
	// Renames are manual hints (old name → new name) for renames that are
	// neither published by Tabler nor detected automatically. They are
	// consulted whenever a release removes the old name.
	Renames map[string]string `json:"renames,omitempty"`
	// Aliases are the deprecated names of renamed icons, sorted by name.
	Aliases []alias `json:"aliases"`
}

type alias struct {
	Name   string `json:"name"`   // old icon name
	Target string `json:"target"` // current icon name
	Since  string `json:"since"`  // release that renamed the icon
	Source string `json:"source"` // how the rename was found: tabler, name, shape or manual
}

// loadState reads the state file. Without one, the state is bootstrapped from
// the version mentioned in the README.
func loadState(file, readme string) (*state, error) {
	data, err := os.ReadFile(file)
	if errors.Is(err, os.ErrNotExist) {
		version, err := readmeVersionOf(readme)
		if err != nil {
			return nil, fmt.Errorf("no %s and %w", file, err)
		}
		return &state{Version: version, Releases: []string{version}}, nil
	}
	if err != nil {
		return nil, err
	}
	var st state
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	if st.Version == "" || !slices.Contains(st.Releases, st.Version) {
		return nil, fmt.Errorf("%s: version %q must be listed in releases", file, st.Version)
	}
	for _, a := range st.Aliases {
		if a.Name == "" || a.Target == "" || !slices.Contains(st.Releases, a.Since) {
			return nil, fmt.Errorf("%s: alias %+v needs name, target and a since listed in releases", file, a)
		}
	}
	return &st, nil
}

func (st *state) save(file string) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(st); err != nil {
		return err
	}
	return os.WriteFile(file, buf.Bytes(), 0o644)
}

// major returns the major version of a Tabler release ("3.52.0" → 3).
func major(version string) int {
	n, _ := strconv.Atoi(strings.SplitN(version, ".", 2)[0])
	return n
}

// expired reports whether the deprecation period of a is over. Like Tabler,
// aliases are kept until the next major release of Tabler Icons.
func (st *state) expired(a alias) bool {
	return major(st.Version) > major(a.Since)
}

// aliasChange is an alias that was dropped or retargeted, with the reason.
type aliasChange struct {
	alias
	Note string
}

// apply updates the aliases for an imported icon set: renames become new
// aliases, aliases follow renamed targets, and aliases whose target vanished,
// whose name is an icon again or whose deprecation period is over are dropped.
func (st *state) apply(icons map[string][]byte, d iconDiff) (added []alias, retargeted, dropped []aliasChange) {
	var kept []alias
	for _, a := range st.Aliases {
		switch {
		case icons[a.Name] != nil:
			dropped = append(dropped, aliasChange{a, "the name is an icon again"})
			continue
		case icons[a.Target] == nil:
			r, ok := d.Renames[a.Target]
			if !ok {
				dropped = append(dropped, aliasChange{a, fmt.Sprintf("its target %q was removed", a.Target)})
				continue
			}
			retargeted = append(retargeted, aliasChange{a, fmt.Sprintf("%q was renamed to %q", a.Target, r.Target)})
			a.Target = r.Target
		}
		if st.expired(a) {
			dropped = append(dropped, aliasChange{a, fmt.Sprintf("deprecated since Tabler Icons %s", a.Since)})
			continue
		}
		kept = append(kept, a)
	}
	for _, name := range d.Removed {
		if r, ok := d.Renames[name]; ok {
			a := alias{Name: name, Target: r.Target, Since: st.Version, Source: r.Source}
			added = append(added, a)
			kept = append(kept, a)
		}
	}
	slices.SortFunc(kept, func(a, b alias) int { return strings.Compare(a.Name, b.Name) })
	st.Aliases = kept
	return added, retargeted, dropped
}
