package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

func generateHTMLFiles(dir string, count int, r *rand.Rand) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tags := []string{"div", "p", "a", "span", "h1", "h2", "ul", "li", "strong", "em"}
	words := []string{"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit"}

	for i := 0; i < count; i++ {
		fileName := filepath.Join(dir, fmt.Sprintf("doc_%d.html", i))
		var sb strings.Builder

		sb.WriteString("<!DOCTYPE html>\n<html>\n<body>\n")
		
		numTags := r.Intn(450) + 50 
		for j := 0; j < numTags; j++ {
			tag := tags[r.Intn(len(tags))]
			content := words[r.Intn(len(words))]
			
			if r.Float32() < 0.20 {
				innerTag := tags[r.Intn(len(tags))]
				sb.WriteString(fmt.Sprintf("<%s><%s>%s</%s></%s>\n", tag, innerTag, content, innerTag, tag))
			} else {
				sb.WriteString(fmt.Sprintf("<%s>%s</%s>\n", tag, content, tag))
			}
		}
		
		sb.WriteString("</body>\n</html>")

		if err := os.WriteFile(fileName, []byte(sb.String()), 0644); err != nil {
			return err
		}
	}
	return nil
}

func GenerateLargeArray(size int, r *rand.Rand) []float64 {
	arr := make([]float64, size)
	for i := 0; i < size; i++ {
		arr[i] = r.ExpFloat64() * 100.0 
	}
	return arr
}

func GenerateLargeMatrix(rows, cols int, r *rand.Rand) [][]float64 {
	matrix := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		matrix[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			matrix[i][j] = r.Float64() * 10.0
		}
	}
	return matrix
}
