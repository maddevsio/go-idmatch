package preprocessing

import (
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
	img.CopyTo(original)
	gocv.ApplyColorMap(img, img, gocv.ColormapHot)
	gocv.CvtColor(img, img, gocv.ColorRGBToGray)
	Img = gocv.NewMat()
	img.CopyTo(Img)
	gocv.GaussianBlur(img, cleanCanny, image.Point{config.Preprocessing.CleanCannyBlurSize, config.Preprocessing.CleanCannyBlurSize},
		config.Preprocessing.CleanCannyBlurSigma, config.Preprocessing.CleanCannyBlurSigma, gocv.BorderDefault)
	gocv.Canny(cleanCanny, cleanCanny, config.Preprocessing.CleanCannyT1, config.Preprocessing.CleanCannyT2)

	rotation := rotate(cleanCanny)
	gocv.WarpAffine(img, img, rotation, image.Point{img.Cols(), img.Rows()})
	gocv.WarpAffine(original, original, rotation, image.Point{original.Cols(), original.Rows()})

	OriginalArea = original.Cols() * original.Rows()
	OriginalRatio = float64(85.60) / float64(53.98)

	val := GetFactors()
	defer Img.Close()

	if int(val[0])%2 != 1 {
		val[0]++
	}

	gocv.MedianBlur(img, img, int(val[0]))
	gocv.Canny(img, img, float32(val[1]), float32(val[2]))
	contours := gocv.FindContours(img, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	var maxArea int
	var maxRect image.Rectangle
	for _, v := range contours {
		rect := gocv.BoundingRect(v)
		if area := rect.Dx() * rect.Dy(); area > maxArea {
			maxArea = area
			maxRect = rect
		}
	}
	if log.IsDebug() {
		gocv.Rectangle(original, maxRect, color.RGBA{255, 0, 0, 255}, 2)
		utils.ShowImage(img)
		utils.ShowImage(original.Region(maxRect))
	}
	return original.Region(maxRect)
}
