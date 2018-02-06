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
	/*old python variant*/
	original := img.Clone()

	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(original, gray, gocv.ColorBGRToGray)
	utils.ShowImage(gray)

	//WARNING!!!
	//we need some document size and dpi based value!!!!!
	//We need to know maximum symbol width and hight in pixels
	//and stroke's width
	kernel := gocv.GetStructuringElement(gocv.MorphEllipse, image.Point{9, 9})
	defer kernel.Close()
	grad := gocv.NewMat()
	defer grad.Close()

	gocv.MorphologyEx(gray, grad, gocv.MorphGradient, kernel)
	utils.ShowImage(grad)

	binarization := gocv.NewMat()
	defer binarization.Close()

	gocv.Threshold(grad, binarization, 0.0, 255.0, gocv.ThresholdBinary|gocv.ThresholdOtsu)
	utils.ShowImage(binarization)

	opening := gocv.NewMat()
	defer opening.Close()
	kernel = gocv.GetStructuringElement(gocv.MorphRect, image.Point{7, 7})
	gocv.MorphologyEx(binarization, opening, gocv.MorphOpen, kernel)
	utils.ShowImage(opening)

	//AHTUNG!!!!! size is (height, width). not (width, height)
	kernel = gocv.GetStructuringElement(gocv.MorphRect, image.Point{1, 15})
	connected := gocv.NewMat()
	defer connected.Close()

	gocv.MorphologyEx(opening, connected, gocv.MorphClose, kernel)
	utils.ShowImage(connected)
	return gocv.FindContours(connected, gocv.RetrievalCComp, gocv.ChainApproxSimple)
}

func RecognizeRegions(img gocv.Mat, regions [][]image.Point, preview string) (result []block, path string) {
	client := gosseract.NewClient()
	defer client.Close()
	client.SetLanguage("kir", "eng")

	// gray := gocv.NewMat()
	// defer gray.Close()

	// gocv.CvtColor(img, gray, gocv.ColorBGRToGray)

	for k, v := range regions {
		region := gocv.BoundingRect(v)
		// Replace absolute size with relative values
		if region.Dx() < 10 || region.Dy() < 16 || region.Dy() > 60 {
			// fmt.Println("%d %d", region.Dx(), region.Dy())
			continue
		}

		roi := img.Region(region)
		file := strconv.Itoa(k) + ".jpeg"

		gocv.IMWrite(file, roi)
		client.SetImage(file)
		text, err := client.Text()

		if err != nil {
			continue
		}

		fmt.Println(text)

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
