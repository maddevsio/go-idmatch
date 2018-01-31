package processing

import (
	"image"
	"os"
	"strconv"

	"github.com/otiai10/gosseract"
	"gocv.io/x/gocv"
)

type block struct {
	x, y, h, w int
	text       string
}

func TextRegions(img gocv.Mat) [][]image.Point {
	binarized := gocv.NewMat()
	gocv.CvtColor(img, binarized, gocv.ColorBGRToGray)
	kernel := gocv.GetStructuringElement(gocv.MorphEllipse, image.Point{5, 5})
	gocv.MorphologyEx(binarized, binarized, gocv.MorphGradient, kernel)
	gocv.Threshold(binarized, binarized, 0, 255, gocv.ThresholdBinary|gocv.ThresholdOtsu)

	connected := gocv.NewMat()
	kernel = gocv.GetStructuringElement(gocv.MorphEllipse, image.Point{9, 1})
	gocv.MorphologyEx(binarized, connected, gocv.MorphClose, kernel)

	return gocv.FindContours(connected, gocv.RetrievalCComp, gocv.ChainApproxSimple)
}

func extractText(file string) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()
	client.SetLanguage("kir", "eng")
	client.SetImage(file)
	text, err := client.Text()
	return text, err
}

func RecognizeRegions(img gocv.Mat, regions [][]image.Point) (result []block) {
	for k, v := range regions {
		region := gocv.BoundingRect(v)
		// Replace absolute size with relative values
		if region.Dx() < 16 || region.Dy() < 16 || region.Dy() > 64 {
			continue
		}
		roi := img.Region(region)
		file := strconv.Itoa(k) + ".jpeg"
		gocv.IMWrite(file, roi)
		text, err := extractText(file)
		if err != nil {
			continue
		}
		result = append(result, block{
			x:    region.Min.X,
			y:    region.Min.Y,
			w:    region.Dx(),
			h:    region.Dy(),
			text: text,
		})
		os.Remove(file)
		// gocv.Rectangle(img, gocv.BoundingRect(v), color.RGBA{255, 0, 0, 255}, 2)
	}
	// utils.ShowImage(img)
	return result
}
