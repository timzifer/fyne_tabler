// Command tablerupdate imports a release of Tabler Icons from the npm registry
// (package @tabler/icons) and replaces the embedded SVG files.
//
// Outline icons keep their name, filled icons get a "-filled" suffix
// (icons/outline/home.svg → home.svg, icons/filled/home.svg → home-filled.svg),
// matching the flat naming of Tabler Icons 2.x.
//
// Renamed icons keep their old name as deprecated alias for a number of
// further releases. Renames are taken from Tabler's aliases.json, from names
// ("cart-x" → "shopping-cart-x") and from identical shapes. Imported releases
// and aliases are recorded in tabler.json; aliases can be added there by hand
// (source "manual"), followed by a run with -offline.
//
// Every imported release adds a section to CHANGELOG.md listing renamed,
// removed, added and changed icons. The upstream license and the version in
// the README are updated as well.
//
// It runs as part of go generate, before iconconst regenerates the Icon*
// variables:
//
//	//go:generate go run ./internal/tablerupdate
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type options struct {
	version   string
	dir       string
	license   string
	readme    string
	stateFile string
	aliases   string
	changelog string
	notes     string
	pkg       string
	offline   bool
}

func main() {
	var o options
	flag.StringVar(&o.version, "version", "latest", `release to import, e.g. "3.52.0", or "latest"`)
	flag.StringVar(&o.dir, "dir", "icons", "directory receiving the *.svg files")
	flag.StringVar(&o.license, "license", "LICENSE-tabler-icons", `upstream license file ("" to skip)`)
	flag.StringVar(&o.readme, "readme", "README.md", `README whose "Tabler Icons (vX.Y.Z)" is updated`)
	flag.StringVar(&o.stateFile, "state", "tabler.json", "imported releases and deprecated aliases")
	flag.StringVar(&o.aliases, "out", "tabler_gen.go", "generated Go file with TablerVersion and the deprecated aliases")
	flag.StringVar(&o.changelog, "changelog", "CHANGELOG.md", `changelog receiving the release notes ("" to skip)`)
	flag.StringVar(&o.notes, "notes", "", "also write the release notes of an imported release to this file, shortened to fit a pull request body")
	flag.StringVar(&o.pkg, "pkg", "", "package name of the generated file (default $GOPACKAGE or fyne_tabler)")
	flag.BoolVar(&o.offline, "offline", false, "only regenerate the -out file from the state file and the icons directory")
	flag.Parse()
	if o.pkg == "" {
		o.pkg = os.Getenv("GOPACKAGE")
	}
	if o.pkg == "" {
		o.pkg = "fyne_tabler"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := run(ctx, o); err != nil {
		fmt.Fprintln(os.Stderr, "tablerupdate:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, o options) error {
	st, err := loadState(o.stateFile, o.readme)
	if err != nil {
		return err
	}
	old, err := readIcons(o.dir)
	if err != nil {
		return err
	}
	if o.offline {
		return writeAliases(o, st, old)
	}

	rel, err := fetchRelease(ctx, o.version)
	if err != nil {
		return err
	}
	data, err := download(ctx, rel.Dist.Tarball)
	if err != nil {
		return err
	}
	if err := verify(data, rel.Dist.Integrity); err != nil {
		return err
	}
	pkg, err := extract(data)
	if err != nil {
		return err
	}
	tablerAliases, err := fetchAliases(ctx, rel.Version)
	if err != nil {
		return err
	}

	d := compare(old, pkg.icons, tablerAliases, st.Renames)
	from := st.Version
	newRelease := rel.Version != st.Version
	if newRelease {
		if slices.Contains(st.Releases, rel.Version) {
			return fmt.Errorf("release %s was imported before %s already", rel.Version, st.Version)
		}
		st.Version = rel.Version
		st.Releases = append(st.Releases, rel.Version)
	}
	aliased, retargeted, dropped := st.apply(pkg.icons, d)

	// Validate the aliases before touching any file.
	src, err := generateAliases(o.pkg, st, keys(pkg.icons))
	if err != nil {
		return err
	}
	if err := writeIcons(o.dir, pkg.icons); err != nil {
		return err
	}
	if err := os.WriteFile(o.aliases, src, 0o644); err != nil {
		return err
	}
	if err := st.save(o.stateFile); err != nil {
		return err
	}
	if o.license != "" {
		if err := os.WriteFile(o.license, pkg.license, 0o644); err != nil {
			return err
		}
	}
	if err := updateReadme(o.readme, st.Version); err != nil {
		return err
	}

	fmt.Printf("tablerupdate: Tabler Icons %s, %d icons (%d added, %d removed, %d renamed, %d changed, %d aliases dropped)\n",
		st.Version, len(pkg.icons), len(d.Added), len(d.Removed), len(aliased), len(d.Changed), len(dropped))
	if !newRelease {
		return nil
	}

	n := notes{From: from, To: st.Version, Total: len(pkg.icons), Diff: d, Aliased: aliased,
		Retargeted: retargeted, Dropped: dropped}
	if o.changelog != "" {
		if err := prependChangelog(o.changelog, n.markdown(true)); err != nil {
			return err
		}
	}
	if o.notes != "" {
		md := n.markdown(true)
		if len(md) > maxNotes {
			md = n.markdown(false)
		}
		if err := os.WriteFile(o.notes, []byte(md), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func writeAliases(o options, st *state, icons map[string][]byte) error {
	src, err := generateAliases(o.pkg, st, keys(icons))
	if err != nil {
		return err
	}
	return os.WriteFile(o.aliases, src, 0o644)
}

func keys(icons map[string][]byte) map[string]bool {
	names := make(map[string]bool, len(icons))
	for name := range icons {
		names[name] = true
	}
	return names
}

// readIcons returns the *.svg files of dir by icon name.
func readIcons(dir string) (map[string][]byte, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return map[string][]byte{}, nil
	}
	if err != nil {
		return nil, err
	}
	icons := map[string][]byte{}
	for _, e := range entries {
		name, ok := strings.CutSuffix(e.Name(), ".svg")
		if !ok || e.IsDir() {
			continue
		}
		if icons[name], err = os.ReadFile(filepath.Join(dir, e.Name())); err != nil {
			return nil, err
		}
	}
	return icons, nil
}

// writeIcons makes dir contain exactly the given icons. Only *.svg files are
// written or removed; unchanged files are not touched.
func writeIcons(dir string, icons map[string][]byte) error {
	old, err := readIcons(dir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for name := range old {
		if _, keep := icons[name]; !keep {
			if err := os.Remove(filepath.Join(dir, name+".svg")); err != nil {
				return err
			}
		}
	}
	for _, name := range sortedKeys(icons) {
		if string(old[name]) == string(icons[name]) && old[name] != nil {
			continue
		}
		if err := os.WriteFile(filepath.Join(dir, name+".svg"), icons[name], 0o644); err != nil {
			return err
		}
	}
	return nil
}
