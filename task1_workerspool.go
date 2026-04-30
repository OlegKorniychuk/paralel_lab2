package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

func countTagsWorkersPool(dir string) (map[string]int, time.Duration) {
	start := time.Now()

	files, err := filepath.Glob(filepath.Join(dir, "*.html"))
	if err != nil || len(files) == 0 {
		fmt.Println("Error: directory does not exist or no files found")
		return nil, 0
	}

	numWorkers := runtime.NumCPU()
	jobs := make(chan string, len(files))
	results := make(chan map[string]int, len(files))

	re := regexp.MustCompile(`<([a-zA-Z0-9]+)>`)

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range jobs {
				counts := make(map[string]int)
				data, err := os.ReadFile(file)
				if err == nil {
					matches := re.FindAllStringSubmatch(string(data), -1)
					for _, match := range matches {
						if len(match) > 1 {
							tagName := strings.ToLower(match[1])
							counts[tagName]++
						}
					}
				}
				results <- counts
			}
		}()
	}

	for _, file := range files {
		jobs <- file
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	finalCounts := make(map[string]int)
	for counts := range results {
		for k, v := range counts {
			finalCounts[k] += v
		}
	}

	return finalCounts, time.Since(start)
}

func calcStatsWorkersPool(arr []float64) (min, max, mean, median float64, duration time.Duration) {
	start := time.Now()

	if len(arr) == 0 {
		return 0, 0, 0, 0, time.Since(start)
	}

	type partialStats struct {
		min, max, sum float64
	}

	numChunks := 16
	if len(arr) < numChunks {
		numChunks = 1
	}
	chunkSize := (len(arr) + numChunks - 1) / numChunks

	numWorkers := runtime.NumCPU()
	jobs := make(chan []float64, numChunks)
	results := make(chan partialStats, numChunks)

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chunk := range jobs {
				pMin, pMax := chunk[0], chunk[0]
				pSum := 0.0
				for _, v := range chunk {
					if v < pMin {
						pMin = v
					}
					if v > pMax {
						pMax = v
					}
					pSum += v
				}
				results <- partialStats{pMin, pMax, pSum}
			}
		}()
	}

	sentChunks := 0
	for i := 0; i < numChunks; i++ {
		startIdx := i * chunkSize
		if startIdx >= len(arr) {
			break
		}
		endIdx := startIdx + chunkSize
		if endIdx > len(arr) {
			endIdx = len(arr)
		}
		jobs <- arr[startIdx:endIdx]
		sentChunks++
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	totalSum := 0.0
	min, max = arr[0], arr[0]
	initialized := false
	for p := range results {
		if !initialized {
			min, max = p.min, p.max
			initialized = true
		} else {
			if p.min < min {
				min = p.min
			}
			if p.max > max {
				max = p.max
			}
		}
		totalSum += p.sum
	}
	mean = totalSum / float64(len(arr))

	sortedArr := make([]float64, len(arr))
	copy(sortedArr, arr)
	sort.Float64s(sortedArr)
	n := len(sortedArr)
	if n%2 == 0 {
		median = (sortedArr[n/2-1] + sortedArr[n/2]) / 2.0
	} else {
		median = sortedArr[n/2]
	}

	return min, max, mean, median, time.Since(start)
}

func multiplyMatricesWorkersPool(A, B [][]float64) ([][]float64, time.Duration) {
	start := time.Now()

	rowsA := len(A)
	colsA := len(A[0])
	colsB := len(B[0])

	type resultRow struct {
		idx int
		row []float64
	}

	numWorkers := runtime.NumCPU()
	jobs := make(chan int, rowsA)
	results := make(chan resultRow, rowsA)

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				resRow := make([]float64, colsB)
				for j := 0; j < colsB; j++ {
					sum := 0.0
					for k := 0; k < colsA; k++ {
						sum += A[i][k] * B[k][j]
					}
					resRow[j] = sum
				}
				results <- resultRow{i, resRow}
			}
		}()
	}

	for i := 0; i < rowsA; i++ {
		jobs <- i
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	C := make([][]float64, rowsA)
	for res := range results {
		C[res.idx] = res.row
	}

	return C, time.Since(start)
}
