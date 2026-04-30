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


func countTagsSequential(dir string) (map[string]int, time.Duration) {
	start := time.Now()
	
	files, err := filepath.Glob(filepath.Join(dir, "*.html"))
	if err != nil || len(files) == 0 {
		fmt.Println("Error: directory does not exist or no fiels found")
		return nil, 0
	}

	tagCounts := make(map[string]int)
	

	re := regexp.MustCompile(`<([a-zA-Z0-9]+)>`)

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		matches := re.FindAllStringSubmatch(string(data), -1)
		for _, match := range matches {
			if len(match) > 1 {
				tagName := strings.ToLower(match[1])
				tagCounts[tagName]++
			}
		}
	}

	return tagCounts, time.Since(start)
}


func calcStatsSequential(arr []float64) (min, max, mean, median float64, duration time.Duration) {
	start := time.Now()
	
	if len(arr) == 0 {
		return 0, 0, 0, 0, time.Since(start)
	}

	min, max = arr[0], arr[0]
	sum := 0.0

	for _, v := range arr {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}
	mean = sum / float64(len(arr))

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

func multiplyMatricesSequential(A, B [][]float64) ([][]float64, time.Duration) {
	start := time.Now()
	
	rowsA := len(A)
	colsA := len(A[0])
	colsB := len(B[0])

	C := make([][]float64, rowsA)
	for i := range C {
		C[i] = make([]float64, colsB)
		for j := 0; j < colsB; j++ {
			sum := 0.0
			for k := 0; k < colsA; k++ {
				sum += A[i][k] * B[k][j]
			}
			C[i][j] = sum
		}
	}

	return C, time.Since(start)
}