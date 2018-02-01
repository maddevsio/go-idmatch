package preprocessing

import (
	"image"

	"gocv.io/x/gocv"
)

func Contours(file string) gocv.Mat {
	var maxArea int
	var rect image.Rectangle
	edged := gocv.NewMat()
	original := gocv.NewMat()
	defer edged.Close()
	defer original.Close()

	img := gocv.IMRead(file, gocv.IMReadColor)
	img.CopyTo(original)
	gocv.CvtColor(img, img, gocv.ColorBGRToGray)
	gocv.MedianBlur(img, img, 7)
	gocv.Canny(img, edged, 20, 180)
	contours := gocv.FindContours(edged, gocv.RetrievalList, gocv.ChainApproxSimple)
	for _, v := range contours {
		r := gocv.BoundingRect(v)
		if contourArea := r.Dx() * r.Dy(); contourArea > maxArea {
			rect = r
			maxArea = contourArea
		}
	}
	// gocv.Rectangle(edged, rect, color.RGBA{255, 0, 0, 255}, 2)
	// utils.ShowImage(edged)
	return original.Region(rect)
}
