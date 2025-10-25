package rt

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	Pi = 3.1415926535897932385
)

func DegreesToRadians(degrees float64) float64 {
	return degrees * Pi / 180.0
}

func RandomDouble() float64 {
	return rand.Float64()
}

func RandomDoubleRange(min, max float64) float64 {
	return min + (max-min)*RandomDouble()
}
func RandomInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

// PrintRenderSettings displays all camera and scene settings before rendering
func PrintRenderSettings(camera *Camera, objectCount int) {
	fmt.Println("\n========================================")
	fmt.Println("           RENDER SETTINGS")
	fmt.Println("========================================")
	fmt.Printf("Image Size:         %dx%d\n", camera.ImageWidth, camera.ImageHeight)
	fmt.Printf("Aspect Ratio:       %.2f:1\n", camera.AspectRatio)
	fmt.Printf("Samples Per Pixel:  %d\n", camera.SamplesPerPixel)
	fmt.Printf("Max Bounce Depth:   %d\n", camera.MaxDepth)
	fmt.Printf("Field of View:      %.1f°\n", camera.Vfov)
	fmt.Printf("Defocus Angle:      %.2f°\n", camera.DefocusAngle)
	fmt.Printf("Focus Distance:     %.1f\n", camera.FocusDist)
	fmt.Printf("Camera Position:    (%.1f, %.1f, %.1f)\n", camera.LookFrom.X, camera.LookFrom.Y, camera.LookFrom.Z)
	fmt.Printf("Camera Target:      (%.1f, %.1f, %.1f)\n", camera.LookAt.X, camera.LookAt.Y, camera.LookAt.Z)
	fmt.Printf("Camera Motion Blur: %t\n", camera.CameraMotion)
	fmt.Printf("Objects in Scene:   %d\n", objectCount)

	fmt.Println("========================================")
}

// FormatDuration converts a duration to human-readable format
func FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// PrintRenderStats displays render completion statistics
func PrintRenderStats(elapsed time.Duration, imageWidth, imageHeight int) {
	fmt.Println("\n========================================")
	fmt.Println("         RENDER COMPLETE!")
	fmt.Println("========================================")
	fmt.Printf("Total Render Time:  %s\n", FormatDuration(elapsed))
	totalPixels := imageWidth * imageHeight
	fmt.Printf("Total Pixels:       %d\n", totalPixels)
	fmt.Printf("Average per pixel:  %.2fms\n", elapsed.Seconds()*1000/float64(totalPixels))
	fmt.Println("========================================")
}
