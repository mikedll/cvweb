package pkg

import (
	"fmt"
	"log"
	"math"
	"image"
	"image/color"
	"gocv.io/x/gocv"
)

const originEpsilon = 1
const matchDistanceFactor = 0.70
const expectedOriginRatio = 0.2

type FindResult struct {
	Found bool `json:"found"`
	Mat gocv.Mat
}

//
// Returns pointer to 4 element int array that describes a rectangle { x0, y0, x1, y1 }.
// x0 and y0 comprise the rectangle's origin, drawn from the top left of some image. So x0 = 10 means
// 10 points right of the left side of the coordinate space. y0 = 15 means 15 points below the top
// of the coordinate space.
//
func calcOrigin(good []gocv.DMatch, haystackKps []gocv.KeyPoint, needleKps []gocv.KeyPoint, needleImg gocv.Mat) *[]int {
	// capture number of origins in the training image implied by the matches
	var origins [][]float64
	originCount := make(map[int]int)
	for _, dMatch := range good {
		needleKp := needleKps[dMatch.QueryIdx]
		trainKp := haystackKps[dMatch.TrainIdx]
		trainOrigin := []float64{ trainKp.X - needleKp.X, trainKp.Y - needleKp.Y }

		originIdx := -1
		recognized := false
		for i, origin := range origins {
			if math.Abs(trainOrigin[0] - origin[0]) < originEpsilon && math.Abs(trainOrigin[1] - origin[1]) < originEpsilon {
				recognized = true
				originIdx = i
			}
		}
		
		if !recognized {
			origins = append(origins, []float64{ trainOrigin[0], trainOrigin[1] } )
			originIdx = len(origins) - 1
			if _, ok := originCount[originIdx]; !ok {
				originCount[originIdx] = 0
			}
		}

		if originIdx == -1 {
			log.Fatalf("logic error: bad originIdx\n")
		}

		originCount[originIdx] += 1		
	}

	// If there is at least one origin, and there aren't too many origins, pick the most popular one
	foundOrigin := -1
	if len(origins) >= 1 && (expectedOriginRatio * float64(len(good))) > float64(len(origins)) {
		foundOrigin = 0
		for originIdx, count := range originCount {
			if Debug {
				fmt.Printf("idx=%d, count=%d\n", originIdx, count)
			}
			if count > originCount[foundOrigin] {
				foundOrigin = originIdx
			}
		}
	}

	if foundOrigin != -1 {
		if Debug {
			fmt.Printf("There is a reasonably unique origin among %d origins\n", len(origins))
		}
		retOrigin := []int{
			int(math.Round(origins[foundOrigin][0])),
			int(math.Round(origins[foundOrigin][1])),
			needleImg.Cols(),
			needleImg.Rows(),
		}
		return &retOrigin
	} else {
		return nil
	}
}

//
// Returns a gocv.Mat
// 
// Caller should call Close on it.
// 
func matchRender(needleImg gocv.Mat, needleKps []gocv.KeyPoint, hayStackImg gocv.Mat, haystackKps []gocv.KeyPoint,
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
	gocv.DrawMatches(needleImg, needleKps, hayStackImg, haystackKps, good, &out, c1, c2, mask, gocv.DrawDefault)

	return out
}

//
// Returns a FindResult. The caller must call Close on the result.Mat.
//
// The Mat is an image of the haystack image, with the needle, if found, drawn on top of it.
//
func FindNeedle(haystackFile, needleFile string) FindResult {	
	hayStackImg := gocv.IMRead(haystackFile, gocv.IMReadColor)
	defer hayStackImg.Close()

	needleImg := gocv.IMRead(needleFile, gocv.IMReadColor)
	defer needleImg.Close()

	sift := gocv.NewSIFT()
	defer sift.Close()

	needleMask := gocv.NewMat()
	defer needleMask.Close()
	needleKps, needleDesc := sift.DetectAndCompute(needleImg, needleMask)
	defer needleDesc.Close()

	haystackMask := gocv.NewMat()
	defer haystackMask.Close()
	haystackKps, haystackDesc := sift.DetectAndCompute(hayStackImg, haystackMask)
	defer haystackDesc.Close()

	if Debug {
		fmt.Printf("Haystack cols=%d, rows=%d\n", hayStackImg.Cols(), hayStackImg.Rows())
		fmt.Printf("Needle cols=%d, rows=%d\n", needleImg.Cols(), needleImg.Rows())
		for _, keyPoint := range needleKps {
			fmt.Printf("Needle key point at (%.2f, %.2f)\n", keyPoint.X, keyPoint.Y)
		}
	}
		
	flannMatcher := gocv.NewFlannBasedMatcher()
	defer flannMatcher.Close()

	dontUnderstand := 2
	// Needle is the query, haystack is the train
	matches := flannMatcher.KnnMatch(needleDesc, haystackDesc, dontUnderstand)
	if Debug {
		fmt.Printf("Here we go: %p, number of matches is %d\n", matches, len(matches))
	}

	// dunno what this loop is doing. I know without it, we get too many bad matches.
	var good []gocv.DMatch
	for _, m := range matches {
		if len(m) > 1 {
			needleKp := needleKps[m[0].QueryIdx]
			trainKp := haystackKps[m[0].TrainIdx]
			if m[0].Distance < matchDistanceFactor * m[1].Distance {
				if Debug {
					fmt.Printf("Hopefully a query key point (%.2f %.2f), train key point (%.2f, %.2f), and two distances: %.2f, %.2f, and image index of %d\n",
						needleKp.X, needleKp.Y, trainKp.X, trainKp.Y, m[0].Distance, m[1].Distance, m[0].ImgIdx)
				}
				good = append(good, m[0])
			} else {
				if Debug {
					fmt.Printf("Bad query key point (%.2f %.2f), and two distances: %.2f, %.2f\n", needleKp.X, needleKp.Y, m[0].Distance, m[1].Distance)
				}
			}
		}
	}

	// This isn't being used at the moment
	out := matchRender(needleImg, needleKps, hayStackImg, haystackKps, good)
	defer out.Close()

	forWindow := hayStackImg.Clone()	

	origin := calcOrigin(good, haystackKps, needleKps, needleImg)
	if origin != nil {
		if Debug {
			fmt.Printf("Origin in training image: (%d, %d, %d, %d)\n", (*origin)[0], (*origin)[1], (*origin)[2], (*origin)[3])
		}
		blue := color.RGBA{0, 0, 255, 0}
		gocv.Rectangle(&forWindow, image.Rect((*origin)[0], (*origin)[1], (*origin)[0] + (*origin)[2], (*origin)[1] + (*origin)[3]), blue, 2)		
	}

	return FindResult{
		Found: origin != nil,
		Mat: forWindow,
	}
}

