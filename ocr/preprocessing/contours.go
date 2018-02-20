package preprocessing

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/maddevsio/go-idmatch/config"
	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/templates"
	"github.com/maddevsio/go-idmatch/utils"
	"gocv.io/x/gocv"
)

var (
	Img           gocv.Mat
	OriginalArea  int
	OriginalRatio float64
)

func rotate(edged gocv.Mat) gocv.Mat {
	var theta, maxTheta, maxDistance float64
	lines := gocv.NewMat()
	defer lines.Close()

	gocv.HoughLinesP(edged, lines, 1, math.Pi/180, config.Preprocessing.HoughThreshold)
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

// Contours takes image file path and crops it by contour
func Contours(file string, card templates.Card) gocv.Mat {
	original := gocv.NewMat()
	defer original.Close()

	cleanCanny := gocv.NewMat()
	defer cleanCanny.Close()

	img := gocv.IMRead(file, gocv.IMReadColor)
	k := float64(560) / float64(img.Rows())
	gocv.Resize(img, img, image.Point{0, 0}, k, k, gocv.InterpolationCubic)
	img.CopyTo(original)
	gocv.ApplyColorMap(img, img, gocv.ColormapHot)
	gocv.CvtColor(img, img, gocv.ColorRGBToGray)
	Img = gocv.NewMat()
	img.CopyTo(Img)

	OriginalArea = original.Cols() * original.Rows()
	OriginalRatio = card.AspectRatio

	GetFactors()
	defer Img.Close()

	if log.IsDebug() {
		gocv.Rectangle(original, Rect, color.RGBA{255, 0, 0, 255}, 2)
		// utils.ShowImage(img)
		fmt.Println(Rect)
		utils.ShowImage(original.Region(Rect))
	}
	return original.Region(Rect)
}
