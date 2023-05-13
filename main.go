// What it does:
//
// This example draws two examples, an atom and a rook, based on:
// https://docs.opencv.org/2.4/doc/tutorials/core/basic_geometric_drawing/basic_geometric_drawing.html.
//
// How to run:
//
// 		go run ./cmd/basic-drawing/main.go
//

package main

import (
	"os"
	"fmt"
	"image/color"
	"gocv.io/x/gocv"
)

var w = 400

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: ./cli haystack.png needle.png\n")
		return
	}

	needleImg := gocv.IMRead(os.Args[2], gocv.IMReadColor)
	defer needleImg.Close()

	hayStackImg := gocv.IMRead(os.Args[1], gocv.IMReadColor)
	defer hayStackImg.Close()
	
	window := gocv.NewWindow("Needle in Haystack")

	sift := gocv.NewSIFT()
	defer sift.Close()

	needleKp, needleDesc := sift.DetectAndCompute(needleImg, gocv.NewMat())
	hayStackKp, hayStackDesc := sift.DetectAndCompute(hayStackImg, gocv.NewMat())

	fmt.Printf("Haystack rows=%d, cols=%d\n", hayStackImg.Rows(), hayStackImg.Cols())
	
	flannMatcher := gocv.NewFlannBasedMatcher()
	defer flannMatcher.Close()

	dontUnderstand := 2
	matches := flannMatcher.KnnMatch(needleDesc, hayStackDesc, dontUnderstand)
	fmt.Printf("Here we go: %p, number of matches is %d\n", matches, len(matches))

	var good []gocv.DMatch
	for _, m := range matches {
		if len(m) > 1 {
			if m[0].Distance < 0.70 * m[1].Distance {
				// fmt.Printf("Appending one for %d\n", i)
				good = append(good, m[0])
			}
		}
	}

	out := gocv.NewMat()
	defer out.Close()

	// matches color
	c1 := color.RGBA{
		R: 0,
		G: 255,
		B: 0,
		A: 0,
	}

	// point color
	c2 := color.RGBA{
		R: 255,
		G: 0,
		B: 0,
		A: 0,
	}
	
	mask := make([]byte, 0)
	gocv.DrawMatches(needleImg, needleKp, hayStackImg, hayStackKp, good, &out, c1, c2, mask, gocv.DrawDefault)
	
	forWindow := out.Clone()
	defer forWindow.Close()
	
	for {
		if forWindow.Empty() {
			fmt.Printf("Empty mat, exiting\n")
			break
		}

		window.ResizeWindow(forWindow.Cols(), forWindow.Rows())
		window.IMShow(forWindow)
		window.WaitKey(1)
	}
}
