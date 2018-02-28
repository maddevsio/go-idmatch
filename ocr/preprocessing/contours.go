package preprocessing

import (
	"image"
	"math"

	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/utils"
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
		}
	}
	theta = maxTheta * 180 / math.Pi
	if theta > 45 {
		theta -= 90
	}
	return gocv.GetRotationMatrix2D(image.Point{edged.Cols() / 2, edged.Rows() / 2}, theta, 1)
}

func hBorder(img gocv.Mat) (top, bottom []int) {
	for i := 1; i < img.Rows()/2; i++ {
		if img.GetUCharAt(i, 1) != 0 {
			top = append(top, i)
		}
	}
	for i := img.Rows() - 1; i > img.Rows()/2; i-- {
		if img.GetUCharAt(i, 1) != 0 {
			bottom = append(bottom, i)
		}
	}
	return
}

func vBorder(img gocv.Mat) (left, right []int) {
	for i := 1; i < img.Cols()/2; i++ {
		if img.GetUCharAt(1, i) != 0 {
			left = append(left, i)
		}
	}
	for i := img.Cols() - 1; i > img.Cols()/2; i-- {
		if img.GetUCharAt(1, i) != 0 {
			right = append(right, i)
		}
	}
	return
}

func contour(img gocv.Mat) image.Rectangle {
	var rect image.Rectangle
	hm1 := gocv.GetStructuringElement(gocv.MorphRect, image.Point{17, 1})
	hm2 := gocv.GetStructuringElement(gocv.MorphRect, image.Point{img.Cols() * 2, 1})
	vm1 := gocv.GetStructuringElement(gocv.MorphRect, image.Point{1, 17})
	vm2 := gocv.GetStructuringElement(gocv.MorphRect, image.Point{1, img.Rows() * 2})
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

	if log.IsDebug() {
		res := gocv.NewMat()
		defer res.Close()
		gocv.BitwiseOr(horizontal, vertical, res)
		utils.ShowImage(res)
	}

	top, bottom := hBorder(horizontal)
	left, right := vBorder(vertical)

	// Ugly loop over all crossed lines with aspect ratio and area matching
	bestRatio, bestArea := 1.0, 0.0
	for _, i := range top {
		for _, j := range bottom {
			for _, k := range left {
				for _, l := range right {
					r := image.Rectangle{image.Point{k, i}, image.Point{l, j}}
					ratio := float64(r.Dx()) / float64(r.Dy())
					area := float64(r.Dx() * r.Dy())
					// Move aspect ratio to template
					if math.Abs(1.58-ratio) < bestRatio && area > bestArea {
						bestRatio = ratio
						bestArea = area
						rect = r
					}
				}
			}
		}
	}
	return rect
}

// Contours takes image file path and crops it by contour
func Contours(file string) gocv.Mat {
	original := gocv.NewMat()
	defer original.Close()
	cleanCanny := gocv.NewMat()
	defer cleanCanny.Close()

	img := gocv.IMRead(file, gocv.IMReadColor)
	img.CopyTo(original)
	gocv.ApplyColorMap(img, img, gocv.ColormapHot)
	gocv.CvtColor(img, img, gocv.ColorRGBToGray)
	// gocv.MedianBlur(img, img, 3)
	gocv.GaussianBlur(img, img, image.Point{3, 3}, 7, 7, gocv.BorderDefault)
	gocv.Canny(img, cleanCanny, 30, 170)
	gocv.Canny(img, img, 10, 70)

	rotation := rotate(cleanCanny)
	gocv.WarpAffine(img, img, rotation, image.Point{img.Cols(), img.Rows()})
	gocv.WarpAffine(original, original, rotation, image.Point{img.Cols(), img.Rows()})

	roi := original.Region(contour(img))

	if log.IsDebug() {
		utils.ShowImage(roi)
	}
	return roi
}
