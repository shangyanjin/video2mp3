package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const (
	// Default directories
	defaultOutputDir = "output"
	tempDir          = "temp"

	// Start worker goroutines
	numWorkers = 4

	// FFmpeg related paths
	ffmpegRootDir = "ffmpeg"
	ffmpegBinDir  = ffmpegRootDir + "/bin"
	ffmpegExe     = ffmpegBinDir + "/ffmpeg.exe"
	ffprobeExe    = ffmpegBinDir + "/ffprobe.exe"

	// Error messages
	errFFmpegNotFound = "FFmpeg not found. Please ensure FFmpeg is installed in the correct directory"

	// Progress bar settings
	progressWidth = 40

	// Default settings
	defaultScreenshotTime = 1.0 // Default screenshot at 1 second
	defaultFrameCount     = 25  // Default number of frames for animated cover
	defaultFrameInterval  = 0.5 // Default interval between frames in seconds

	// Application info
	appName    = "Video2MP3"
	appVersion = "v2025.02.01"
	appDesc    = "Convert video to MP3 with cover image"
)

//go:embed assets/ffmpeg.exe
var ffmpegBinary []byte

//go:embed assets/ffprobe.exe
var ffprobeBinary []byte

type ConversionStatus struct {
	total     int
	completed int
	mutex     sync.Mutex
}

func (cs *ConversionStatus) increment() {
	cs.mutex.Lock()
	cs.completed++
	cs.displayProgress()
	cs.mutex.Unlock()
}

func (cs *ConversionStatus) displayProgress() {
	percentage := float64(cs.completed) * 100 / float64(cs.total)
	completed := int(float64(progressWidth) * float64(cs.completed) / float64(cs.total))

	fmt.Printf("\r[")
	for i := 0; i < progressWidth; i++ {
		if i < completed {
			fmt.Print("=")
		} else if i == completed {
			fmt.Print(">")
		} else {
			fmt.Print(" ")
		}
	}
	fmt.Printf("] %.1f%% (%d/%d)", percentage, cs.completed, cs.total)
}

func ensureFFmpeg() error {
	// Create temp directory if it doesn't exist
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}

	// Set FFmpeg paths to temp directory
	ffmpegPath := filepath.Join(tempDir, "ffmpeg.exe")
	ffprobePath := filepath.Join(tempDir, "ffprobe.exe")

	// Write FFmpeg binaries to temp directory
	if err := os.WriteFile(ffmpegPath, ffmpegBinary, 0755); err != nil {
		return fmt.Errorf("failed to write ffmpeg binary: %v", err)
	}
	if err := os.WriteFile(ffprobePath, ffprobeBinary, 0755); err != nil {
		return fmt.Errorf("failed to write ffprobe binary: %v", err)
	}

	return nil
}

func cleanup() {
	os.RemoveAll(tempDir)
}

func main() {
	// Define command line arguments
	inputDir := flag.String("d", ".", "Input directory path")
	outputDir := flag.String("o", "", "Output directory path")
	screenshotTime := flag.Float64("t", defaultScreenshotTime, "Screenshot time in seconds (e.g., 3.5)")
	flag.Parse()

	// Determine actual output directory
	actualOutputDir := defaultOutputDir
	if *outputDir != "" {
		actualOutputDir = *outputDir
	}

	// Print application info with directories
	fmt.Printf("%s %s - %s\n", appName, appVersion, appDesc)
	fmt.Printf("Input: %s\nOutput: %s\n\n", *inputDir, actualOutputDir)

	// Extract FFmpeg binaries
	if err := ensureFFmpeg(); err != nil {
		fmt.Printf("Error setting up FFmpeg: %v\n", err)
		return
	}
	defer cleanup()

	// Validate screenshot time
	if *screenshotTime < 0 {
		fmt.Println("Screenshot time must be positive")
		return
	}

	if err := run(*inputDir, *outputDir, *screenshotTime); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func run(inputDir, outputDir string, screenshotTime float64) error {
	// If output directory is empty, use default output directory
	if outputDir == "" {
		outputDir = defaultOutputDir
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Count total video files first
	totalFiles := 0
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			switch ext {
			case ".mp4", ".avi", ".mov", ".mkv":
				totalFiles++
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to count files: %v", err)
	}

	status := &ConversionStatus{
		total: totalFiles,
	}

	// Create channels
	videoFiles := make(chan string)
	errors := make(chan error, 1)
	done := make(chan bool)

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		go func() {
			for relPath := range videoFiles {
				if err := convertToMP3(relPath, inputDir, outputDir, screenshotTime); err != nil {
					fmt.Printf("\nConversion failed for %s: %v\n", relPath, err)
				}
				status.increment()
			}
			done <- true
		}()
	}

	// Walk through directory in a separate goroutine
	go func() {
		err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				ext := strings.ToLower(filepath.Ext(path))
				switch ext {
				case ".mp4", ".avi", ".mov", ".mkv":
					// Get relative path from input directory
					relPath, err := filepath.Rel(inputDir, path)
					if err != nil {
						return fmt.Errorf("failed to get relative path: %v", err)
					}
					videoFiles <- relPath
				}
			}
			return nil
		})

		if err != nil {
			errors <- fmt.Errorf("failed to traverse directory: %v", err)
		}
		close(videoFiles)
	}()

	// Wait for all workers to finish
	for i := 0; i < numWorkers; i++ {
		<-done
	}

	// Print final newline
	fmt.Println()

	// Check if there were any errors during directory traversal
	select {
	case err := <-errors:
		return err
	default:
		return nil
	}
}

func convertToMP3(relPath, inputDir, outputDir string, screenshotTime float64) error {
	// Get full input path
	videoPath := filepath.Join(inputDir, relPath)

	// Create output directory structure
	outPath := filepath.Join(outputDir, filepath.Dir(relPath))
	if err := os.MkdirAll(outPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory structure: %v", err)
	}

	// Get filename without extension
	filename := filepath.Base(videoPath)
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Temporary cover image path
	coverPath := filepath.Join(tempDir, nameWithoutExt+"-cover.jpg")
	// Output MP3 path preserving directory structure
	mp3Path := filepath.Join(outPath, nameWithoutExt+".mp3")

	ffmpegPath := filepath.Join(tempDir, "ffmpeg.exe")
	startTime := fmt.Sprintf("%f", screenshotTime)

	// Extract single frame as cover with high quality
	cmdCover := exec.Command(ffmpegPath,
		"-ss", startTime,
		"-i", videoPath,
		"-vframes", "1",
		"-vf", "scale=1024:-1", // Increased resolution to 1024px width
		"-q:v", "1", // Highest quality (1-31, lower is better)
		"-qmin", "1", // Force minimum quantization to highest quality
		"-qmax", "1", // Force maximum quantization to highest quality
		"-y", // Overwrite output file
		coverPath)

	if err := cmdCover.Run(); err != nil {
		return fmt.Errorf("failed to extract cover: %v", err)
	}

	// Convert to MP3 and add cover
	cmdMP3 := exec.Command(ffmpegPath,
		"-i", videoPath,
		"-i", coverPath,
		"-map", "0:a",
		"-map", "1",
		"-c:a", "libmp3lame",
		"-q:a", "0", // Highest quality MP3
		"-id3v2_version", "3",
		"-metadata:s:v", "title=Album cover",
		"-metadata:s:v", "comment=Cover (front)",
		"-disposition:v:0", "attached_pic",
		"-metadata", "title="+nameWithoutExt,
		"-y", // Overwrite output file
		mp3Path)

	if err := cmdMP3.Run(); err != nil {
		os.Remove(coverPath) // Clean up cover image
		return fmt.Errorf("failed to convert to MP3: %v", err)
	}

	// Delete temporary cover image
	os.Remove(coverPath)

	return nil
}
