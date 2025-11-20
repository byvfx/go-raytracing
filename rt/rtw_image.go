package rt

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"slices"
)

// RtwImage manages loading and accessing image data for textures
// C++: class rtw_image
type RtwImage struct {
	data             []Color
	imageWidth       int
	imageHeight      int
	bytesPerPixel    int
	bytesPerScanline int
}

// NewRtwImage creates an empty image
func NewRtwImage() *RtwImage {
	return &RtwImage{
		bytesPerPixel: 3,
	}
}

func NewRtwImageFromFile(filename string) *RtwImage {
	img := NewRtwImage()

	imagedir := os.Getenv("RTW_IMAGES")

	searchPaths := []string{}

	if imagedir != "" {
		searchPaths = append(searchPaths, filepath.Join(imagedir, filename))
	}

	searchPaths = append(searchPaths,
		filename,
		filepath.Join("images", filename),
		filepath.Join("..", "images", filename),
		filepath.Join("..", "..", "images", filename),
		filepath.Join("..", "..", "..", "images", filename),
		filepath.Join("..", "..", "..", "..", "images", filename),
		filepath.Join("..", "..", "..", "..", "..", "images", filename),
		filepath.Join("..", "..", "..", "..", "..", "..", "images", filename),
	)

	if slices.ContainsFunc(searchPaths, img.Load) {
		return img
	}

	fmt.Fprintf(os.Stderr, "ERROR: Could not load image file '%s'.\n", filename)
	return img
}

// Load loads the linear (gamma=1) image data from the given file
func (img *RtwImage) Load(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()

	// Decode image (automatically detects format)
	decoded, _, err := image.Decode(file)
	if err != nil {
		return false
	}

	bounds := decoded.Bounds()
	img.imageWidth = bounds.Dx()
	img.imageHeight = bounds.Dy()
	img.bytesPerScanline = img.imageWidth * img.bytesPerPixel

	// Convert image to linear color space [0.0, 1.0]
	totalPixels := img.imageWidth * img.imageHeight
	img.data = make([]Color, totalPixels)

	for y := 0; y < img.imageHeight; y++ {
		for x := 0; x < img.imageWidth; x++ {
			r, g, b, _ := decoded.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()

			// RGBA() returns values in range [0, 65535], convert to [0.0, 1.0]
			// Also apply inverse gamma correction (assume sRGB input, gamma ≈ 2.2)
			idx := y*img.imageWidth + x
			img.data[idx] = Color{
				X: LinearToGamma(float64(r) / 65535.0),
				Y: LinearToGamma(float64(g) / 65535.0),
				Z: LinearToGamma(float64(b) / 65535.0),
			}
		}
	}

	return true
}

func (img *RtwImage) Width() int {
	if img.data == nil {
		return 0
	}
	return img.imageWidth
}

func (img *RtwImage) Height() int {
	if img.data == nil {
		return 0
	}
	return img.imageHeight
}

func (img *RtwImage) PixelData(x, y int) Color {
	// Return magenta if no image data
	if img.data == nil {
		return Color{X: 1.0, Y: 0.0, Z: 1.0}
	}

	// Clamp to valid range
	x = clamp(x, 0, img.imageWidth)
	y = clamp(y, 0, img.imageHeight)

	idx := y*img.imageWidth + x
	return img.data[idx]
}

func clamp(x, low, high int) int {
	if x < low {
		return low
	}
	if x < high {
		return x
	}
	return high - 1
}
