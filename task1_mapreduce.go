package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

func countTagsMapReduce(dir string) (map[string]int, time.Duration) {
	start := time.Now()

	files, err := filepath.Glob(filepath.Join(dir, "*.html"))
	if err != nil || len(files) == 0 {
		fmt.Println("Error: directory does not exist or no files found")
		return nil, 0
	}

	re := regexp.MustCompile(`<([a-zA-Z0-9]+)>`)
	
	mapChan := make(chan map[string]int, len(files))

	// Map phase
	for _, file := range files {
		go func(f string) {
			counts := make(map[string]int)
			data, err := os.ReadFile(f)
			if err == nil {
				matches := re.FindAllStringSubmatch(string(data), -1)
				for _, match := range matches {
					if len(match) > 1 {
						tagName := strings.ToLower(match[1])
						counts[tagName]++
					}
				}
			}
			mapChan <- counts
		}(file)
	}

	// Reduce phase
	finalCounts := make(map[string]int)
	for i := 0; i < len(files); i++ {
		counts := <-mapChan
		for k, v := range counts {
			finalCounts[k] += v
		}
	}

	return finalCounts, time.Since(start)
}

func calcStatsMapReduce(arr []float64) (min, max, mean, median float64, duration time.Duration) {
	start := time.Now()

	if len(arr) == 0 {
		return 0, 0, 0, 0, time.Since(start)
	}

	type partialStats struct {
		min, max, sum float64
		count         int
	}

	numChunks := 8
	if len(arr) < numChunks {
		numChunks = 1
	}
	chunkSize := (len(arr) + numChunks - 1) / numChunks
	mapChan := make(chan partialStats, numChunks)

	// Map phase
	for i := 0; i < numChunks; i++ {
		startIdx := i * chunkSize
		if startIdx >= len(arr) {
			numChunks = i // Adjust if we have fewer chunks than planned
			break
		}
		endIdx := startIdx + chunkSize
		if endIdx > len(arr) {
			endIdx = len(arr)
		}

		go func(chunk []float64) {
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
			mapChan <- partialStats{pMin, pMax, pSum, len(chunk)}
		}(arr[startIdx:endIdx])
	}

	// Reduce phase
	totalSum := 0.0
	min, max = arr[0], arr[0]
	initialized := false
	for i := 0; i < numChunks; i++ {
		p := <-mapChan
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

func multiplyMatricesMapReduce(A, B [][]float64) ([][]float64, time.Duration) {
	start := time.Now()

	rowsA := len(A)
	colsA := len(A[0])
	colsB := len(B[0])

	type resultRow struct {
		idx int
		row []float64
	}

	mapChan := make(chan resultRow, rowsA)

	// Map phase
	for i := 0; i < rowsA; i++ {
		go func(rowIdx int) {
			resRow := make([]float64, colsB)
			for j := 0; j < colsB; j++ {
				sum := 0.0
				for k := 0; k < colsA; k++ {
					sum += A[rowIdx][k] * B[k][j]
				}
				resRow[j] = sum
			}
			mapChan <- resultRow{rowIdx, resRow}
		}(i)
	}

	// Reduce phase
	C := make([][]float64, rowsA)
	for i := 0; i < rowsA; i++ {
		res := <-mapChan
		C[res.idx] = res.row
	}

	return C, time.Since(start)
}
