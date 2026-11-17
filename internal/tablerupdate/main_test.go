package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha512"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const svg = `<svg stroke="currentColor"></svg>`

func tarball(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(zw)
	for name, content := range files {
		hdr := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(content)), Typeflag: tar.TypeReg}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestExtract(t *testing.T) {
	pkg, err := extract(tarball(t, map[string]string{
		"package/LICENSE":                   "MIT",
		"package/icons/outline/home.svg":    svg,
		"package/icons/filled/home.svg":     `<svg fill="currentColor"></svg>`,
		"package/categories/outline/x.svg":  svg,
		"package/icons/outline/readme.txt":  "ignored",
		"package/icons/outline/sub/sub.svg": svg,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if string(pkg.license) != "MIT" {
		t.Errorf("license = %q", pkg.license)
	}
	if len(pkg.icons) != 2 || string(pkg.icons["home"]) != svg || pkg.icons["home-filled"] == nil {
		t.Errorf("icons = %v", pkg.icons)
	}
}

func TestExtractRejects(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"collision": {
			"package/LICENSE":                       "MIT",
			"package/icons/outline/home-filled.svg": svg,
			"package/icons/filled/home.svg":         svg,
		},
		"bad name":      {"package/LICENSE": "MIT", "package/icons/outline/Home.svg": svg},
		"not stainable": {"package/LICENSE": "MIT", "package/icons/outline/home.svg": `<svg stroke="#000"></svg>`},
		"no icons":      {"package/LICENSE": "MIT"},
		"no license":    {"package/icons/outline/home.svg": svg},
	} {
		if _, err := extract(tarball(t, files)); err == nil {
			t.Errorf("%s: extract succeeded", name)
		}
	}
}

func TestVerify(t *testing.T) {
	data := []byte("data")
	sum := sha512.Sum512(data)
	good := "sha512-" + base64.StdEncoding.EncodeToString(sum[:])
	if err := verify(data, good); err != nil {
		t.Error(err)
	}
	if err := verify([]byte("other"), good); err == nil {
		t.Error("verify accepted wrong data")
	}
	if err := verify(data, "sha1-abc"); err == nil {
		t.Error("verify accepted unsupported algorithm")
	}
}

func TestWriteIcons(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{"keep.svg": "old", "same.svg": "same", "gone.svg": "x", "other.txt": "x"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	err := writeIcons(dir, map[string][]byte{
		"keep": []byte("new"),
		"same": []byte("same"),
		"new":  []byte("new"),
	})
	if err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if got := strings.Join(names, " "); got != "keep.svg new.svg other.txt same.svg" {
		t.Errorf("files = %s", got)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "keep.svg")); string(b) != "new" {
		t.Errorf("keep.svg = %q", b)
	}
}

func TestUpdateReadme(t *testing.T) {
	file := filepath.Join(t.TempDir(), "README.md")
	if err := os.WriteFile(file, []byte("[Tabler Icons](https://tabler.io/icons) (v2.30.0) for"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := updateReadme(file, "3.52.0"); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(file); string(b) != "[Tabler Icons](https://tabler.io/icons) (v3.52.0) for" {
		t.Errorf("README = %q", b)
	}
	if v, err := readmeVersionOf(file); err != nil || v != "3.52.0" {
		t.Errorf("readmeVersionOf = %q, %v", v, err)
	}
	if err := os.WriteFile(file, []byte("no version"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := updateReadme(file, "3.52.0"); err == nil {
		t.Error("updateReadme accepted README without version")
	}
}
