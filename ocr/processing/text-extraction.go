package processing

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"os"
	"strconv"

	"github.com/otiai10/gosseract"
	"github.com/tzununbekov/go-idmatch/utils"
	"gocv.io/x/gocv"
)

type block struct {
	x, y, h, w float64
	text       string
}

func TextRegions(img gocv.Mat) [][]image.Point {
	/*old golang variant*/
	// binarized := gocv.NewMat()
	// gocv.CvtColor(img, binarized, gocv.ColorBGRToGray)
	// kernel := gocv.GetStructuringElement(gocv.MorphEllipse, image.Point{5, 5})
	// gocv.MorphologyEx(binarized, binarized, gocv.MorphGradient, kernel)
	// gocv.Threshold(binarized, binarized, 0, 255, gocv.ThresholdBinary|gocv.ThresholdOtsu)

	// connected := gocv.NewMat()
	// kernel = gocv.GetStructuringElement(gocv.MorphEllipse, image.Point{9, 1})
	// gocv.MorphologyEx(binarized, connected, gocv.MorphClose, kernel)

	// return gocv.FindContours(connected, gocv.RetrievalCComp, gocv.ChainApproxSimple)
	/*old python variant*/
	original := img.Clone()

	gray := gocv.NewMat()
	defer gray.Close()

	kernel := gocv.GetStructuringElement(gocv.MorphEllipse, image.Point{5, 7})
	defer kernel.Close()

	opening := gocv.NewMat()
	defer opening.Close()
	gocv.MorphologyEx(original, opening, gocv.MorphGradient, kernel)

	gocv.CvtColor(opening, gray, gocv.ColorBGRToGray)
	binarization := gocv.NewMat()
	defer binarization.Close()
	gocv.Threshold(gray, binarization, 0.0, 255.0, gocv.ThresholdBinary|gocv.ThresholdOtsu)

	kernel = gocv.GetStructuringElement(gocv.MorphRect, image.Point{9, 1})
	connected := gocv.NewMat()
	defer connected.Close()

	gocv.MorphologyEx(binarization, connected, gocv.MorphClose, kernel)
	return gocv.FindContours(connected, gocv.RetrievalCComp, gocv.ChainApproxSimple)

	/*old C variant*/
	// gray := gocv.NewMat()
	// defer gray.Close()

	// grad := gocv.NewMat()
	// defer grad.Close()

	// binarized := gocv.NewMat()
	// defer binarized.Close()

	// kernel := gocv.GetStructuringElement(gocv.MorphEllipse, image.Point{5, 7})
	// defer kernel.Close()

	// //convert to gray
	// gocv.CvtColor(img, gray, gocv.ColorBGRToGray)
	// gocv.MorphologyEx(gray, grad, gocv.MorphGradient, kernel)
	// gocv.Threshold(grad, binarized, 0, 255, gocv.ThresholdBinary|gocv.ThresholdOtsu)

	// connected := gocv.NewMat()
	// defer connected.Close()

	// kernel = gocv.GetStructuringElement(gocv.MorphRect, image.Point{9, 1})
	// gocv.MorphologyEx(binarized, connected, gocv.MorphClose, kernel)

	// return gocv.FindContours(connected, gocv.RetrievalCComp, gocv.ChainApproxSimple)
}

func RecognizeRegions(img gocv.Mat, regions [][]image.Point, preview string) (result []block, path string) {

	fmt.Println(preview)
	client := gosseract.NewClient()
	defer client.Close()
	client.SetLanguage("rus", "eng")

	gray := gocv.NewMat()
	defer gray.Close()

	gocv.CvtColor(img, gray, gocv.ColorBGRToGray)

	for k, v := range regions {
		region := gocv.BoundingRect(v)
		// Replace absolute size with relative values
		if region.Dx() < 20 || region.Dy() < 20 {
			continue
		}

		roi := gray.Region(region)
		file := strconv.Itoa(k) + ".jpeg"

		gocv.IMWrite(file, roi)
		client.SetImage(file)
		text, err := client.Text()

		if err != nil {
			continue
		}

		result = append(result, block{
			x:    float64(region.Min.X) / float64(img.Cols()),
			y:    float64(region.Min.Y) / float64(img.Rows()),
			w:    float64(region.Dx()) / float64(img.Cols()),
			h:    float64(region.Dy()) / float64(img.Rows()),
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
		utils.ShowImage(img)
	}
	return result, path
}
