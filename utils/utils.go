package utils

import "gocv.io/x/gocv"

func showImageInternal(image gocv.Mat, winName string) {
	window := gocv.NewWindow(winName)
	defer window.Close()
	for {
		window.ResizeWindow(800, 600)
		window.IMShow(image)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}

func ShowImage(image gocv.Mat) {
	showImageInternal(image, "Hello")
}

func ShowImageInNamedWindow(image gocv.Mat, winName string) {
	showImageInternal(image, winName)
}
