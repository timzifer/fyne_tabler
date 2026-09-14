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
	names := Names()
	if len(names) < 500 {
		t.Fatalf("only %d icons embedded", len(names))
	}
	for _, name := range names {
		data, err := PNG(name, 24, color.White)
		if err != nil {
			t.Errorf("PNG(%q): %v", name, err)
			continue
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Errorf("PNG(%q) is not decodable: %v", name, err)
			continue
		}
		if isBlank(img) {
			t.Errorf("PNG(%q) is blank", name)
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

func TestIcon(t *testing.T) {
	test.NewApp()
	defer test.NewApp()

	res, err := Icon(Names()[0], color.NRGBA{R: 0xff, A: 0xff})
	if err != nil {
		t.Fatal(err)
	}
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

	if got := MustIcon("does-not-exist"); got != theme.ErrorIcon() {
		t.Errorf("MustIcon(does-not-exist) = %v, want theme.ErrorIcon()", got)
	}
}
