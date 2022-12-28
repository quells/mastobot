package main

import (
	"github.com/nfnt/resize"
	"image/jpeg"
	"log"
	"os"
	"time"
)

/*
GOES-17

https://www.star.nesdis.noaa.gov/goes/index.php

256 thumb
1280 full

NOAA/NESDIS/STAR - GOES-West - GeoColor Composite by CIRA/NOAA
*/

func main() {
	file, err := os.Open("goes-west.jpg")
	if err != nil {
		log.Fatal(err)
	}

	img, err := jpeg.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	_ = file.Close()

	start := time.Now()
	m := resize.Resize(256, 0, img, resize.Lanczos3)
	elapsed := time.Since(start)
	log.Println(elapsed)
	_ = m

	out, err := os.Create("goes-west-256.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	_ = jpeg.Encode(out, m, nil)
}
