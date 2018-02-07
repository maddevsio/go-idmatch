package preprocessing

import (
	"image"
	"math"

	"gocv.io/x/gocv"
)

func rotate(edged, original gocv.Mat) gocv.Mat {
	var positive, negative float64
	var posCount, negCount int
	lines := gocv.NewMat()
	defer lines.Close()

	gocv.HoughLinesP(edged, lines, 1, math.Pi/180, 200)
	for row := 0; row < lines.Rows(); row++ {
		x1, y1, x2, y2 := lines.GetIntAt(row, 0), lines.GetIntAt(row, 1), lines.GetIntAt(row, 2), lines.GetIntAt(row, 3)
		if distance := math.Sqrt(math.Pow(float64(x2-x1), 2) + math.Pow(float64(y2-y1), 2)); distance > 20 {
			theta := math.Atan2(float64(y2-y1), float64(x2-x1))
			if math.Abs(theta) == math.Pi/2 {
				continue
			}
			theta *= 180 / math.Pi
			if theta > 0 {
				positive += theta
				posCount++
			} else {
				negative += theta
				negCount++
			}
		}
	}
	if posCount > 0 {
		positive /= float64(posCount)
	}
	if negCount > 0 {
		negative /= float64(negCount)
	}

	if math.Abs(positive) == math.Abs(negative) {
		return gocv.GetRotationMatrix2D(image.Point{edged.Cols() / 2, edged.Rows() / 2}, 0, 1)
	} else if math.Abs(positive) < math.Abs(negative) {
		return gocv.GetRotationMatrix2D(image.Point{edged.Cols() / 2, edged.Rows() / 2}, positive, 1)
	}
	return gocv.GetRotationMatrix2D(image.Point{edged.Cols() / 2, edged.Rows() / 2}, negative, 1)
}

// Contours takes image file path and crops it by contour
func Contours(file string) gocv.Mat {
	var maxArea int
	var rect image.Rectangle

	original := gocv.NewMat()
	defer original.Close()

	img := gocv.IMRead(file, gocv.IMReadColor)
	img.CopyTo(original)
	// gocv.ApplyColorMap(img, img, gocv.ColormapHot)
	gocv.CvtColor(img, img, gocv.ColorRGBToGray)
	gocv.MedianBlur(img, img, 7)
	gocv.Canny(img, img, 20, 170)

	// kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{1, 10})
	// gocv.MorphologyEx(img, img, gocv.MorphClose, kernel)

	rotation := rotate(img, original)
	gocv.WarpAffine(img, img, rotation, image.Point{img.Cols(), img.Rows()})
	gocv.WarpAffine(original, original, rotation, image.Point{img.Cols(), img.Rows()})

	contours := gocv.FindContours(img, gocv.RetrievalList, gocv.ChainApproxNone)
	for _, v := range contours {
		r := gocv.BoundingRect(v)
		if contourArea := r.Dx() * r.Dy(); contourArea > maxArea {
			rect = r
			maxArea = contourArea
		}
	}

	return original.Region(rect)
}
