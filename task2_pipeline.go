package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type imageData struct {
	img      image.Image
	filename string
}

func processImagesPipeline(inputDir, outputDir string) time.Duration {
	start := time.Now()
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.MkdirAll(outputDir, 0755)
	}

	files, _ := filepath.Glob(filepath.Join(inputDir, "*.jpg"))
	numWorkers := runtime.NumCPU()

	// Produce file paths
	paths := make(chan string, numWorkers)
	go func() {
		for _, f := range files {
			paths <- f
		}
		close(paths)
	}()

	// Decode
	decoded := make(chan imageData, numWorkers)
	var wgDecode sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wgDecode.Add(1)
		go func() {
			defer wgDecode.Done()
			for path := range paths {
				file, err := os.Open(path)
				if err == nil {
					img, _, err := image.Decode(file)
					if err == nil {
						decoded <- imageData{img, filepath.Base(path)}
					}
					file.Close()
				}
			}
		}()
	}
	go func() {
		wgDecode.Wait()
		close(decoded)
	}()

	// Greyscale
	filtered := make(chan imageData, numWorkers)
	var wgFilter sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wgFilter.Add(1)
		go func() {
			defer wgFilter.Done()
			for data := range decoded {
				bounds := data.img.Bounds()
				grayImg := image.NewGray(bounds)
				for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
					for x := bounds.Min.X; x < bounds.Max.X; x++ {
						grayImg.Set(x, y, color.GrayModel.Convert(data.img.At(x, y)))
					}
				}
				filtered <- imageData{grayImg, data.filename}
			}
		}()
	}
	go func() {
		wgFilter.Wait()
		close(filtered)
	}()

	// Watermark & Encode
	var wgEncode sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wgEncode.Add(1)
		go func() {
			defer wgEncode.Done()
			for data := range filtered {
				bounds := data.img.Bounds()
				watermarked := image.NewRGBA(bounds)
				// Redraw gray pixels into rgba for watermarking
				for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
					for x := bounds.Min.X; x < bounds.Max.X; x++ {
						watermarked.Set(x, y, data.img.At(x, y))
					}
				}
				drawLab2(watermarked, bounds.Max.X-100, bounds.Max.Y-40, color.RGBA{255, 0, 0, 255})

				outPath := filepath.Join(outputDir, data.filename)
				outFile, err := os.Create(outPath)
				if err == nil {
					jpeg.Encode(outFile, watermarked, nil)
					outFile.Close()
				}
			}
		}()
	}

	wgEncode.Wait()
	return time.Since(start)
}
