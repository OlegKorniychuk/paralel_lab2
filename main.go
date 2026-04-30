package main

import (
	"math/rand"
	"time"
)

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	generateHTMLFiles("data_html", 20, r)
}