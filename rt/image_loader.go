package rt

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
)

type ImageLoader struct {
	data             []Color
	imageWidth       int
	imageHeight      int
	bytesPerPixel    int
	bytesPerScanline int
}

// NewImageLoader creates an empty image
func NewImageLoader() *ImageLoader {
	return &ImageLoader{
		bytesPerPixel: 3,
	}
}

func NewImageLoaderFromFile(filename string) *ImageLoader {
	img := NewImageLoader()

	path, err := FindAsset(filename, "images")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Could not resolve image file path '%s'.\n", filename)
		return img
	}
	img.Load(path)
	return img
}

// Load loads the linear (gamma=1) image data from the given file
func (img *ImageLoader) Load(filename string) bool {
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

func (img *ImageLoader) Width() int {
	if img.data == nil {
		return 0
	}
	return img.imageWidth
}

func (img *ImageLoader) Height() int {
	if img.data == nil {
		return 0
	}
	return img.imageHeight
}

func (img *ImageLoader) PixelData(x, y int) Color {
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

// TODO do i need this clamp function i think i can just use math.Min math.Max
func clamp(x, low, high int) int {
	if x < low {
		return low
	}
	if x < high {
		return x
	}
	return high - 1
}

func FindAsset(filename string, assetType string) (string, error) {
	searchPaths := []string{
		filename,
		filepath.Join(assetType, filename),
		filepath.Join("assets", assetType, filename),
		filepath.Join("..", assetType, filename),
		filepath.Join("..", "assets", assetType, filename),
	}
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("asset file '%s' not found in any search paths", filename)
}
