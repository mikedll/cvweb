
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

	forWindow := pkg.FindNeedle(os.Args[1], os.Args[2])	
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
