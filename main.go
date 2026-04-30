package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	htmlDir := "data_html"
	matrixRows := 1000
	matrixColumns := 1000
	arrayLength := 100000
	filesCount := 50000

	r := rand.New(rand.NewSource(time.Now().UnixNano()));

	fmt.Printf("Generating HTML data...\n")
	if err := generateHTMLFiles(htmlDir, filesCount, r); err != nil {
		fmt.Printf("HTML generation error: %v\n", err)
	}
	defer func() {
		fmt.Printf("Cleaning up: deleting %s...\n", htmlDir)
		os.RemoveAll(htmlDir)
	}()

	matrix1 := GenerateLargeMatrix(matrixRows, matrixColumns, r)
	matrix2 := GenerateLargeMatrix(matrixRows, matrixColumns, r)
	array := GenerateLargeArray(arrayLength, r)

	tagCounts, timeTags := countTagsSequential(htmlDir)
	fmt.Printf("\n=== Html tags counts ===\n")
	fmt.Printf("Execution time (%d files): %v\n", filesCount, timeTags)
	fmt.Printf("Unique tags found: %d\n", len(tagCounts))
	fmt.Println(tagCounts)

	min, max, mean, median, timeArray := calcStatsSequential(array)
	fmt.Print("\n=== Array stats calculation ===\n")
	fmt.Printf("Execution time (length %d): %v\n", arrayLength, timeArray)
	fmt.Printf("Min: %.4f | Max: %.4f | Mean: %.4f | Median: %.4f\n", min, max, mean, median)

	_, timeMultiply := multiplyMatricesSequential(matrix1, matrix2)
	fmt.Printf("\n=== Matrices multiplication ===\n")
	fmt.Printf("Execution time (%dx%d): %v\n", matrixRows, matrixColumns, timeMultiply)
}