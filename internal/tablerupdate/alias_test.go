package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func icon(paths ...string) []byte {
	s := `<svg xmlns="http://www.w3.org/2000/svg" stroke="currentColor"><path stroke="none" d="M0 0h24v24H0z" fill="none"/>`
	for _, d := range paths {
		s += `<path d="` + d + `"/>`
	}
	return []byte(s + "</svg>")
}

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

func TestShape(t *testing.T) {
	v2 := []byte(`<svg class="icon icon-tabler-x" stroke="currentColor" fill="none">
  <path stroke="none" d="M0 0h24v24H0z" fill="none"/>
  <path d="M5 12l-2  0" stroke-width="0" fill="currentColor" />
</svg>`)
	v3 := []byte("<svg\n  fill=\"currentColor\"\n  class=\"icons-tabler-filled\"\n>\n  <path d=\"M5 12l-2 0\" />\n</svg>")
	if !sameShape(v2, v3) {
		t.Error("reformatted icon counts as changed")
	}
	if sameShape(v2, icon("M5 12l-3 0")) {
		t.Error("different path counts as same shape")
	}
}

func TestCompare(t *testing.T) {
	old := map[string][]byte{
		"home":      icon("M1"),
		"edit":      icon("M2"),
		"123":       icon("M3"),
		"cart-x":    icon("M4"),
		"tikto":     icon("M5"),
		"gone":      icon("M6"),
		"ambigous":  icon("M7"),
		"hint":      icon("M8"),
		"help":      icon("M8b"),
		"plus":      icon("M9"),
		"cart-plus": icon("M9b"), // existed before: found by the fallback to all icons
	}
	cur := map[string][]byte{
		"home":            icon("M1"),
		"edit":            icon("M2b"),
		"number-123":      icon("M3b"),
		"shopping-cart-x": icon("M4b"),
		"tiktok":          icon("M5"),
		"a":               icon("M7"),
		"b":               icon("M7"),
		"help":            icon("M8b"),
		"cart-plus":       icon("M9b"),
	}
	d := compare(old, cur, map[string]string{"123": "number-123", "gone": "missing", "hint": "wrong"}, map[string]string{"hint": "help"})

	if got := strings.Join(d.Changed, " "); got != "edit" {
		t.Errorf("changed = %s", got)
	}
	if got := strings.Join(d.Removed, " "); got != "123 ambigous cart-x gone hint plus tikto" {
		t.Errorf("removed = %s", got)
	}
	want := map[string]rename{
		"123":    {Target: "number-123", Source: sourceTabler, ShapeChanged: true},
		"cart-x": {Target: "shopping-cart-x", Source: sourceName, ShapeChanged: true},
		"tikto":  {Target: "tiktok", Source: sourceShape},
		"hint":   {Target: "help", Source: sourceManual, ShapeChanged: true},
		"plus":   {Target: "cart-plus", Source: sourceName, ShapeChanged: true},
	}
	if len(d.Renames) != len(want) {
		t.Errorf("renames = %v", d.Renames)
	}
	for k, v := range want {
		if d.Renames[k] != v {
			t.Errorf("renames[%q] = %v, want %v", k, d.Renames[k], v)
		}
	}
}

func TestApply(t *testing.T) {
	st := &state{
		Version:  "4.1.0",
		Releases: []string{"3.5.0", "4.0.0", "4.1.0"},
		Aliases: []alias{
			{Name: "expired", Target: "home", Since: "3.5.0"},
			{Name: "kept", Target: "home", Since: "4.0.0"},
			{Name: "back", Target: "home", Since: "4.0.0"},
			{Name: "orphan", Target: "vanished", Since: "4.0.0"},
			{Name: "follow", Target: "old-target", Since: "4.0.0"},
		},
	}
	icons := map[string][]byte{"home": icon(), "back": icon(), "new-target": icon(), "renamed": icon()}
	d := iconDiff{
		Removed: []string{"old-target", "vanished", "was"},
		Renames: map[string]rename{"old-target": {Target: "new-target", Source: sourceTabler}, "was": {Target: "renamed", Source: sourceShape}},
	}
	added, retargeted, dropped := st.apply(icons, d)

	var names []string
	for _, a := range st.Aliases {
		names = append(names, a.Name+"→"+a.Target)
	}
	if got := strings.Join(names, " "); got != "follow→new-target kept→home old-target→new-target was→renamed" {
		t.Errorf("aliases = %s", got)
	}
	if len(added) != 2 || added[0].Since != "4.1.0" {
		t.Errorf("added = %v", added)
	}
	if len(retargeted) != 1 || retargeted[0].Name != "follow" {
		t.Errorf("retargeted = %v", retargeted)
	}
	if len(dropped) != 3 {
		t.Errorf("dropped = %v", dropped)
	}
	for _, a := range dropped {
		if a.Name == "expired" && a.Note != "deprecated since Tabler Icons 3.5.0" {
			t.Errorf("expired note = %q", a.Note)
		}
	}
}

func TestGenerateAliases(t *testing.T) {
	st := &state{
		Version:  "3.52.0",
		Releases: []string{"2.30.0", "3.52.0"},
		Aliases:  []alias{{Name: "123", Target: "number-123", Since: "3.52.0", Source: sourceTabler}},
	}
	src, err := generateAliases("fyne_tabler", st, map[string]bool{"number-123": true})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"// Deprecated: Tabler Icons 3.52.0 renamed \"123\" to \"number-123\", use IconNumber123. The alias is removed with Tabler Icons 4.",
		"Icon123 = IconNumber123",
		`const TablerVersion = "3.52.0"`,
		`"123": "number-123",`,
	} {
		if !strings.Contains(string(src), want) {
			t.Errorf("generated file lacks %q:\n%s", want, src)
		}
	}

	if _, err := generateAliases("p", st, map[string]bool{}); err == nil {
		t.Error("alias with missing target accepted")
	}
	if _, err := generateAliases("p", st, map[string]bool{"number-123": true, "123": true}); err == nil {
		t.Error("alias shadowing an icon accepted")
	}
}

func TestNotes(t *testing.T) {
	n := notes{
		From: "2.30.0", To: "3.52.0", Total: 3,
		Diff: iconDiff{
			Added:   []string{"fresh", "number-123"},
			Removed: []string{"123", "gone"},
			Changed: []string{"home"},
			Renames: map[string]rename{"123": {Target: "number-123", Source: sourceTabler, ShapeChanged: true}},
		},
		Aliased: []alias{{Name: "123", Target: "number-123", Source: sourceTabler}},
		Dropped: []aliasChange{{alias{Name: "old", Target: "home"}, "deprecated since Tabler Icons 2.30.0"}},
	}
	md := n.markdown()
	for _, want := range []string{
		"## Tabler Icons 3.52.0",
		"Breaking: 2 `Icon*` variables",
		"| `123` | `number-123` ⚠️ | `Icon123` → `IconNumber123` | Tabler aliases |",
		"until Tabler Icons 4",
		"### Removed (1)",
		"- `gone` (`IconGone`)",
		"- `old` (`IconOld`), use `IconHome`: deprecated since Tabler Icons 2.30.0",
		"### Added (1)",
		"- `fresh` (`IconFresh`)",
		"### Changed shape (1)",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("notes lack %q:\n%s", want, md)
		}
	}
	if shorten(md, "CHANGELOG.md") != md {
		t.Error("short notes were shortened")
	}
}

func TestShorten(t *testing.T) {
	long := "## X\n\n### Added (1)\n\n<details>\n<summary>Show icons</summary>\n\n" + strings.Repeat("- `a`\n", maxNotes) + "\n</details>\n\n### Changed shape (1)\n"
	got := shorten(long, "CHANGELOG.md")
	if len(got) > maxNotes || strings.Contains(got, "<details>") ||
		!strings.Contains(got, "### Added (1)\n\nThe list is in [CHANGELOG.md](CHANGELOG.md).\n") || !strings.Contains(got, "### Changed shape (1)") {
		t.Errorf("shorten = %q", got)
	}
}

func TestChangelogSection(t *testing.T) {
	changelog := "# Changelog\n\n## Tabler Icons 3.53.0\n\nnew\n\n## Tabler Icons 3.52.0\n\nold\n"
	for version, want := range map[string]string{
		"3.53.0": "## Tabler Icons 3.53.0\n\nnew\n",
		"3.52.0": "## Tabler Icons 3.52.0\n\nold\n",
	} {
		if got, err := changelogSection(changelog, version); err != nil || got != want {
			t.Errorf("changelogSection(%s) = %q, %v", version, got, err)
		}
	}
	if _, err := changelogSection(changelog, "3.5.0"); err == nil {
		t.Error("missing section accepted")
	}
}

func TestChangelog(t *testing.T) {
	file := filepath.Join(t.TempDir(), "CHANGELOG.md")
	if err := prependChangelog(file, "## B\n"); err != nil {
		t.Fatal(err)
	}
	if err := prependChangelog(file, "## C\n"); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(file); string(b) != "# Changelog\n\n## C\n\n## B\n" {
		t.Errorf("changelog = %q", b)
	}
}
