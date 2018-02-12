package processing

import (
	"crypto/md5"
	"encoding/hex"
	"image"
	"image/color"
	"os"
	"strconv"

	"github.com/maddevsio/go-idmatch/log"
	"github.com/otiai10/gosseract"
	"gocv.io/x/gocv"
)

type block struct {
	x, y, h, w float64
	text       string
}

func TextRegions(img gocv.Mat) [][]image.Point {

	//We have to get these values from JSON or somehow from document
	hc := 1.0 / 25.0
	wc := 1.0 / 41.0
	sw := int(float64(img.Cols()) * wc)
	sh := int(float64(img.Rows()) * hc)

	original := img.Clone()

	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(original, gray, gocv.ColorBGRToGray)
	// utils.ShowImageInNamedWindow(gray, "text regions: gray")

	//WARNING!!!
	//we need some document size and dpi based value!!!!!
	//We need to know maximum symbol width and hight in pixels
	//and stroke's width
	kernel := gocv.GetStructuringElement(gocv.MorphEllipse, image.Point{sw / 2, sh / 2})
	defer kernel.Close()
	grad := gocv.NewMat()
	defer grad.Close()

	//try canny
	gocv.MorphologyEx(gray, grad, gocv.MorphGradient, kernel)
	// utils.ShowImageInNamedWindow(grad, "text regions: gradient")
	binarization := gocv.NewMat()
	defer binarization.Close()

	gocv.Threshold(grad, binarization, 0.0, 255.0, gocv.ThresholdBinary|gocv.ThresholdOtsu)
	// utils.ShowImageInNamedWindow(binarization, "text regions: binarization")
	//

	opening := gocv.NewMat()
	defer opening.Close()
	kernel = gocv.GetStructuringElement(gocv.MorphRect, image.Point{sw / 2, sh * 2 / 3})
	gocv.MorphologyEx(binarization, opening, gocv.MorphOpen, kernel)
	// utils.ShowImageInNamedWindow(opening, "text regions: opening")

	//AHTUNG!!!!! size is (height, width). not (width, height)
	kernel = gocv.GetStructuringElement(gocv.MorphRect, image.Point{1, sw})
	connected := gocv.NewMat()
	defer connected.Close()

	gocv.MorphologyEx(opening, connected, gocv.MorphClose, kernel)
	// utils.ShowImageInNamedWindow(connected, "text regions: connected")

	return gocv.FindContours(connected, gocv.RetrievalCComp, gocv.ChainApproxSimple)
}

func RecognizeRegions(img gocv.Mat, regions [][]image.Point, preview string) (result []block, path string) {
	//We have to get these values from JSON or somehow from document
	hc := 1.0 / 25.0
	wc := 1.0 / 41.0
	sw := int(float64(img.Cols()) * wc)
	sh := int(float64(img.Rows()) * hc)

	client := gosseract.NewClient()
	defer client.Close()
	client.SetLanguage("kir", "eng")

	for k, v := range regions {
		rect := gocv.BoundingRect(v)
		// Replace absolute size with relative values
		if rect.Dx() < sw || rect.Dy() < sh || rect.Dy() > sh*3 {
			continue
		}

		roi := img.Region(rect)
		file := strconv.Itoa(k) + ".jpeg"

		roix2 := gocv.NewMat()
		defer roix2.Close()

		gocv.Resize(roi, roix2, image.Point{0, 0}, 4, 4, gocv.InterpolationCubic)

		gocv.IMWrite(file, roix2)
		client.SetImage(file)

		text, err := client.Text()
		if err != nil {
			continue
		}

		log.Print(log.DebugLevel, text)
		// utils.ShowImageInNamedWindow(roix2, fmt.Sprintf("RecognizeRegions: %d %d", rect.Dx(), rect.Dy()))

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
		// utils.ShowImage(img)
	}

	return result, path
}
