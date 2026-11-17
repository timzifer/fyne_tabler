package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
)

const (
	registry   = "https://registry.npmjs.org/@tabler/icons/"
	aliasesURL = "https://raw.githubusercontent.com/tabler/tabler-icons/v%s/aliases.json"
)

var errNotFound = errors.New("not found")

// release is the subset of the npm registry's version document we need.
type release struct {
	Version string `json:"version"`
	Dist    struct {
		Tarball   string `json:"tarball"`
		Integrity string `json:"integrity"`
	} `json:"dist"`
}

func fetchRelease(ctx context.Context, version string) (*release, error) {
	body, err := get(ctx, registry+url.PathEscape(version))
	if err != nil {
		return nil, err
	}
	var rel release
	if err := json.Unmarshal(body, &rel); err != nil {
		return nil, fmt.Errorf("decode release %q: %w", version, err)
	}
	if rel.Version == "" || rel.Dist.Tarball == "" {
		return nil, fmt.Errorf("release %q: no version or tarball in registry response", version)
	}
	return &rel, nil
}

func download(ctx context.Context, tarball string) ([]byte, error) {
	u, err := url.Parse(tarball)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "https" {
		return nil, fmt.Errorf("refusing non-https tarball URL %q", tarball)
	}
	return get(ctx, tarball)
}

// fetchAliases returns the renames Tabler publishes in aliases.json of the
// given release, keyed by old icon name (filled icons with "-filled" suffix).
// A release without aliases.json yields an empty map.
func fetchAliases(ctx context.Context, version string) (map[string]string, error) {
	body, err := get(ctx, fmt.Sprintf(aliasesURL, url.PathEscape(version)))
	if errors.Is(err, errNotFound) {
		return map[string]string{}, nil
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

func get(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "tablerupdate")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		return io.ReadAll(resp.Body)
	case http.StatusNotFound:
		return nil, fmt.Errorf("GET %s: %w", rawURL, errNotFound)
	default:
		return nil, fmt.Errorf("GET %s: %s", rawURL, resp.Status)
	}
}

// verify checks data against an npm Subresource Integrity string
// ("sha512-<base64>").
func verify(data []byte, integrity string) error {
	want, ok := strings.CutPrefix(integrity, "sha512-")
	if !ok {
		return fmt.Errorf("unsupported integrity %q", integrity)
	}
	sum := sha512.Sum512(data)
	if got := base64.StdEncoding.EncodeToString(sum[:]); got != want {
		return fmt.Errorf("tarball integrity mismatch: got sha512-%s, want %s", got, integrity)
	}
	return nil
}

const filledSuffix = "-filled"

type tablerPackage struct {
	icons   map[string][]byte // name without ".svg" → SVG content
	license []byte
}

var validName = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// extract reads the icons and the license from an npm package tarball.
// Outline icons keep their name, filled icons get the "-filled" suffix.
func extract(tgz []byte) (*tablerPackage, error) {
	zr, err := gzip.NewReader(bytes.NewReader(tgz))
	if err != nil {
		return nil, err
	}
	pkg := &tablerPackage{icons: map[string][]byte{}}
	tr := tar.NewReader(zr)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		var name string
		switch dir, file := path.Split(hdr.Name); {
		case hdr.Name == "package/LICENSE":
			if pkg.license, err = io.ReadAll(tr); err != nil {
				return nil, err
			}
			continue
		case dir == "package/icons/outline/" && strings.HasSuffix(file, ".svg"):
			name = strings.TrimSuffix(file, ".svg")
		case dir == "package/icons/filled/" && strings.HasSuffix(file, ".svg"):
			name = strings.TrimSuffix(file, ".svg") + filledSuffix
		default:
			continue
		}

		if !validName.MatchString(name) {
			return nil, fmt.Errorf("%s: unexpected icon name %q", hdr.Name, name)
		}
		if _, dup := pkg.icons[name]; dup {
			return nil, fmt.Errorf("%s: icon %q exists twice", hdr.Name, name)
		}
		src, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		if !bytes.Contains(src, []byte("currentColor")) {
			return nil, fmt.Errorf("%s: icon does not use currentColor and cannot be stained", hdr.Name)
		}
		pkg.icons[name] = src
	}

	if len(pkg.icons) == 0 {
		return nil, errors.New("tarball contains no icons")
	}
	if pkg.license == nil {
		return nil, errors.New("tarball contains no LICENSE")
	}
	return pkg, nil
}
