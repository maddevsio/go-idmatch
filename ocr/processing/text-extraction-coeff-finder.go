package processing

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"github.com/maddevsio/go-idmatch/utils"
	"gocv.io/x/gocv"
)

type frect struct {
	x0, y0, x1, y1 int
}

var newIDHQRects []frect = []frect{
	{1584, 1727, 3220, 1862}, {2780, 2297, 3410, 2416}, {2794, 2017, 3446, 2128},
	{1584, 1967, 2200, 2072}, {1590, 1527, 1906, 1632}, {1593, 1320, 2248, 1425},
	{1587, 1130, 1998, 1228}, {1587, 1005, 2015, 1109}, {1585, 808, 2498, 919},
	{1589, 679, 2464, 790},
}

var newIDLowQRects []frect = []frect{
	{457, 372, 561, 391}, {457, 326, 567, 345}, {262, 317, 366, 336},
	{263, 280, 527, 300}, {263, 246, 318, 264}, {263, 212, 373, 232},
	{263, 182, 330, 200}, {263, 163, 337, 180}, {263, 131, 413, 151},
	{263, 111, 408, 131},
}

var newIDLowQRects2 []frect = []frect{
	{184, 219, 261, 232}, {184, 192, 376, 206}, {184, 144, 316, 157},
	{184, 168, 222, 180},
	{184, 123, 243, 135}, {184, 109, 242, 121}, {184, 87, 291, 100},
	{184, 73, 283, 86},
}

type extractTextRegionIntCoeff struct {
	w1, h1, w2, h2 int
}

type extractTextRegionFloatCoeff struct {
	w1, h1, w2, h2 float64
}

var newIDLowQFloatCoeffArr = []extractTextRegionFloatCoeff{
	{0.008511, 0.006689, 0.010638, 0.006689}, {0.006383, 0.006689, 0.008511, 0.006689},
}

var newIDLowQIntCoeffArr = []extractTextRegionIntCoeff{
	{4, 2, 5, 2}, {3, 2, 4, 2},
}

func compareRects(x00, y00, x01, y01, x10, y10, x11, y11 int, devX, devY float64) bool {
	return math.Abs(float64(x10-x00)) <= devX &&
		math.Abs(float64(y10-y00)) <= devY &&
		math.Abs(float64(x11-x01)) <= devX &&
		math.Abs(float64(y11-y01)) <= devY
}

func checkRegionsNewID(regions [][]image.Point, rects []frect, devX, devY float64) bool {
	count := 0
	for _, regIn := range regions {
		rect := gocv.BoundingRect(regIn)
		for _, regEt := range rects {
			if compareRects(rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y,
				regEt.x0, regEt.y0, regEt.x1, regEt.y1, devX, devY) {
				count++
			}
		}
	}
	return count >= len(rects) //warning!
}

func tryToFindCoeffForNewID(img gocv.Mat) {
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
					regions, err := textRegionsInternal(img, extractTextRegionIntCoeff{w, h, w2, h2})
					if err != nil {
						continue
					}

					if !checkRegionsNewID(regions, newIDLowQRects2, 3.0, 3.0) {
						continue
					}

					original2 := img.Clone()
					for _, v := range regions {
						rect := gocv.BoundingRect(v)
						gocv.Rectangle(&original2, rect, color.RGBA{0, 255, 0, 255}, 1)
						// utils.ShowImageInNamedWindow(original2, "tototo")
						// fmt.Printf("{%d, %d, %d, %d}, ", rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y)
					}
					// utils.ShowImageInNamedWindow(original2, "tototo")
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

func buildFloatCoeffs(img gocv.Mat) {
	for ix, fc := range newIDLowQIntCoeffArr {
		w1c := float64(fc.w1) / float64(img.Cols())
		h1c := float64(fc.h1) / float64(img.Rows())
		w2c := float64(fc.w2) / float64(img.Cols())
		h2c := float64(fc.h2) / float64(img.Rows())

		if ix%10 == 0 {
			fmt.Printf("\n")
		}
		fmt.Printf("{%f, %f, %f, %f}, ", w1c, h1c, w2c, h2c)
	}
}

func showExampleRectangles(img gocv.Mat) {
	original2 := img.Clone()
	defer original2.Close()

	for ix, r := range newIDHQRects {
		rect := image.Rectangle{image.Point{r.x0, r.y0}, image.Point{r.x1, r.y1}}
		gocv.Rectangle(&original2, rect, color.RGBA{0, 255, 0, 255}, 1)
		// fmt.Printf("{%d, %d, %d, %d}, ", r.x0, r.y0, r.x1, r.y1)
		w1 := float64(rect.Min.X) / float64(img.Cols())
		h1 := float64(rect.Min.Y) / float64(img.Rows())
		fmt.Printf("%f, %f\n", w1, h1)
		utils.ShowImageInNamedWindow(original2, fmt.Sprintf("%d", ix))
	}
}

func testCoefficientsForID(img gocv.Mat) {
	arr := make([]int, 0)

	for ix, fc := range newIDLowQFloatCoeffArr {
		w1c := fc.w1 * float64(img.Cols())
		h1c := fc.h1 * float64(img.Rows())
		w2c := fc.w2 * float64(img.Cols())
		h2c := fc.h2 * float64(img.Rows())

		regions, err := textRegionsInternal(img, extractTextRegionIntCoeff{
			int(w1c), int(h1c), int(w2c), int(h2c)})

		if err != nil {
			fmt.Println(err)
			continue
		}

		// if !checkRegionsNewID(regions, newIDHQRects, 50.0, 50.0) {
		// 	continue
		// }

		original2 := img.Clone()
		for _, v := range regions {
			rect := gocv.BoundingRect(v)
			gocv.Rectangle(&original2, rect, color.RGBA{0, 255, 0, 255}, 2)
		}
		arr = append(arr, ix)
		utils.ShowImageInNamedWindow(original2, fmt.Sprintf("%d", ix))
		original2.Close()
	}

	fmt.Println("******************************")
	for ix, it := range arr {
		if ix%4 == 0 {
			fmt.Printf("\n")
		}
		fc := newIDLowQIntCoeffArr[it]
		fmt.Printf("{%d, %d, %d, %d}, ", fc.w1, fc.h1, fc.w2, fc.h2)
	}

	fmt.Println("******************************")
	for ix, it := range arr {
		if ix%4 == 0 {
			fmt.Printf("\n")
		}
		fc := newIDLowQFloatCoeffArr[it]
		fmt.Printf("{%f, %f, %f, %f}, ", fc.w1, fc.h1, fc.w2, fc.h2)
	}
	fmt.Println("******************************")
}
