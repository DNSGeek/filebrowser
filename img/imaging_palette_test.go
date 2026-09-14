package img

import (
	"image"
	"image/color"
	"testing"

	"github.com/boxes-ltd/imaging"
)

// TestResizePalettedImageWithShortPalette guards against CVE-2023-36308: a
// paletted image (as produced by a crafted TIFF) whose pixels index past the
// end of its palette made disintegration/imaging panic while resizing.
func TestResizePalettedImageWithShortPalette(t *testing.T) {
	img := image.NewPaletted(image.Rect(0, 0, 64, 64), color.Palette{color.Black, color.White})
	for i := range img.Pix {
		img.Pix[i] = 200
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("resizing a paletted image with out-of-range indices panicked: %v", r)
		}
	}()

	if got := imaging.Fit(img, 16, 16, imaging.Lanczos).Bounds().Dx(); got != 16 {
		t.Fatalf("Fit width = %d, want 16", got)
	}
	if got := imaging.Fill(img, 16, 16, imaging.Center, imaging.Box).Bounds().Dx(); got != 16 {
		t.Fatalf("Fill width = %d, want 16", got)
	}
}
