package main

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

func processImagesProdCons(inputDir, outputDir string) time.Duration {
	start := time.Now()
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.MkdirAll(outputDir, 0755)
	}

	files, _ := filepath.Glob(filepath.Join(inputDir, "*.jpg"))
	jobs := make(chan string, len(files))

	// Producer
	go func() {
		for _, f := range files {
			jobs <- f
		}
		close(jobs)
	}()

	// Consumers
	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU()
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				outPath := filepath.Join(outputDir, filepath.Base(path))
				processImageSequential(path, outPath)
			}
		}()
	}

	wg.Wait()
	return time.Since(start)
}
