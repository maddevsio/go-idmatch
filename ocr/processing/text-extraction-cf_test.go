package processing

import (
	"encoding/json"
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

type coefFinder struct {
	HqImage     string          `json:"hqImage"`
	HqRects     []cfRect        `json:"hqRects"`
	HqDeviation cfRectDeviation `json:"hqDeviation"`

	LqImage     string          `json:"lqImage"`
	LqRects     []cfRect        `json:"lqRects"`
	LqDeviation cfRectDeviation `json:"lqDeviation"`
}

type extractTextRegionIntCoeff struct {
	w1, h1, w2, h2 int
}

type extractTextRegionFloatCoeff struct {
	w1, h1, w2, h2 float64
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

func tryToFindCoeffForNewID(img gocv.Mat, card templates.Card, cf coefFinder) {
	//takes much time
	const max = 20
	maxW := max
	maxH := max
	maxW2 := max
	maxH2 := max
	index := 0

	fmt.Println("****************************")
	start := time.Now()
	for w := maxW; w >= 2; w-- {
		for h := maxH; h >= 2; h-- {
			for w2 := maxW2; w2 >= 2; w2-- {
				for h2 := maxH2; h2 >= 2; h2-- {
					regions, err := textRegionsInternal(img, card, extractTextRegionIntCoeff{w, h, w2, h2})
					if err != nil {
						continue
					}

					if !checkRegions(regions, cf.LqRects, cf.LqDeviation.X, cf.LqDeviation.Y) {
						continue
					}

					original2 := img.Clone()
					for _, v := range regions {
						rect := gocv.BoundingRect(v)
						gocv.Rectangle(&original2, rect, color.RGBA{0, 255, 0, 255}, 2)
					}
					fmt.Printf("{%d, %d, %d, %d}, ", w, h, w2, h2)
					index++
					if index == 5 {
						fmt.Printf("\n")
						index = 0
					}
					original2.Close()
				} //for h2
			} //for w2
		} //for h
	} // for w
	fmt.Printf("\n")

	end := time.Now()
	diff := end.Sub(start)
	fmt.Println(diff)
}

func TestShowExampleHighQualityRectangles(t *testing.T) {
	var cf coefFinder
	fp, err := filepath.Glob("json/kg_idcard_new.json")

	if !assert.NoError(t, err) {
		return
	}

	if !assert.True(t, len(fp) > 0) {
		return
	}

	jsonFile, err := ioutil.ReadFile(fp[0])
	if !assert.NoError(t, err) {
		return
	}

	err = json.Unmarshal(jsonFile, &cf)
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
		gocv.Rectangle(&img, rect, color.RGBA{0, 255, 0, 255}, 2)
	}

	utils.ShowImageInNamedWindow(img, "TestShowExampleHighQualityRectangles")
}

func TestShowExampleLowQualityRectangles(t *testing.T) {
	var cf coefFinder
	fp, err := filepath.Glob("json/kg_idcard_new.json")

	if !assert.NoError(t, err) {
		return
	}

	if !assert.True(t, len(fp) > 0) {
		return
	}

	jsonFile, err := ioutil.ReadFile(fp[0])
	if !assert.NoError(t, err) {
		return
	}

	err = json.Unmarshal(jsonFile, &cf)
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
		gocv.Rectangle(&img, rect, color.RGBA{0, 255, 0, 255}, 2)
		fmt.Printf("{\"X0\":%d, \"Y0\": %d,\"X1\": %d,\"Y1\": %d},\n",
			r.X0+xoff, r.Y0+yoff, r.X1+xoff, r.Y1+yoff)
	}

	utils.ShowImageInNamedWindow(img, "TestShowExampleLowQualityRectangles")
}
