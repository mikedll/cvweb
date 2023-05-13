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
	"math"
	"image/color"
	"gocv.io/x/gocv"
)

var w = 400
const originEpsilon = 1
const matchDistanceFactor = 0.70
const expectedOriginRatio = 0.1

func calcOrigin(good []gocv.DMatch, hayStackKps []gocv.KeyPoint, needleKps []gocv.KeyPoint, needleImg gocv.Mat) *[]float64 {
	// capture number of origins in the training image implied by the matches
	var origins [][]float64
	originCount := make(map[int]int)
	for _, dMatch := range good {
		needleKp := needleKps[dMatch.QueryIdx]
		trainKp := hayStackKps[dMatch.TrainIdx]
		trainOrigin := []float64{ trainKp.X - needleKp.X, trainKp.Y - needleKp.Y }

		recognized := false
		for _, origin := range origins {
			if math.Abs(trainOrigin[0] - origin[0]) < originEpsilon && math.Abs(trainOrigin[1] - origin[1]) < originEpsilon {
				recognized = true
			}
		}
		
		if !recognized {
			origins = append(origins, []float64{ trainOrigin[0], trainOrigin[1] } )
			originIdx := len(origins) - 1
			if _, ok := originCount[originIdx]; !ok {
				originCount[originIdx] = 0
			}
			originCount[originIdx] += 1
		}		
	}

	// If there is at least one origin, and there aren't too many origins, pick the most popular one
	foundOrigin := -1
	if len(origins) >= 1 && (expectedOriginRatio * float64(len(good))) > float64(len(origins)) {
		foundOrigin = 0
		for originIdx, count := range originCount {
			if count > originCount[foundOrigin] {
				foundOrigin = originIdx
			}
		}
	}

	if foundOrigin != -1 {
		fmt.Printf("There is a reasonably unique origin among %d origins\n", len(origins))
		retOrigin := []float64{
			origins[foundOrigin][0],
			origins[foundOrigin][1],
			float64(needleImg.Cols()),
			float64(needleImg.Rows()),
		}
		return &retOrigin
	} else {
		return nil
	}
}

//
// Caller should call close on the returned Mat.
// 
func matchRender(needleImg gocv.Mat, needleKps []gocv.KeyPoint, hayStackImg gocv.Mat, hayStackKps []gocv.KeyPoint,
	good []gocv.DMatch) gocv.Mat {
	out := gocv.NewMat()

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
	gocv.DrawMatches(needleImg, needleKps, hayStackImg, hayStackKps, good, &out, c1, c2, mask, gocv.DrawDefault)

	return out
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: ./cli haystack.png needle.png\n")
		return
	}

	needleImg := gocv.IMRead(os.Args[2], gocv.IMReadColor)
	defer needleImg.Close()

	hayStackImg := gocv.IMRead(os.Args[1], gocv.IMReadColor)
	defer hayStackImg.Close()
	
	sift := gocv.NewSIFT()
	defer sift.Close()

	needleKps, needleDesc := sift.DetectAndCompute(needleImg, gocv.NewMat())
	hayStackKps, hayStackDesc := sift.DetectAndCompute(hayStackImg, gocv.NewMat())

	fmt.Printf("Haystack cols=%d, rows=%d\n", hayStackImg.Cols(), hayStackImg.Rows())

	fmt.Printf("Needle cols=%d, rows=%d\n", needleImg.Cols(), needleImg.Rows())
	for _, keyPoint := range needleKps {
		fmt.Printf("Needle key point at (%.2f, %.2f)\n", keyPoint.X, keyPoint.Y)
	}
	
	flannMatcher := gocv.NewFlannBasedMatcher()
	defer flannMatcher.Close()

	dontUnderstand := 2
	// Needle is the query, haystack is the train
	matches := flannMatcher.KnnMatch(needleDesc, hayStackDesc, dontUnderstand)
	fmt.Printf("Here we go: %p, number of matches is %d\n", matches, len(matches))

	// dunno what this loop is doing. I know without it, we get too many bad matches.
	var good []gocv.DMatch
	for _, m := range matches {
		if len(m) > 1 {
			needleKp := needleKps[m[0].QueryIdx]
			trainKp := hayStackKps[m[0].TrainIdx]
			if m[0].Distance < matchDistanceFactor * m[1].Distance {
				fmt.Printf("Hopefully a query key point (%.2f %.2f), train key point (%.2f, %.2f), and two distances: %.2f, %.2f, and image index of %d\n",
					needleKp.X, needleKp.Y, trainKp.X, trainKp.Y, m[0].Distance, m[1].Distance, m[0].ImgIdx)			
				good = append(good, m[0])
			} else {
				fmt.Printf("Bad query key point (%.2f %.2f), and two distances: %.2f, %.2f\n", needleKp.X, needleKp.Y, m[0].Distance, m[1].Distance)
			}
		}
	}

	origin := calcOrigin(good, hayStackKps, needleKps, needleImg)
	if origin != nil {
		fmt.Printf("Origin in training image: (%.2f, %.2f, %.2f, %.2f)\n", (*origin)[0], (*origin)[1], (*origin)[2], (*origin)[3])
	}		
	
	out := matchRender(needleImg, needleKps, hayStackImg, hayStackKps, good)
	defer out.Close()

	forWindow := out.Clone()
	defer forWindow.Close()
	
	window := gocv.NewWindow("Needle in Haystack")
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
