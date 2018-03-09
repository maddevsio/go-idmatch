package processing

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
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

func TextRegions(img gocv.Mat) ([][]image.Point, error) {
	// We have to get these values from JSON or somehow from document
	symbolWidth := int(float64(img.Cols()) * symbolWidthCoeff)
	symbolHeight := int(float64(img.Rows()) * symbolHeightCoeff)

	if symbolWidth < 2 || symbolHeight < 2 {
		return nil, errors.New("Symbol's size too small. Something wrong with region at all")
	}

	original := img.Clone()
	gray := gocv.NewMat()
	defer gray.Close()

	gocv.CvtColor(original, gray, gocv.ColorBGRToGray)
	utils.ShowImageInNamedWindow(gray, "text regions: gray")

	log.Print(log.DebugLevel, fmt.Sprintf("%d %d", symbolWidth, symbolHeight))

	kernel := gocv.GetStructuringElement(gocv.MorphEllipse,
		image.Point{10, 10})
	defer kernel.Close()

	grad := gocv.NewMat()
	defer grad.Close()
	binarization := gocv.NewMat()
	defer binarization.Close()
	opening := gocv.NewMat()
	defer opening.Close()
	connected := gocv.NewMat()
	defer connected.Close()

	for w := 10; w >= 2; w-- {
		for h := 10; h >= 2; h-- {
			kernel = gocv.GetStructuringElement(gocv.MorphEllipse, image.Point{w, h})
			gocv.MorphologyEx(gray, grad, gocv.MorphGradient, kernel)
			// utils.ShowImageInNamedWindow(grad, fmt.Sprintf("text regions: gradient. w : %d, h : %d", w, h))

			gocv.Threshold(grad, binarization, 0.0, 255.0, gocv.ThresholdBinary|gocv.ThresholdOtsu)
			// utils.ShowImageInNamedWindow(binarization, fmt.Sprintf("text regions: binarization. w : %d, h : %d", w, h))

			for w2 := 10; w2 >= 2; w2-- {
				for h2 := 10; h2 >= 2; h2-- {
					kernel = gocv.GetStructuringElement(gocv.MorphRect,
						image.Point{3, 3})
					gocv.MorphologyEx(binarization, opening, gocv.MorphOpen, kernel)
					// utils.ShowImageInNamedWindow(opening, fmt.Sprintf("text regions: opening. w : %d, h : %d", w2, h2))

					kernel = gocv.GetStructuringElement(gocv.MorphRect, image.Point{symbolWidth, 1})

					gocv.MorphologyEx(opening, connected, gocv.MorphClose, kernel)
					utils.ShowImageInNamedWindow(connected,
						fmt.Sprintf("text regions: connected. w1 : %d, h1 : %d, w2 : %d, h2 : %d", w, h, w2, h2))
				}
			}
		}
	}

	return gocv.FindContours(connected, gocv.RetrievalCComp, gocv.ChainApproxSimple), nil
}

func RecognizeRegions(img gocv.Mat, regions [][]image.Point, preview string) (result []block, path string) {
	//We have to get these values from JSON or somehow from document

	symbolWidth := int(float64(img.Cols()) * symbolWidthCoeff)
	symbolHeight := int(float64(img.Rows()) * symbolHeightCoeff)

	client := gosseract.NewClient()
	defer client.Close()

	client.SetLanguage("rus")

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

		text := "hoho"
		// text, err := client.Text()
		// if err != nil {
		// 	continue
		// }

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
