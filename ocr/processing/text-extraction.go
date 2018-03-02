package processing

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"os"
	"strconv"

	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/utils"
	"github.com/otiai10/gosseract"
	"gocv.io/x/gocv"
)

type block struct {
	x, y, h, w float64
	text       string
}

var symbolHeightCoeff float64 = 38.0 / 737.0
var symbolWidthCoeff float64 = 27.0 / 1170.0

var strokeWidthCoeff float64 = 5.0 / 1170.0
var strokeHeightCoeff float64 = 5.0 / 737.0

func TextRegions(img gocv.Mat) [][]image.Point {
	// We have to get these values from JSON or somehow from document
	symbolWidth := int(float64(img.Cols()) * symbolWidthCoeff)
	symbolHeight := int(float64(img.Rows()) * symbolHeightCoeff)

	strokeWidth := int(float64(img.Cols()) * strokeWidthCoeff)
	strokeHeight := int(float64(img.Rows()) * strokeHeightCoeff)

	fmt.Println(symbolWidth, symbolHeight, strokeWidth, strokeHeight)
	original := img.Clone()

	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(original, gray, gocv.ColorBGRToGray)
	utils.ShowImageInNamedWindow(gray, "text regions: gray")

	//WARNING!!!
	//we need some document size and dpi based value!!!!!
	//We need to know maximum symbol width and hight in pixels
	//and stroke's width
	kernel := gocv.GetStructuringElement(gocv.MorphEllipse,
		image.Point{symbolHeight / 2, symbolWidth / 2})
	defer kernel.Close()
	grad := gocv.NewMat()
	defer grad.Close()

	gocv.MorphologyEx(gray, grad, gocv.MorphGradient, kernel)
	utils.ShowImageInNamedWindow(grad, "text regions: gradient")
	binarization := gocv.NewMat()
	defer binarization.Close()

	gocv.Threshold(grad, binarization, 0.0, 255.0, gocv.ThresholdBinary|gocv.ThresholdOtsu)
	utils.ShowImageInNamedWindow(binarization, "text regions: binarization")

	opening := gocv.NewMat()
	defer opening.Close()
	kernel = gocv.GetStructuringElement(gocv.MorphRect,
		image.Point{symbolHeight * 2 / 3, symbolWidth / 2})
	gocv.MorphologyEx(binarization, opening, gocv.MorphOpen, kernel)
	utils.ShowImageInNamedWindow(opening, "text regions: opening")

	kernel = gocv.GetStructuringElement(gocv.MorphRect, image.Point{symbolWidth, 1})
	connected := gocv.NewMat()
	defer connected.Close()

	gocv.MorphologyEx(opening, connected, gocv.MorphClose, kernel)
	utils.ShowImageInNamedWindow(connected, "text regions: connected")

	return gocv.FindContours(connected, gocv.RetrievalCComp, gocv.ChainApproxSimple)
}

func RecognizeRegions(img gocv.Mat, regions [][]image.Point, preview string) (result []block, path string) {
	//We have to get these values from JSON or somehow from document

	symbolWidth := int(float64(img.Cols()) * symbolWidthCoeff)
	symbolHeight := int(float64(img.Rows()) * symbolHeightCoeff)

	client := gosseract.NewClient()
	defer client.Close()
	client.SetLanguage("kir", "eng")

	gray := gocv.NewMat()
	defer gray.Close()

	gocv.CvtColor(img, gray, gocv.ColorBGRToGray)

	for k, v := range regions {
		rect := gocv.BoundingRect(v)
		// Replace absolute size with relative values
		// roi := img.Region(rect)
		roi := gray.Region(rect)
		if rect.Dx() < symbolWidth || rect.Dy() < symbolHeight/2 || rect.Dy() > symbolHeight*3 {
			continue
		}

		file := strconv.Itoa(k) + ".jpeg"

		roix4 := gocv.NewMat()
		defer roix4.Close()

		gocv.Resize(roi, roix4, image.Point{0, 0}, 4, 4, gocv.InterpolationCubic)
		gocv.IMWrite(file, roix4)
		client.SetImage(file)

		text, err := client.Text()
		if err != nil {
			continue
		}

		log.Print(log.DebugLevel, text)
		// utils.ShowImageInNamedWindow(roix4, fmt.Sprintf("RecognizeRegions: %d %d", rect.Dx(), rect.Dy()))

		result = append(result, block{
			x:    float64(rect.Min.X) / float64(img.Cols()),
			y:    float64(rect.Min.Y) / float64(img.Rows()),
			w:    float64(rect.Dx()) / float64(img.Cols()),
			h:    float64(rect.Dy()) / float64(img.Rows()),
			text: text,
		})

		os.Remove(file)
		gocv.Rectangle(img, gocv.BoundingRect(v), color.RGBA{255, 0, 0, 255}, 2)
	}

	if len(preview) != 0 {
		hash := md5.New()
		hash.Write(img.ToBytes())
		path = preview + "/" + hex.EncodeToString(hash.Sum(nil)) + ".jpeg"
		gocv.IMWrite(path, img)
	}

	utils.ShowImage(img)

	return result, path
}
