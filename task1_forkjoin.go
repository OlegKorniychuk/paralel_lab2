package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

func countTagsForkJoin(dir string) (map[string]int, time.Duration) {
	start := time.Now()

	files, err := filepath.Glob(filepath.Join(dir, "*.html"))
	if err != nil || len(files) == 0 {
		fmt.Println("Error: directory does not exist or no files found")
		return nil, 0
	}

	re := regexp.MustCompile(`<([a-zA-Z0-9]+)>`)

	var forkJoinCount func([]string) map[string]int
	forkJoinCount = func(f []string) map[string]int {
		if len(f) <= 2 {
			counts := make(map[string]int)
			for _, file := range f {
				data, err := os.ReadFile(file)
				if err != nil {
					continue
				}
				matches := re.FindAllStringSubmatch(string(data), -1)
				for _, match := range matches {
					if len(match) > 1 {
						tagName := strings.ToLower(match[1])
						counts[tagName]++
					}
				}
			}
			return counts
		}

		mid := len(f) / 2
		var leftCounts, rightCounts map[string]int
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			leftCounts = forkJoinCount(f[:mid])
		}()
		go func() {
			defer wg.Done()
			rightCounts = forkJoinCount(f[mid:])
		}()

		wg.Wait()

		// Merge
		for k, v := range rightCounts {
			leftCounts[k] += v
		}
		return leftCounts
	}

	res := forkJoinCount(files)
	return res, time.Since(start)
}

func calcStatsForkJoin(arr []float64) (min, max, mean, median float64, duration time.Duration) {
	start := time.Now()

	if len(arr) == 0 {
		return 0, 0, 0, 0, time.Since(start)
	}

	type partialStats struct {
		min, max, sum float64
	}

	var forkJoinStats func([]float64) partialStats
	forkJoinStats = func(a []float64) partialStats {
		if len(a) <= 1000 {
			pMin, pMax := a[0], a[0]
			pSum := 0.0
			for _, v := range a {
				if v < pMin {
					pMin = v
				}
				if v > pMax {
					pMax = v
				}
				pSum += v
			}
			return partialStats{pMin, pMax, pSum}
		}

		mid := len(a) / 2
		var left, right partialStats
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			left = forkJoinStats(a[:mid])
		}()
		go func() {
			defer wg.Done()
			right = forkJoinStats(a[mid:])
		}()

		wg.Wait()

		res := partialStats{
			min: left.min,
			max: left.max,
			sum: left.sum + right.sum,
		}
		if right.min < res.min {
			res.min = right.min
		}
		if right.max > res.max {
			res.max = right.max
		}
		return res
	}

	stats := forkJoinStats(arr)
	min = stats.min
	max = stats.max
	mean = stats.sum / float64(len(arr))

	// Median still needs sort
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

func multiplyMatricesForkJoin(A, B [][]float64) ([][]float64, time.Duration) {
	start := time.Now()

	rowsA := len(A)
	colsA := len(A[0])
	colsB := len(B[0])

	C := make([][]float64, rowsA)
	for i := range C {
		C[i] = make([]float64, colsB)
	}

	var forkJoinMult func(int, int)
	forkJoinMult = func(rowStart, rowEnd int) {
		if rowEnd-rowStart <= 10 {
			for i := rowStart; i < rowEnd; i++ {
				for j := 0; j < colsB; j++ {
					sum := 0.0
					for k := 0; k < colsA; k++ {
						sum += A[i][k] * B[k][j]
					}
					C[i][j] = sum
				}
			}
			return
		}

		mid := (rowStart + rowEnd) / 2
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			forkJoinMult(rowStart, mid)
		}()
		go func() {
			defer wg.Done()
			forkJoinMult(mid, rowEnd)
		}()
		wg.Wait()
	}

	forkJoinMult(0, rowsA)

	return C, time.Since(start)
}
