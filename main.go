package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"reflect"
	"time"
)

type TaskResults struct {
	TagCounts map[string]int
	Min       float64
	Max       float64
	Mean      float64
	Median    float64
	Matrix    [][]float64
	Durations struct {
		Tags     time.Duration
		Stats    time.Duration
		Multiply time.Duration
		Total    time.Duration
	}
}

func runSequential(htmlDir string, array []float64, A, B [][]float64) TaskResults {
	var res TaskResults
	start := time.Now()
	res.TagCounts, res.Durations.Tags = countTagsSequential(htmlDir)
	res.Min, res.Max, res.Mean, res.Median, res.Durations.Stats = calcStatsSequential(array)
	res.Matrix, res.Durations.Multiply = multiplyMatricesSequential(A, B)
	res.Durations.Total = time.Since(start)
	return res
}

func runForkJoin(htmlDir string, array []float64, A, B [][]float64) TaskResults {
	var res TaskResults
	start := time.Now()
	res.TagCounts, res.Durations.Tags = countTagsForkJoin(htmlDir)
	res.Min, res.Max, res.Mean, res.Median, res.Durations.Stats = calcStatsForkJoin(array)
	res.Matrix, res.Durations.Multiply = multiplyMatricesForkJoin(A, B)
	res.Durations.Total = time.Since(start)
	return res
}

func runMapReduce(htmlDir string, array []float64, A, B [][]float64) TaskResults {
	var res TaskResults
	start := time.Now()
	res.TagCounts, res.Durations.Tags = countTagsMapReduce(htmlDir)
	res.Min, res.Max, res.Mean, res.Median, res.Durations.Stats = calcStatsMapReduce(array)
	res.Matrix, res.Durations.Multiply = multiplyMatricesMapReduce(A, B)
	res.Durations.Total = time.Since(start)
	return res
}

func runWorkersPool(htmlDir string, array []float64, A, B [][]float64) TaskResults {
	var res TaskResults
	start := time.Now()
	res.TagCounts, res.Durations.Tags = countTagsWorkersPool(htmlDir)
	res.Min, res.Max, res.Mean, res.Median, res.Durations.Stats = calcStatsWorkersPool(array)
	res.Matrix, res.Durations.Multiply = multiplyMatricesWorkersPool(A, B)
	res.Durations.Total = time.Since(start)
	return res
}

func compareResults(seq, par TaskResults, name string) {
	fmt.Printf("\n--- Verification for %s ---\n", name)

	// Tags
	if reflect.DeepEqual(seq.TagCounts, par.TagCounts) {
		fmt.Println("[OK] Tag counts match")
	} else {
		fmt.Println("[FAIL] Tag counts do not match")
	}

	// Stats
	eps := 1e-9
	statsMatch := math.Abs(seq.Min-par.Min) < eps &&
		math.Abs(seq.Max-par.Max) < eps &&
		math.Abs(seq.Mean-par.Mean) < eps &&
		math.Abs(seq.Median-par.Median) < eps

	if statsMatch {
		fmt.Println("[OK] Stats match")
	} else {
		fmt.Printf("[FAIL] Stats mismatch: Seq(%.4f, %.4f, %.4f, %.4f) vs Par(%.4f, %.4f, %.4f, %.4f)\n",
			seq.Min, seq.Max, seq.Mean, seq.Median, par.Min, par.Max, par.Mean, par.Median)
	}

	// Matrix
	matrixMatch := true
	if len(seq.Matrix) != len(par.Matrix) || len(seq.Matrix[0]) != len(par.Matrix[0]) {
		matrixMatch = false
	} else {
		for i := range seq.Matrix {
			for j := range seq.Matrix[i] {
				if math.Abs(seq.Matrix[i][j]-par.Matrix[i][j]) > eps {
					matrixMatch = false
					break
				}
			}
			if !matrixMatch {
				break
			}
		}
	}

	if matrixMatch {
		fmt.Println("[OK] Matrix results match")
	} else {
		fmt.Println("[FAIL] Matrix results mismatch")
	}
}

func printTimings(res TaskResults, name string) {
	fmt.Printf("\n=== %s Timings ===\n", name)
	fmt.Printf("Tags Parsing: %v\n", res.Durations.Tags)
	fmt.Printf("Stats Calc:   %v\n", res.Durations.Stats)
	fmt.Printf("Matrix Mult:  %v\n", res.Durations.Multiply)
	fmt.Printf("TOTAL TASK:   %v\n", res.Durations.Total)
}

func main() {
	htmlDir := "data_html"
	matrixRows := 500
	matrixColumns := 500
	arrayLength := 100000
	filesCount := 1000

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	fmt.Printf("Generating Data...\n")
	if err := generateHTMLFiles(htmlDir, filesCount, r); err != nil {
		fmt.Printf("HTML generation error: %v\n", err)
	}
	defer os.RemoveAll(htmlDir)

	matrix1 := GenerateLargeMatrix(matrixRows, matrixColumns, r)
	matrix2 := GenerateLargeMatrix(matrixRows, matrixColumns, r)
	array := GenerateLargeArray(arrayLength, r)

	fmt.Println("Running Sequential...")
	seqRes := runSequential(htmlDir, array, matrix1, matrix2)
	printTimings(seqRes, "Sequential")

	fmt.Println("\nRunning Fork-Join...")
	fjRes := runForkJoin(htmlDir, array, matrix1, matrix2)
	printTimings(fjRes, "Fork-Join")
	compareResults(seqRes, fjRes, "Fork-Join")

	fmt.Println("\nRunning Map-Reduce...")
	mrRes := runMapReduce(htmlDir, array, matrix1, matrix2)
	printTimings(mrRes, "Map-Reduce")
	compareResults(seqRes, mrRes, "Map-Reduce")

	fmt.Println("\nRunning Workers Pool...")
	wpRes := runWorkersPool(htmlDir, array, matrix1, matrix2)
	printTimings(wpRes, "Workers Pool")
	compareResults(seqRes, wpRes, "Workers Pool")
}