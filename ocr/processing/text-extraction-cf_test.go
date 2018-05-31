package processing

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"math"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/templates"
	"github.com/maddevsio/go-idmatch/utils"
	"gocv.io/x/gocv"
)

type cfRect struct {
	X0 int `json:"x0"`
	Y0 int `json:"y0"`
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
}

type cfRectDeviation struct {
	X float64 `json:"X"`
	Y float64 `json:"Y"`
}

type cfRealSize struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type coefFinder struct {
	MaxSizes templates.MaxQualitySizesT `json:"maxQualitySizes"`

	HqImage     string          `json:"hqImage"`
	HqRects     []cfRect        `json:"hqRects"`
	HqDeviation cfRectDeviation `json:"hqDeviation"`
	HqRealSize  cfRealSize      `json:"hqRealSize"`

	LqImage     string          `json:"lqImage"`
	LqRects     []cfRect        `json:"lqRects"`
	LqDeviation cfRectDeviation `json:"lqDeviation"`
	LqRealSize  cfRealSize      `json:"lqRealSize"`
}

func compareRects(x00, y00, x01, y01, x10, y10, x11, y11 int, devX, devY float64) bool {
	return math.Abs(float64(x10-x00)) <= devX &&
		math.Abs(float64(y10-y00)) <= devY &&
		math.Abs(float64(x11-x01)) <= devX &&
		math.Abs(float64(y11-y01)) <= devY
}

func checkRegions(regions [][]image.Point, rects []cfRect, devX, devY float64) bool {
	count := 0
	for _, regIn := range regions {
		rect := gocv.BoundingRect(regIn)
		for _, regEt := range rects {
			if compareRects(rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y,
				regEt.X0, regEt.Y0, regEt.X1, regEt.Y1, devX, devY) {
				count++
			}
		}
	}
	return count >= len(rects) //warning!
}

func loadFinderJSON(jsonFilePath string) (coefFinder, error) {
	var cf coefFinder
	fp, err := filepath.Glob(jsonFilePath)

	if err != nil {
		return cf, err
	}

	if len(fp) <= 0 {
		return cf, errors.New("len(fp) <= 0")
	}

	jsonFile, err := ioutil.ReadFile(fp[0])
	if err != nil {
		return cf, err
	}

	err = json.Unmarshal(jsonFile, &cf)
	if err != nil {
		return cf, err
	}
	return cf, nil
}

//takes much time. WARNING!!!
func tryToFindCoeff(img gocv.Mat, cf coefFinder) []extractTextRegionIntCoeff {
	lstResult := make([]extractTextRegionIntCoeff, 0)
	const max = 22
	maxW := max
	maxH := max
	maxW2 := max
	maxH2 := max
	start := time.Now()
	for w := maxW; w >= 2; w-- {
		for h := maxH; h >= 2; h-- {
			for w2 := maxW2; w2 >= 2; w2-- {
				for h2 := maxH2; h2 >= 2; h2-- {
					regions, err := textRegionsInternal(img, cf.MaxSizes, extractTextRegionIntCoeff{w, h, w2, h2})
					if err != nil {
						continue
					}

					if !checkRegions(regions, cf.LqRects, cf.LqDeviation.X, cf.LqDeviation.Y) {
						continue
					}

					nItem := extractTextRegionIntCoeff{w1: w, h1: h, w2: w2, h2: h2}
					lstResult = append(lstResult, nItem)
				} //for h2
			} //for w2
		} //for h
	} // for w
	end := time.Now()
	diff := end.Sub(start)
	fmt.Println("Coefficient search took: ", diff)
	return lstResult
}

func TestTryToFindCoeff(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	cf, err := loadFinderJSON("json/kg_idcard_old.json")
	if !assert.NoError(t, err) {
		return
	}

	imgLq := gocv.IMRead(cf.LqImage, gocv.IMReadColor)
	defer imgLq.Close()
	imgHq := gocv.IMRead(cf.HqImage, gocv.IMReadColor)
	defer imgHq.Close()

	lstIntCoefficients := tryToFindCoeff(imgLq, cf)
	lstFloatCoefficients := make([]extractTextRegionFloatCoeff, 0, len(lstIntCoefficients))

	for _, ic := range lstIntCoefficients {
		var nItem extractTextRegionFloatCoeff
		nItem.w1 = float64(ic.w1) / float64(imgLq.Cols())
		nItem.w2 = float64(ic.w2) / float64(imgLq.Cols())
		nItem.h1 = float64(ic.h1) / float64(imgLq.Rows())
		nItem.h2 = float64(ic.h2) / float64(imgLq.Rows())
		lstFloatCoefficients = append(lstFloatCoefficients, nItem)
	}

	lstHqFloatCoefficients := make([]extractTextRegionFloatCoeff, 0, len(lstFloatCoefficients))
	for _, fc := range lstFloatCoefficients {
		w1c := fc.w1 * float64(imgHq.Cols())
		h1c := fc.h1 * float64(imgHq.Rows())

		w2c := fc.w2 * float64(imgHq.Cols())
		h2c := fc.h2 * float64(imgHq.Rows())

		regions, err := textRegionsInternal(imgHq, cf.MaxSizes, extractTextRegionIntCoeff{int(w1c), int(h1c), int(w2c), int(h2c)})
		if err != nil {
			continue
		}

		if !checkRegions(regions, cf.HqRects, cf.HqDeviation.X, cf.HqDeviation.Y) {
			continue
		}
		lstHqFloatCoefficients = append(lstHqFloatCoefficients, fc)
	}

	fmt.Println("FOUND COEFFICIENTS:", len(lstHqFloatCoefficients))
	for _, fc := range lstHqFloatCoefficients {
		fmt.Printf("{\"w1\":%f, \"h1\":%f, \"w2\":%f, \"h2\":%f},\n", fc.w1, fc.h1, fc.w2, fc.h2)
	}
	fmt.Println("*********")
}

func TestApplySetOfCoefficientsToSetOfImages(t *testing.T) {
	t.Fail()
}

func TestShowExampleHighQualityRectangles(t *testing.T) {
	cf, err := loadFinderJSON("json/kg_idcard_old.json")
	if !assert.NoError(t, err) {
		return
	}

	img := gocv.IMRead(cf.HqImage, gocv.IMReadColor)
	defer img.Close()

	const xoff = 0
	const yoff = 0
	log.SetLevel(log.DebugLevel)
	for _, r := range cf.HqRects {
		rect := image.Rectangle{image.Point{r.X0 + xoff, r.Y0 + yoff}, image.Point{r.X1 + xoff, r.Y1 + yoff}}
		gocv.Rectangle(&img, rect, color.RGBA{0, 255, 0, 255}, 1)
		fmt.Printf("{\"X0\":%d,\"Y0\":%d,\"X1\":%d,\"Y1\":%d},\n", r.X0+xoff, r.Y0+yoff, r.X1+xoff, r.Y1+yoff)
	}

	utils.ShowImageInNamedWindow(img, "TestShowExampleHighQualityRectangles")
}

func TestShowExampleLowQualityRectangles(t *testing.T) {
	cf, err := loadFinderJSON("json/kg_idcard_old.json")
	if !assert.NoError(t, err) {
		return
	}

	img := gocv.IMRead(cf.LqImage, gocv.IMReadColor)
	defer img.Close()

	const xoff = 0
	const yoff = 0
	log.SetLevel(log.DebugLevel)
	for _, r := range cf.LqRects {
		rect := image.Rectangle{image.Point{r.X0 + xoff, r.Y0 + yoff}, image.Point{r.X1 + xoff, r.Y1 + yoff}}
		gocv.Rectangle(&img, rect, color.RGBA{0, 255, 0, 255}, 1)
		fmt.Printf("{\"X0\":%d,\"Y0\":%d,\"X1\":%d,\"Y1\":%d},\n", r.X0+xoff, r.Y0+yoff, r.X1+xoff, r.Y1+yoff)
		// utils.ShowImageInNamedWindow(img, "ol")
	}
	utils.ShowImageInNamedWindow(img, "TestShowExampleLowQualityRectangles")
}
