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

	window := gocv.NewWindow("Needle in Haystack")

	forWindow := needleImg.Clone()
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
