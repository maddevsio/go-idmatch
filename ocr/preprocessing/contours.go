package preprocessing

import (
	"image"
	"math"

	"gocv.io/x/gocv"
)

func rotate(edged gocv.Mat) gocv.Mat {
	var theta, maxTheta, maxDistance float64
	lines := gocv.NewMat()
	defer lines.Close()

	gocv.HoughLinesP(edged, lines, 1, math.Pi/180, 200)
	for row := 0; row < lines.Rows(); row++ {
		x1, y1, x2, y2 := lines.GetIntAt(row, 0), lines.GetIntAt(row, 1), lines.GetIntAt(row, 2), lines.GetIntAt(row, 3)
		if distance := math.Sqrt(math.Pow(float64(x2-x1), 2) + math.Pow(float64(y2-y1), 2)); distance > maxDistance {
			theta = math.Atan2(float64(y2-y1), float64(x2-x1))
			if math.Abs(theta) == math.Pi/2 {
				continue
			}
			maxDistance = distance
			maxTheta = theta
			// gocv.Line(original, image.Point{int(x1), int(y1)}, image.Point{int(x2), int(y2)}, color.RGBA{255, 0, 0, 255}, 2)
		}
	}
	theta = maxTheta * 180 / math.Pi
	if theta > 45 {
		theta -= 90
	}
	// utils.ShowImage(original)
	return gocv.GetRotationMatrix2D(image.Point{edged.Cols() / 2, edged.Rows() / 2}, theta, 1)
}

func hBorder(img gocv.Mat) (top, bottom int) {
	for i := 1; i < img.Rows(); i++ {
		if img.GetUCharAt(i, 1) != 0 {
			top = i
			break
		}
	}
	for i := img.Rows() - 1; i > 0; i-- {
		if img.GetUCharAt(i, 1) != 0 {
			bottom = i
			break
		}
	}
	return
}

func vBorder(img gocv.Mat) (left, right int) {
	for i := 1; i < img.Cols(); i++ {
		if img.GetUCharAt(1, i) != 0 {
			left = i
			break
		}
	}
	for i := img.Cols() - 1; i > 0; i-- {
		if img.GetUCharAt(1, i) != 0 {
			right = i
			break
		}
	}
	return
}

func contour(img gocv.Mat) image.Rectangle {
	hm1 := gocv.GetStructuringElement(gocv.MorphRect, image.Point{1, 20})
	hm2 := gocv.GetStructuringElement(gocv.MorphRect, image.Point{1, img.Cols() * 2})
	vm1 := gocv.GetStructuringElement(gocv.MorphRect, image.Point{20, 1})
	vm2 := gocv.GetStructuringElement(gocv.MorphRect, image.Point{img.Rows() * 2, 1})
	defer hm1.Close()
	defer hm2.Close()
	defer vm1.Close()
	defer vm2.Close()

	horizontal := gocv.NewMat()
	vertical := gocv.NewMat()
	defer horizontal.Close()
	defer vertical.Close()

	gocv.Erode(img, horizontal, hm1)
	gocv.Dilate(horizontal, horizontal, hm2)

	gocv.Erode(img, vertical, vm1)
	gocv.Dilate(vertical, vertical, vm2)

	top, bottom := hBorder(horizontal)
	left, right := vBorder(vertical)

	p1 := image.Point{left, top}
	p2 := image.Point{right, bottom}

	return image.Rectangle{p1, p2}
}

// Contours takes image file path and crops it by contour
func Contours(file string) gocv.Mat {
	original := gocv.NewMat()
	defer original.Close()

	img := gocv.IMRead(file, gocv.IMReadColor)
	img.CopyTo(original)
	gocv.ApplyColorMap(img, img, gocv.ColormapHot)
	gocv.CvtColor(img, img, gocv.ColorRGBToGray)
	gocv.GaussianBlur(img, img, image.Point{3, 3}, 7, 7, gocv.BorderDefault)
	gocv.Canny(img, img, 20, 150)

	rotation := rotate(img)
	gocv.WarpAffine(img, img, rotation, image.Point{img.Cols(), img.Rows()})
	gocv.WarpAffine(original, original, rotation, image.Point{img.Cols(), img.Rows()})

	// utils.ShowImage(original.Region(contour(img)))

	return original.Region(contour(img))
}
