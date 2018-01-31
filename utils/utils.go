package utils

import "gocv.io/x/gocv"

func ShowImage(image gocv.Mat) {
	window := gocv.NewWindow("Hello")
	for {
		window.IMShow(image)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}
