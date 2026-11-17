package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"slices"
	"strings"
)

// Rename sources, recorded per alias.
const (
	sourceTabler = "tabler" // aliases.json of the Tabler release
	sourceName   = "name"   // new name ends with "-<old name>", e.g. cart-x → shopping-cart-x
	sourceShape  = "shape"  // new icon draws exactly the same shape
	sourceManual = "manual" // renames or aliases added to the state file by hand
)

type rename struct {
	Target       string
	Source       string
	ShapeChanged bool // the new icon looks different than the old one
}

// iconDiff describes how one icon set turns into another.
type iconDiff struct {
	Added   []string // includes rename targets
	Removed []string // includes renamed icons
	Changed []string // icons on both sides whose shape differs
	Renames map[string]rename
}

// compare diffs two icon sets and finds the new names of removed icons:
// first from hints (manual renames), then from Tabler's aliases, the name
// ("cart-x" → "shopping-cart-x", preferring added icons) and identical shapes.
// Only shapes count as changes: reformatted SVGs that draw the same are equal.
func compare(old, cur map[string][]byte, tablerAliases, hints map[string]string) iconDiff {
	d := iconDiff{Renames: map[string]rename{}}
	for name, src := range cur {
		oldSrc, ok := old[name]
		switch {
		case !ok:
			d.Added = append(d.Added, name)
		case !sameShape(oldSrc, src):
			d.Changed = append(d.Changed, name)
		}
	}
	for name := range old {
		if _, ok := cur[name]; !ok {
			d.Removed = append(d.Removed, name)
		}
	}
	slices.Sort(d.Added)
	slices.Sort(d.Removed)
	slices.Sort(d.Changed)

	byShape := map[string][]string{}
	for _, name := range d.Added {
		if s, err := shape(cur[name]); err == nil {
			byShape[s] = append(byShape[s], name)
		}
	}
	all := sortedKeys(cur)
	for _, name := range d.Removed {
		if target, ok := hints[name]; ok && cur[target] != nil {
			d.Renames[name] = rename{Target: target, Source: sourceManual}
			continue
		}
		if target, ok := tablerAliases[name]; ok && cur[target] != nil {
			d.Renames[name] = rename{Target: target, Source: sourceTabler}
			continue
		}
		suffix := func(n string) bool { return strings.HasSuffix(n, "-"+name) }
		if target, ok := unique(d.Added, suffix); ok {
			d.Renames[name] = rename{Target: target, Source: sourceName}
			continue
		}
		if target, ok := unique(all, suffix); ok {
			d.Renames[name] = rename{Target: target, Source: sourceName}
			continue
		}
		if s, err := shape(old[name]); err == nil && len(byShape[s]) == 1 {
			d.Renames[name] = rename{Target: byShape[s][0], Source: sourceShape}
		}
	}
	for name, r := range d.Renames {
		r.ShapeChanged = !sameShape(old[name], cur[r.Target])
		d.Renames[name] = r
	}
	return d
}

func unique(names []string, match func(string) bool) (string, bool) {
	var found []string
	for _, n := range names {
		if match(n) {
			found = append(found, n)
		}
	}
	if len(found) != 1 {
		return "", false
	}
	return found[0], true
}

func sameShape(a, b []byte) bool {
	if bytes.Equal(a, b) {
		return true
	}
	sa, errA := shape(a)
	sb, errB := shape(b)
	return errA == nil && errB == nil && sa == sb
}

// geometry lists the attributes that define what an element draws.
var geometry = []string{"d", "points", "cx", "cy", "r", "rx", "ry", "x", "y", "width", "height", "x1", "y1", "x2", "y2"}

// shape returns a canonical description of the elements an SVG draws,
// ignoring the root element, styling, formatting and Tabler's invisible
// 24×24 bounding-box path.
func shape(svg []byte) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(svg))
	var out strings.Builder
	depth := 0
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if depth == 1 {
				continue
			}
			attrs := map[string]string{}
			for _, a := range t.Attr {
				attrs[a.Name.Local] = strings.Join(strings.Fields(a.Value), " ")
			}
			if t.Name.Local == "path" && attrs["d"] == "M0 0h24v24H0z" {
				continue
			}
			out.WriteString(t.Name.Local)
			for _, g := range geometry {
				if v, ok := attrs[g]; ok {
					out.WriteString(" " + g + "=" + v)
				}
			}
			out.WriteByte('\n')
		case xml.EndElement:
			depth--
		}
	}
	return out.String(), nil
}
