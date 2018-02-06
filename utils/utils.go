package utils

import "gocv.io/x/gocv"

func ShowImage(image gocv.Mat) {
	window := gocv.NewWindow("Hello")
	defer window.Close()
	for {
		window.ResizeWindow(800, 600)
		window.IMShow(image)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}
