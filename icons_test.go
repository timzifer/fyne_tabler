package fyne_tabler

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

// TestAllIconsRender makes sure every embedded icon can be stained and
// rasterized and actually draws something.
func TestAllIconsRender(t *testing.T) {
	all := All()
	if len(all) < 500 {
		t.Fatalf("only %d icons embedded", len(all))
	}
	for _, i := range all {
		data, err := PNG(i, 24, color.White)
		if err != nil {
			t.Errorf("PNG(%s): %v", i, err)
			continue
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Errorf("PNG(%s) is not decodable: %v", i, err)
			continue
		}
		if isBlank(img) {
			t.Errorf("PNG(%s) is blank", i)
		}
	}
}

func isBlank(img image.Image) bool {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if _, _, _, a := img.At(x, y).RGBA(); a != 0 {
				return false
			}
		}
	}
	return true
}

func TestLookup(t *testing.T) {
	if i, ok := Lookup("home"); !ok || i != IconHome {
		t.Errorf("Lookup(home) = %v, %v", i, ok)
	}
	if i, ok := Lookup(IconHome.String()); !ok || i != IconHome {
		t.Errorf("Lookup(String()) = %v, %v", i, ok)
	}
	for _, name := range []string{"", "does-not-exist", "../icons/home"} {
		if _, ok := Lookup(name); ok {
			t.Errorf("Lookup(%q) succeeded", name)
		}
	}
}

func TestIcon(t *testing.T) {
	test.NewApp()
	defer test.NewApp()

	res := Icon(IconHome, color.NRGBA{R: 0xff, A: 0xff})
	stainable, ok := res.(StainableResource)
	if !ok {
		t.Fatalf("Icon() returned %T, want StainableResource", res)
	}
	if bytes.Contains(res.Content(), []byte("currentColor")) || !bytes.Contains(res.Content(), []byte("#ff0000")) {
		t.Errorf("icon is not stained: %s", res.Content())
	}
	if blue := stainable.MustStain(color.NRGBA{B: 0xff, A: 0xff}); !bytes.Contains(blue.Content(), []byte("#0000ff")) {
		t.Errorf("MustStain() did not restain: %s", blue.Content())
	}
	if !bytes.Contains(Source(IconHome), []byte("currentColor")) {
		t.Error("Source() must return the unstained SVG")
	}

	var zero icon
	if got := Icon(zero); got != theme.ErrorIcon() {
		t.Errorf("Icon(zero value) = %v, want theme.ErrorIcon()", got)
	}
	if Source(zero) != nil {
		t.Error("Source(zero value) must be nil")
	}
}
