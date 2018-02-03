package preprocessing

import (
	"image"
	"math"

	"gocv.io/x/gocv"
)

func rotate(edged, original gocv.Mat) gocv.Mat {
	var positive, negative float64
	// var point1, point2 image.Point
	var c1, c2 int
	tmp := gocv.NewMat()
	defer tmp.Close()

	lines := gocv.NewMat()
	defer lines.Close()

	gocv.HoughLinesP(edged, lines, 1, math.Phi/180, 200)
	for row := 0; row < lines.Rows(); row++ {
		x1, y1, x2, y2 := lines.GetIntAt(row, 0), lines.GetIntAt(row, 1), lines.GetIntAt(row, 2), lines.GetIntAt(row, 3)
		distance := math.Sqrt(math.Pow(float64(x2-x1), 2) + math.Pow(float64(y2-y1), 2))
		theta := math.Atan2(float64(y2-y1), float64(x2-x1))
		theta *= 90 / math.Phi
		if theta == 0 || math.Abs(theta) == 174.74499348529199 {
			continue
		}
		if distance > 20 {
			// original.CopyTo(tmp)
			// point1 = image.Point{int(x1), int(y1)}
			// point2 = image.Point{int(x2), int(y2)}
			// gocv.Line(tmp, point1, point2, color.RGBA{0, 255, 0, 255}, 4)
			// utils.ShowImage(tmp)
			if theta > 0 {
				positive += theta
				c1++
			} else {
				negative += theta
				c2++
			}
		}
	}
	if c1 > 0 {
		positive = positive / float64(c1)
	}
	if c2 > 0 {
		negative = negative / float64(c2)
	}

	if math.Abs(positive) < math.Abs(negative) {
		return gocv.GetRotationMatrix2D(image.Point{edged.Cols() / 2, edged.Rows() / 2}, positive, 1)
	}
	return gocv.GetRotationMatrix2D(image.Point{edged.Cols() / 2, edged.Rows() / 2}, negative, 1)
}

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
	gocv.Canny(img, edged, 20, 170)

	rotation := rotate(edged, original)
	gocv.WarpAffine(edged, edged, rotation, image.Point{edged.Cols(), edged.Cols()})
	gocv.WarpAffine(original, original, rotation, image.Point{edged.Cols(), edged.Cols()})

	contours := gocv.FindContours(edged, gocv.RetrievalList, gocv.ChainApproxSimple)
	for _, v := range contours {
		r := gocv.BoundingRect(v)
		if contourArea := r.Dx() * r.Dy(); contourArea > maxArea {
			rect = r
			maxArea = contourArea
		}
	}
	return original.Region(rect)
}
