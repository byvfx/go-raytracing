package main

import (
	"flag"
	"fmt"
	"go-raytracing/rt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// Profiling flags
	enableProfile := flag.Bool("profile", false, "Enable profiling (CPU, memory)")
	cpuProfile := flag.Bool("cpu-profile", true, "Enable CPU profiling (requires -profile)")
	memProfile := flag.Bool("mem-profile", true, "Enable memory profiling (requires -profile)")
	traceProfile := flag.Bool("trace", false, "Enable execution tracing (requires -profile)")
	blockProfile := flag.Bool("block-profile", false, "Enable block profiling (requires -profile)")
	profileDir := flag.String("profile-dir", "profiles", "Directory to save profile files")
	showMemStats := flag.Bool("mem-stats", false, "Show memory statistics after render")

	flag.Parse()

	// Configure profiler
	profileConfig := &rt.ProfileConfig{
		Enabled:      *enableProfile,
		CPUProfile:   *cpuProfile,
		MemProfile:   *memProfile,
		TraceEnabled: *traceProfile,
		BlockProfile: *blockProfile,
		OutputDir:    *profileDir,
		SampleRate:   100,
	}

	profiler := rt.NewProfiler(profileConfig)

	// Start profiling if enabled
	if *enableProfile {
		fmt.Println("🔬 Profiling enabled")
		if err := profiler.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start profiler: %v\n", err)
			os.Exit(1)
		}

		// Handle graceful shutdown for profiling
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigChan
			fmt.Println("\n⚠️  Interrupt received, saving profiles...")
			profiler.Stop()
			profiler.PrintTimingReport()
			if *showMemStats {
				rt.PrintMemStats()
			}
			os.Exit(0)
		}()
	}

	// Reset render stats
	rt.ResetRenderStats()

	// Time BVH construction
	bvhTimer := rt.NewTimer("BVH Construction")
	world, camera := rt.CornellBoxLucy()
	bvh := rt.NewBVHNodeFromList(world)
	bvhTime := bvhTimer.Stop()
	rt.GlobalRenderStats.BVHConstructTime = bvhTime

	rt.PrintRenderSettings(camera, len(world.Objects))

	bucketSize := 32
	numWorkers := runtime.NumCPU()

	renderer := rt.NewBucketRenderer(camera, bvh, bucketSize, numWorkers)

	// renderer := rt.NewProgressiveRenderer(camera, bvh)

	ebiten.SetWindowSize(camera.ImageWidth, camera.ImageHeight)
	ebiten.SetWindowTitle("Go Raytracer")

	if err := ebiten.RunGame(renderer); err != nil {
		panic(err)
	}

	// Stop profiling and print reports
	if *enableProfile {
		profiler.Stop()
		profiler.PrintTimingReport()
	}

	if *showMemStats {
		rt.PrintMemStats()
	}
}
