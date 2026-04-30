package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"path/filepath"
	"time"
)

func drawLab2(img draw.Image, x, y int, c color.Color) {
	font := map[rune][]uint8{
		'l': {0b10000000, 0b10000000, 0b10000000, 0b10000000, 0b10000000, 0b10000000, 0b10000000},
		'a': {0b01110000, 0b00001000, 0b01111000, 0b10001000, 0b01111000, 0b00000000, 0b00000000},
		'b': {0b10000000, 0b10000000, 0b11110000, 0b10001000, 0b11110000, 0b00000000, 0b00000000},
		'2': {0b11111000, 0b00001000, 0b01111000, 0b10000000, 0b11111000, 0b00000000, 0b00000000},
	}
	chars := []rune{'l', 'a', 'b', '2'}
	scale := 3
	spacing := 2 * scale
	currentX := x
	for _, char := range chars {
		mask := font[char]
		for row, b := range mask {
			for col := 0; col < 8; col++ {
				if (b << uint(col)) & 0b10000000 != 0 {
					for sy := 0; sy < scale; sy++ {
						for sx := 0; sx < scale; sx++ {
							img.Set(currentX+col*scale+sx, y+row*scale+sy, c)
						}
					}
				}
			}
		}
		currentX += 6*scale + spacing
	}
}

func processImageSequential(inputPath, outputPath string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	bounds := img.Bounds()
	grayImg := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			grayImg.Set(x, y, color.GrayModel.Convert(img.At(x, y)))
		}
	}

	watermarked := image.NewRGBA(bounds)
	draw.Draw(watermarked, bounds, grayImg, bounds.Min, draw.Src)
	drawLab2(watermarked, bounds.Max.X-100, bounds.Max.Y-40, color.RGBA{255, 0, 0, 255})

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	return jpeg.Encode(outFile, watermarked, nil)
}

func processImagesSequential(inputDir, outputDir string) time.Duration {
	start := time.Now()
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.MkdirAll(outputDir, 0755)
	}
	files, err := filepath.Glob(filepath.Join(inputDir, "*.jpg"))
	if err != nil {
		return 0
	}
	for _, file := range files {
		outputPath := filepath.Join(outputDir, filepath.Base(file))
		processImageSequential(file, outputPath)
	}
	return time.Since(start)
}
