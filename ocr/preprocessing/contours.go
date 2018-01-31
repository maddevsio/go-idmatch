package preprocessing

import (
	"image"

	"gocv.io/x/gocv"
)

func Contours(file string) gocv.Mat {
	var maxArea float64
	var rect image.Rectangle
	edged := gocv.NewMat()
	original := gocv.NewMat()

	img := gocv.IMRead(file, gocv.IMReadColor)
	img.CopyTo(original)
	gocv.CvtColor(img, img, gocv.ColorBGRToGray)
	gocv.MedianBlur(img, img, 7)
	gocv.Canny(img, edged, 20, 180)
	contours := gocv.FindContours(edged, gocv.RetrievalList, gocv.ChainApproxSimple)
	for _, v := range contours {
		contourArea := gocv.ContourArea(v)
		rect = gocv.BoundingRect(v)
		if contourArea > maxArea {
			maxArea = contourArea
		}
	}
	// gocv.Rectangle(original, rect, color.RGBA{255, 0, 0, 255}, 2)
	return original.Region(rect)
}
