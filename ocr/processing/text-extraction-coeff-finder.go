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
	{458, 374, 565, 391}, {458, 327, 571, 345}, {262, 319, 370, 336},
	{264, 282, 531, 300}, {264, 248, 322, 264}, {264, 214, 377, 232},
	{263, 184, 334, 200}, {263, 164, 340, 180}, {264, 133, 416, 151},
	{264, 113, 411, 131},
}

type extractTextRegionIntCoeff struct {
	w1, h1, w2, h2 int
}

type extractTextRegionFloatCoeff struct {
	w1, h1, w2, h2 float64
}

var newIDLowQFloatCoeffArr []extractTextRegionFloatCoeff = []extractTextRegionFloatCoeff{
	{0.016616, 0.007143, 0.037764, 0.011905}, {0.016616, 0.004762, 0.037764, 0.011905}, {0.016616, 0.004762, 0.037764, 0.009524}, {0.016616, 0.004762, 0.037764, 0.007143},
	{0.016616, 0.004762, 0.037764, 0.004762}, {0.016616, 0.004762, 0.036254, 0.011905}, {0.015106, 0.007143, 0.036254, 0.009524}, {0.015106, 0.007143, 0.036254, 0.007143},
	{0.015106, 0.007143, 0.036254, 0.004762}, {0.015106, 0.004762, 0.036254, 0.009524}, {0.015106, 0.004762, 0.036254, 0.007143}, {0.015106, 0.004762, 0.036254, 0.004762},
	{0.015106, 0.004762, 0.034743, 0.009524}, {0.015106, 0.004762, 0.034743, 0.007143}, {0.015106, 0.004762, 0.034743, 0.004762}, {0.015106, 0.004762, 0.033233, 0.009524},
	{0.015106, 0.004762, 0.033233, 0.007143}, {0.015106, 0.004762, 0.033233, 0.004762}, {0.015106, 0.004762, 0.031722, 0.009524}, {0.015106, 0.004762, 0.031722, 0.007143},
	{0.015106, 0.004762, 0.030211, 0.009524}, {0.015106, 0.004762, 0.030211, 0.007143}, {0.013595, 0.004762, 0.030211, 0.004762}, {0.013595, 0.004762, 0.028701, 0.009524},
	{0.013595, 0.004762, 0.028701, 0.007143}, {0.013595, 0.004762, 0.028701, 0.004762},
}

var newIDLowQIntCoeffArr []extractTextRegionIntCoeff = []extractTextRegionIntCoeff{
	{14, 3, 24, 2}, {14, 3, 22, 2}, {14, 3, 20, 2}, {14, 3, 18, 2},
	{14, 3, 16, 2}, {14, 2, 24, 2}, {14, 2, 22, 2}, {14, 2, 20, 2},
	{14, 2, 18, 2}, {14, 2, 16, 2}, {13, 3, 24, 2}, {13, 3, 22, 2},
	{13, 3, 20, 2}, {13, 3, 18, 2}, {13, 3, 16, 2}, {13, 3, 14, 2},
	{13, 2, 24, 2}, {13, 2, 22, 2}, {13, 2, 20, 2}, {13, 2, 18, 2},
	{13, 2, 16, 2}, {13, 2, 14, 2}, {12, 3, 25, 2}, {12, 3, 24, 2},
	{12, 3, 23, 2}, {12, 3, 22, 2}, {12, 3, 21, 2}, {12, 3, 20, 2},
	{12, 3, 19, 2}, {12, 3, 18, 2}, {12, 3, 17, 2}, {12, 3, 16, 2},
	{12, 3, 15, 2}, {12, 3, 14, 2}, {12, 3, 13, 3}, {12, 3, 11, 3},
	{12, 3, 9, 3}, {12, 3, 7, 3}, {12, 2, 25, 2}, {12, 2, 24, 2},
	{12, 2, 23, 2}, {12, 2, 22, 2}, {12, 2, 21, 2}, {12, 2, 20, 2},
	{12, 2, 19, 2}, {12, 2, 18, 2}, {12, 2, 17, 2}, {12, 2, 16, 2},
	{12, 2, 15, 2}, {12, 2, 14, 2}, {12, 2, 13, 2}, {12, 2, 12, 2},
	{12, 2, 11, 2}, {12, 2, 10, 2}, {12, 2, 9, 2}, {12, 2, 8, 2},
	{12, 2, 7, 2}, {12, 2, 6, 2}, {11, 3, 11, 2}, {11, 3, 10, 2},
	{11, 3, 9, 2}, {11, 3, 8, 2}, {11, 3, 7, 3}, {11, 3, 7, 2},
	{11, 3, 6, 2}, {11, 2, 11, 2}, {11, 2, 10, 2}, {11, 2, 9, 2},
	{11, 2, 8, 2}, {11, 2, 7, 2}, {11, 2, 6, 2}, {10, 3, 10, 2},
	{10, 3, 9, 2}, {10, 3, 8, 2}, {10, 3, 7, 2}, {10, 3, 6, 2},
	{10, 2, 10, 2}, {10, 2, 9, 2}, {10, 2, 8, 2}, {10, 2, 7, 2},
	{10, 2, 6, 2}, {9, 5, 20, 2}, {9, 5, 18, 4}, {9, 5, 18, 3},
	{9, 5, 18, 2}, {9, 5, 16, 4}, {9, 4, 20, 2}, {9, 4, 18, 4},
	{9, 4, 18, 3}, {9, 4, 18, 2}, {9, 4, 16, 4}, {9, 4, 16, 3},
	{9, 3, 9, 2}, {9, 3, 8, 2}, {9, 3, 7, 2}, {9, 3, 6, 2},
	{9, 2, 8, 2}, {9, 2, 7, 2}, {9, 2, 6, 2}, {8, 3, 8, 2},
	{8, 3, 6, 2}, {8, 2, 8, 2}, {8, 2, 7, 2}, {7, 3, 8, 2},
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
	const max = 25
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

					if !checkRegionsNewID(regions, newIDLowQRects, 5.0, 5.0) {
						continue
					}

					original2 := img.Clone()
					for _, v := range regions {
						rect := gocv.BoundingRect(v)
						gocv.Rectangle(original2, rect, color.RGBA{0, 255, 0, 255}, 1)
						// fmt.Printf("{%d, %d, %d, %d}, \n", rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y)
						// utils.ShowImageInNamedWindow(original2, fmt.Sprintf("{%d, %d, %d, %d}", rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y))
					}
					if index%10 == 0 {
						fmt.Printf("\n")
					}
					fmt.Printf("{%d, %d, %d, %d}, ", w, h, w2, h2)
					// utils.ShowImageInNamedWindow(original2, fmt.Sprintf("{%d, %d, %d, %d}", w, h, w2, h2))
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
	symbolHeight := symbolHeightCoeff * float64(img.Rows())
	symbolWidth := symbolWidthCoeff * float64(img.Cols())
	for ix, fc := range newIDLowQIntCoeffArr {
		w1c := float64(fc.w1) / symbolHeight
		h1c := float64(fc.h1) / symbolWidth
		w2c := float64(fc.w2) / symbolHeight
		h2c := float64(fc.h2) / symbolWidth

		if ix%10 == 0 {
			fmt.Printf("\n")
		}
		fmt.Printf("{%f, %f, %f, %f}, ", w1c, h1c, w2c, h2c)
	}
}

func showExampleRectangles(img gocv.Mat) {
	// original2 := img.Clone()
	// defer original2.Close()

	// for ix, r := range newIDLowQRects {
	// 	rect := image.Rectangle{image.Point{r.x0, r.y0}, image.Point{r.x1, r.y1}}
	// 	gocv.Rectangle(original2, rect, color.RGBA{0, 255, 0, 255}, 1)
	// 	fmt.Printf("{%d, %d, %d, %d}, ", r.x0, r.y0, r.x1, r.y1)
	// 	utils.ShowImageInNamedWindow(original2, fmt.Sprintf("%d", ix))
	// }

	// symbolHeight := symbolHeightCoeff * float64(img.Rows())
	// symbolWidth := symbolWidthCoeff * float64(img.Cols())
	for ix, fc := range newIDLowQIntCoeffArr {
		if ix%10 == 0 {
			fmt.Printf("\n")
		}
		w1c := float64(fc.w1) / float64(img.Cols())
		h1c := float64(fc.h1) / float64(img.Rows())

		w2c := float64(fc.w2) / float64(img.Cols())
		h2c := float64(fc.h2) / float64(img.Rows())

		fmt.Printf("{%f, %f, %f, %f}, ", w1c, h1c, w2c, h2c)
	}
}

func testCoefficientsForID(img gocv.Mat) {
	for ix, fc := range newIDLowQFloatCoeffArr {
		w1c := fc.w1 * float64(img.Cols())
		h1c := fc.h1 * float64(img.Rows())
		w2c := fc.w2 * float64(img.Cols())
		h2c := fc.h2 * float64(img.Rows())
		regions, err := textRegionsInternal(img, extractTextRegionIntCoeff{
			int(w1c), int(h1c), int(w2c), int(h2c)})

		if err != nil {
			continue
		}

		original2 := img.Clone()
		for _, v := range regions {
			rect := gocv.BoundingRect(v)
			gocv.Rectangle(original2, rect, color.RGBA{0, 255, 0, 255}, 2)
		}
		utils.ShowImageInNamedWindow(original2, fmt.Sprintf("%d", ix))
		original2.Close()
	}
}
