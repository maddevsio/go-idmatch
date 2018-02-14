package preprocessing

import (
	"fmt"
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

func hBorder(img gocv.Mat) (h []int) {
	for i := 1; i < img.Rows(); i++ {
		if img.GetUCharAt(i, 1) != 0 {
			h = append(h, i)
			i += 5
		}
	}
	return
}

func vBorder(img gocv.Mat) (v []int) {
	for i := 1; i < img.Cols(); i++ {
		if img.GetUCharAt(1, i) != 0 {
			v = append(v, i)
			i += 5
		}
	}
	return
}

func contour(img gocv.Mat) image.Rectangle {
	var rect image.Rectangle
	hm1 := gocv.GetStructuringElement(gocv.MorphRect, image.Point{1, 15})
	hm2 := gocv.GetStructuringElement(gocv.MorphRect, image.Point{2, img.Cols() * 2})
	vm1 := gocv.GetStructuringElement(gocv.MorphRect, image.Point{15, 1})
	vm2 := gocv.GetStructuringElement(gocv.MorphRect, image.Point{img.Rows() * 2, 2})
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
		utils.ShowImage(img)
		res := gocv.NewMat()
		defer res.Close()
		gocv.BitwiseOr(horizontal, vertical, res)
		utils.ShowImage(res)
	}

	x := vBorder(vertical)
	y := hBorder(horizontal)

	// Ugly loop over all crossed lines with aspect ratio and area matching
	bestDelta, biggestArea, totalRects, matchRects := 0.1, 0.0, 0, 0
	imageArea := float64(img.Cols() * img.Rows())
	for top := 0; top < len(y)/2; top++ {
		for bottom := len(y) - 1; bottom > len(y)/2; bottom-- {
			for left := 0; left < len(x)/2; left++ {
				for right := len(x) - 1; right > len(x)/2; right-- {
					totalRects++
					r := image.Rectangle{image.Point{x[left], y[top]}, image.Point{x[right], y[bottom]}}
					area := float64(r.Dx() * r.Dy())
					areaRatio := area / imageArea
					switch {
					case areaRatio < 0.33 && biggestArea > 0:
						break
					case areaRatio > 0.97:
						continue
					default:
						matchRects++
						ratio := float64(r.Dx()) / float64(r.Dy())
						// Move aspect ratio to template
						delta := math.Abs(1.58 - ratio)
						if delta < bestDelta && biggestArea < area {
							biggestArea = area
							bestDelta = delta
							rect = r
						}
					}
				}
			}
		}
	}
	log.Print(log.DebugLevel, fmt.Sprintf("%d rectangles out of %d are matched\n", matchRects, totalRects))
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
	gocv.CvtColor(img, img, gocv.ColorRGBToGray)
	gocv.GaussianBlur(img, cleanCanny, image.Point{7, 7}, 10, 10, gocv.BorderDefault)
	gocv.Canny(cleanCanny, cleanCanny, 30, 170)
	gocv.GaussianBlur(img, img, image.Point{3, 3}, 5, 5, gocv.BorderDefault)

	rotation := rotate(cleanCanny)
	gocv.WarpAffine(img, img, rotation, image.Point{img.Cols(), img.Rows()})
	gocv.WarpAffine(original, original, rotation, image.Point{original.Cols(), original.Rows()})

	gocv.Canny(img, img, 10, 50)
	roi := original.Region(contour(img))

	if log.IsDebug() {
		utils.ShowImage(roi)
	}
	return roi
}
