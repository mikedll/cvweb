
package main

import (
	"os"
	"fmt"
	"pkg"
	"gocv.io/x/gocv"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: ./cli haystack.png needle.png\n")
		return
	}

	findResult := pkg.FindNeedle(os.Args[1], os.Args[2])	
	defer findResult.Mat.Close()
	
	window := gocv.NewWindow("Needle in Haystack")
	for {
		if findResult.Mat.Empty() {
			fmt.Printf("Empty mat, exiting\n")
			break
		}

		window.ResizeWindow(findResult.Mat.Cols(), findResult.Mat.Rows())
		window.IMShow(findResult.Mat)
		window.WaitKey(1)
	}
}
