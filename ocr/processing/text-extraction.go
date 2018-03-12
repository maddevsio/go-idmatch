package processing

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/utils"
	"github.com/otiai10/gosseract"
	"gocv.io/x/gocv"
)

type block struct {
	x, y, h, w float64
	text       string
}

type frect struct {
	x0, y0, x1, y1 int
}

var newId2Arr []frect = []frect{
	{858, 710, 1054, 741},
	{861, 677, 1118, 697},
	{862, 656, 1035, 676},
	{489, 641, 715, 663},
	{859, 622, 1066, 655},
	{487, 607, 683, 638},
	{862, 593, 1094, 614},
	// {493, 573, 1013, 595}, //maybe need more interesting one !!!
	{489, 537, 993, 570},
	{490, 507, 826, 527},
	{489, 471, 592, 502},
	{491, 443, 661, 461},
	{490, 407, 697, 439},
	{490, 378, 810, 399},
	{488, 349, 614, 379},
	{488, 311, 626, 341},
	{490, 285, 639, 302},
	{488, 250, 774, 282},
	{489, 210, 764, 243},
	{488, 185, 769, 202},
}

const maxQualitySymWidth = 34.1
const maxQualityWidth = 1239.0
const maxQualitySymHeight = 37.0
const maxQualityHeight = 781.0
const symbolHeightCoeff = maxQualitySymHeight / maxQualityHeight
const symbolWidthCoeff = maxQualitySymWidth / maxQualityWidth

//

func compareRects(x00, y00, x01, y01, x10, y10, x11, y11 int, devX, devY float64) bool {
	return !(math.Abs(float64(x10-x00)) > devX ||
		math.Abs(float64(y10-y00)) > devY ||
		math.Abs(float64(x11-x01)) > devX ||
		math.Abs(float64(y11-y01)) > devY)
}

func checkRegionsNewID2(regions [][]image.Point) bool {
	const devX = 4.0
	const devY = 3.0
	count := 0
	for _, regIn := range regions {
		rect := gocv.BoundingRect(regIn)

		for _, regEt := range newId2Arr {
			if compareRects(rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y,
				regEt.x0, regEt.y0, regEt.x1, regEt.y1, devX, devY) {
				count++
			}
		}
	}
	return count == len(newId2Arr)
}

type extractTextRegionIntCoeff struct {
	w1, h1, w2, h2 int
}

type extractTextRegionFloatCoeff struct {
	w1, h1, w2, h2 float64
}

var newIdCoeffArr []extractTextRegionIntCoeff = []extractTextRegionIntCoeff{
	{18, 3, 16, 4}, {18, 3, 8, 7}, {18, 3, 8, 6}, {18, 3, 6, 7}, {18, 3, 6, 6},
	{18, 3, 4, 7}, {18, 3, 4, 6}, {18, 2, 16, 4}, {18, 2, 14, 4}, {18, 2, 12, 4},
	{18, 2, 10, 5}, {18, 2, 10, 4}, {18, 2, 8, 6}, {18, 2, 8, 5}, {18, 2, 8, 4},
	{18, 2, 6, 6}, {18, 2, 6, 5}, {18, 2, 6, 4}, {18, 2, 4, 7}, {18, 2, 4, 6},
	{18, 2, 4, 5}, {18, 2, 4, 4}, {17, 3, 16, 4}, {17, 3, 14, 4}, {17, 3, 12, 5},
	{17, 3, 12, 4}, {17, 3, 10, 5}, {17, 3, 10, 4}, {17, 3, 8, 7}, {17, 3, 8, 6},
	{17, 3, 8, 5}, {17, 3, 8, 4}, {17, 3, 6, 7}, {17, 3, 6, 6}, {17, 3, 6, 5},
	{17, 3, 6, 4}, {17, 3, 4, 7}, {17, 3, 4, 6}, {17, 3, 4, 5}, {17, 3, 4, 4},
	{17, 2, 16, 4}, {17, 2, 14, 4}, {17, 2, 12, 4}, {17, 2, 10, 5}, {17, 2, 10, 4},
	{17, 2, 8, 6}, {17, 2, 8, 5}, {17, 2, 8, 4}, {17, 2, 6, 6}, {17, 2, 6, 5},
	{17, 2, 6, 4}, {17, 2, 4, 6}, {17, 2, 4, 5}, {17, 2, 4, 4}, {16, 3, 15, 4},
	{16, 3, 14, 4}, {16, 3, 13, 4}, {16, 3, 12, 5}, {16, 3, 12, 4}, {16, 3, 11, 5},
	{16, 3, 11, 4}, {16, 3, 10, 5}, {16, 3, 10, 4}, {16, 3, 9, 5}, {16, 3, 9, 4},
	{16, 3, 8, 6}, {16, 3, 8, 5}, {16, 3, 8, 4}, {16, 3, 7, 7}, {16, 3, 7, 6},
	{16, 3, 7, 5}, {16, 3, 7, 4}, {16, 3, 6, 6}, {16, 3, 6, 5}, {16, 3, 6, 4},
	{16, 3, 5, 7}, {16, 3, 5, 6}, {16, 3, 5, 5}, {16, 3, 5, 4}, {16, 3, 4, 6},
	{16, 3, 4, 5}, {16, 3, 4, 4}, {16, 2, 15, 4}, {16, 2, 14, 4}, {16, 2, 13, 4},
	{16, 2, 12, 4}, {16, 2, 11, 4}, {16, 2, 10, 4}, {16, 2, 9, 5}, {16, 2, 9, 4},
	{16, 2, 8, 5}, {16, 2, 8, 4}, {16, 2, 7, 6}, {16, 2, 7, 5}, {16, 2, 7, 4},
	{16, 2, 6, 5}, {16, 2, 6, 4}, {16, 2, 5, 6}, {16, 2, 5, 5}, {16, 2, 5, 4},
	{16, 2, 4, 6}, {16, 2, 4, 5}, {16, 2, 4, 4}, {15, 3, 8, 6}, {15, 3, 7, 6},
	{15, 3, 6, 6}, {15, 3, 5, 7}, {15, 3, 5, 6}, {15, 2, 14, 4}, {15, 2, 13, 4},
	{15, 2, 12, 4}, {15, 2, 11, 5}, {15, 2, 11, 4}, {15, 2, 10, 5}, {15, 2, 10, 4},
	{15, 2, 9, 5}, {15, 2, 9, 4}, {15, 2, 8, 6}, {15, 2, 8, 5}, {15, 2, 8, 4},
	{15, 2, 7, 6}, {15, 2, 7, 5}, {15, 2, 7, 4}, {15, 2, 6, 6}, {15, 2, 6, 5},
	{15, 2, 6, 4}, {15, 2, 5, 7}, {15, 2, 5, 6}, {15, 2, 5, 5}, {15, 2, 5, 4},
	{14, 3, 10, 5}, {14, 3, 9, 5}, {14, 3, 8, 5}, {14, 3, 7, 6}, {14, 3, 7, 5},
	{14, 3, 6, 6}, {14, 3, 6, 5}, {14, 3, 5, 6}, {14, 3, 5, 5}, {14, 3, 4, 6},
	{14, 2, 13, 4}, {14, 2, 13, 3}, {14, 2, 12, 4}, {14, 2, 12, 3}, {14, 2, 11, 4},
	{14, 2, 11, 3}, {14, 2, 10, 4}, {14, 2, 10, 3}, {14, 2, 9, 4}, {14, 2, 9, 3},
	{14, 2, 8, 4}, {14, 2, 8, 3}, {14, 2, 7, 5}, {14, 2, 7, 4}, {14, 2, 7, 3},
	{14, 2, 6, 5}, {14, 2, 6, 4}, {14, 2, 6, 3}, {14, 2, 5, 5}, {14, 2, 5, 4},
	{14, 2, 5, 3}, {13, 3, 9, 5}, {13, 3, 8, 5}, {13, 3, 7, 5}, {13, 3, 6, 6},
	{13, 3, 5, 6}, {13, 3, 4, 6}, {13, 2, 12, 4}, {13, 2, 12, 3}, {13, 2, 11, 4},
	{13, 2, 11, 3}, {13, 2, 10, 4}, {13, 2, 10, 3}, {13, 2, 9, 4}, {13, 2, 9, 3},
	{13, 2, 8, 5}, {13, 2, 8, 4}, {13, 2, 8, 3}, {13, 2, 7, 5}, {13, 2, 7, 4},
	{13, 2, 7, 3}, {13, 2, 6, 5}, {13, 2, 5, 6}, {13, 2, 5, 5}, {13, 2, 4, 6},
	{13, 2, 4, 5}, {12, 3, 11, 4}, {12, 3, 11, 3}, {12, 3, 10, 4}, {12, 3, 10, 3},
	{12, 3, 9, 4}, {12, 3, 9, 3}, {12, 3, 8, 5}, {12, 3, 8, 4}, {12, 3, 8, 3},
	{12, 3, 7, 5}, {12, 3, 7, 4}, {12, 3, 7, 3}, {12, 3, 6, 6}, {12, 3, 6, 5},
	{12, 3, 6, 4}, {12, 3, 6, 3}, {12, 3, 5, 6}, {12, 3, 5, 5}, {12, 3, 5, 4},
	{12, 3, 5, 3}, {12, 3, 4, 6}, {12, 3, 4, 5}, {12, 3, 4, 4}, {12, 3, 4, 3},
	{12, 2, 11, 4}, {12, 2, 11, 3}, {12, 2, 10, 4}, {12, 2, 10, 3}, {12, 2, 9, 4},
	{12, 2, 9, 3}, {12, 2, 8, 4}, {12, 2, 8, 3}, {12, 2, 7, 5}, {12, 2, 7, 4},
	{12, 2, 7, 3}, {12, 2, 6, 5}, {12, 2, 6, 4}, {12, 2, 6, 3}, {12, 2, 5, 6},
	{12, 2, 5, 5}, {12, 2, 5, 4}, {12, 2, 5, 3}, {12, 2, 4, 6}, {12, 2, 4, 5},
	{12, 2, 4, 4}, {12, 2, 4, 3}, {11, 3, 11, 2}, {11, 3, 10, 4}, {11, 3, 10, 3},
	{11, 3, 9, 4}, {11, 3, 9, 3}, {11, 3, 9, 2}, {11, 3, 8, 4}, {11, 3, 8, 3},
	{11, 3, 7, 5}, {11, 3, 7, 4}, {11, 3, 7, 3}, {11, 3, 7, 2}, {11, 3, 6, 5},
	{11, 3, 6, 4}, {11, 3, 6, 3}, {11, 3, 5, 6}, {11, 3, 5, 5}, {11, 3, 5, 4},
	{11, 3, 5, 3}, {11, 3, 5, 2}, {11, 3, 4, 6}, {11, 3, 4, 5}, {11, 3, 4, 4},
	{11, 3, 4, 3}, {11, 2, 10, 4}, {11, 2, 10, 3}, {11, 2, 9, 4}, {11, 2, 9, 3},
	{11, 2, 9, 2}, {11, 2, 8, 4}, {11, 2, 8, 3}, {11, 2, 7, 4}, {11, 2, 7, 3},
	{11, 2, 7, 2}, {11, 2, 6, 4}, {11, 2, 6, 3}, {11, 2, 5, 5}, {11, 2, 5, 4},
	{11, 2, 5, 3}, {11, 2, 5, 2}, {11, 2, 4, 5}, {11, 2, 4, 4}, {11, 2, 4, 3},
	{10, 3, 11, 2}, {10, 3, 9, 4}, {10, 3, 9, 3}, {10, 3, 9, 2}, {10, 3, 8, 4},
	{10, 3, 8, 3}, {10, 3, 7, 5}, {10, 3, 7, 4}, {10, 3, 7, 3}, {10, 3, 7, 2},
	{10, 3, 6, 5}, {10, 3, 6, 4}, {10, 3, 6, 3}, {10, 3, 5, 6}, {10, 3, 5, 5},
	{10, 3, 5, 4}, {10, 3, 5, 3}, {10, 3, 5, 2}, {10, 3, 4, 6}, {10, 3, 4, 5},
	{10, 3, 4, 4}, {10, 3, 4, 3}, {10, 2, 9, 4}, {10, 2, 9, 3}, {10, 2, 9, 2},
	{10, 2, 8, 4}, {10, 2, 8, 3}, {10, 2, 7, 4}, {10, 2, 7, 3}, {10, 2, 7, 2},
	{10, 2, 6, 4}, {10, 2, 6, 3}, {10, 2, 5, 5}, {10, 2, 5, 4}, {10, 2, 5, 3},
	{10, 2, 5, 2}, {10, 2, 4, 5}, {10, 2, 4, 4}, {10, 2, 4, 3}, {9, 3, 10, 2},
	{9, 3, 9, 2}, {9, 3, 8, 4}, {9, 3, 8, 3}, {9, 3, 8, 2}, {9, 3, 7, 4},
	{9, 3, 7, 3}, {9, 3, 7, 2}, {9, 3, 6, 4}, {9, 3, 6, 3}, {9, 3, 6, 2},
	{9, 3, 5, 5}, {9, 3, 5, 4}, {9, 3, 5, 3}, {9, 3, 5, 2}, {9, 3, 4, 6},
	{9, 3, 4, 5}, {9, 3, 4, 4}, {9, 3, 4, 3}, {9, 3, 4, 2}, {9, 2, 8, 4},
	{9, 2, 8, 3}, {9, 2, 8, 2}, {9, 2, 7, 4}, {9, 2, 7, 3}, {9, 2, 7, 2},
	{9, 2, 6, 4}, {9, 2, 6, 3}, {9, 2, 6, 2}, {9, 2, 5, 5}, {9, 2, 5, 4},
	{9, 2, 5, 3}, {9, 2, 5, 2}, {9, 2, 4, 6}, {9, 2, 4, 5}, {9, 2, 4, 4},
	{9, 2, 4, 3}, {9, 2, 4, 2}, {8, 3, 9, 2}, {8, 3, 8, 2}, {8, 3, 7, 4},
	{8, 3, 7, 3}, {8, 3, 7, 2}, {8, 3, 6, 4}, {8, 3, 6, 3}, {8, 3, 6, 2},
	{8, 3, 5, 4}, {8, 3, 5, 3}, {8, 3, 5, 2}, {8, 3, 4, 6}, {8, 3, 4, 5},
	{8, 3, 4, 4}, {8, 3, 4, 3}, {8, 3, 4, 2}, {8, 2, 7, 4}, {8, 2, 7, 3},
	{8, 2, 7, 2}, {8, 2, 6, 4}, {8, 2, 6, 3}, {8, 2, 6, 2}, {8, 2, 5, 4},
	{8, 2, 5, 3}, {8, 2, 5, 2}, {8, 2, 4, 5}, {8, 2, 4, 4}, {8, 2, 4, 3},
	{8, 2, 4, 2}, {7, 3, 8, 2}, {7, 3, 7, 2}, {7, 3, 6, 4}, {7, 3, 6, 3},
	{7, 3, 6, 2}, {7, 3, 5, 4}, {7, 3, 5, 3}, {7, 3, 5, 2}, {7, 3, 4, 4},
	{7, 3, 4, 3}, {7, 3, 4, 2}, {7, 2, 6, 4}, {7, 2, 6, 3}, {7, 2, 6, 2},
	{7, 2, 5, 4}, {7, 2, 5, 3}, {7, 2, 5, 2}, {7, 2, 4, 4}, {7, 2, 4, 3},
	{7, 2, 4, 2}, {4, 2, 3, 3},
}

var newIdFloatCoeffArr []extractTextRegionFloatCoeff = []extractTextRegionFloatCoeff{
	{0.014400, 0.003812, 0.012800, 0.005083},
	{0.014400, 0.003812, 0.006400, 0.008895}, {0.014400, 0.003812, 0.006400, 0.007624},
	{0.014400, 0.003812, 0.004800, 0.008895}, {0.014400, 0.003812, 0.004800, 0.007624},
	{0.014400, 0.003812, 0.003200, 0.008895}, {0.014400, 0.003812, 0.003200, 0.007624},
	{0.014400, 0.002541, 0.012800, 0.005083}, {0.014400, 0.002541, 0.011200, 0.005083},
	{0.014400, 0.002541, 0.009600, 0.005083}, {0.014400, 0.002541, 0.008000, 0.006353},
	{0.014400, 0.002541, 0.008000, 0.005083}, {0.014400, 0.002541, 0.006400, 0.007624},
	{0.014400, 0.002541, 0.006400, 0.006353}, {0.014400, 0.002541, 0.006400, 0.005083},
	{0.014400, 0.002541, 0.004800, 0.007624}, {0.014400, 0.002541, 0.004800, 0.006353},
	{0.014400, 0.002541, 0.004800, 0.005083}, {0.014400, 0.002541, 0.003200, 0.008895},
	{0.014400, 0.002541, 0.003200, 0.007624}, {0.014400, 0.002541, 0.003200, 0.006353},
	{0.014400, 0.002541, 0.003200, 0.005083}, {0.013600, 0.003812, 0.012800, 0.005083},
	{0.013600, 0.003812, 0.011200, 0.005083}, {0.013600, 0.003812, 0.009600, 0.006353},
	{0.013600, 0.003812, 0.009600, 0.005083}, {0.013600, 0.003812, 0.008000, 0.006353},
	{0.013600, 0.003812, 0.008000, 0.005083}, {0.013600, 0.003812, 0.006400, 0.008895},
	{0.013600, 0.003812, 0.006400, 0.007624}, {0.013600, 0.003812, 0.006400, 0.006353},
	{0.013600, 0.003812, 0.006400, 0.005083}, {0.013600, 0.003812, 0.004800, 0.008895},
	{0.013600, 0.003812, 0.004800, 0.007624}, {0.013600, 0.003812, 0.004800, 0.006353},
	{0.013600, 0.003812, 0.004800, 0.005083}, {0.013600, 0.003812, 0.003200, 0.008895},
	{0.013600, 0.003812, 0.003200, 0.007624}, {0.013600, 0.003812, 0.003200, 0.006353},
	{0.013600, 0.003812, 0.003200, 0.005083}, {0.013600, 0.002541, 0.012800, 0.005083},
	{0.013600, 0.002541, 0.011200, 0.005083}, {0.013600, 0.002541, 0.009600, 0.005083},
	{0.013600, 0.002541, 0.008000, 0.006353}, {0.013600, 0.002541, 0.008000, 0.005083},
	{0.013600, 0.002541, 0.006400, 0.007624}, {0.013600, 0.002541, 0.006400, 0.006353},
	{0.013600, 0.002541, 0.006400, 0.005083}, {0.013600, 0.002541, 0.004800, 0.007624},
	{0.013600, 0.002541, 0.004800, 0.006353}, {0.013600, 0.002541, 0.004800, 0.005083},
	{0.013600, 0.002541, 0.003200, 0.007624}, {0.013600, 0.002541, 0.003200, 0.006353},
	{0.013600, 0.002541, 0.003200, 0.005083}, {0.012800, 0.003812, 0.012000, 0.005083},
	{0.012800, 0.003812, 0.011200, 0.005083}, {0.012800, 0.003812, 0.010400, 0.005083},
	{0.012800, 0.003812, 0.009600, 0.006353}, {0.012800, 0.003812, 0.009600, 0.005083},
	{0.012800, 0.003812, 0.008800, 0.006353}, {0.012800, 0.003812, 0.008800, 0.005083},
	{0.012800, 0.003812, 0.008000, 0.006353}, {0.012800, 0.003812, 0.008000, 0.005083},
	{0.012800, 0.003812, 0.007200, 0.006353}, {0.012800, 0.003812, 0.007200, 0.005083},
	{0.012800, 0.003812, 0.006400, 0.007624}, {0.012800, 0.003812, 0.006400, 0.006353},
	{0.012800, 0.003812, 0.006400, 0.005083}, {0.012800, 0.003812, 0.005600, 0.008895},
	{0.012800, 0.003812, 0.005600, 0.007624}, {0.012800, 0.003812, 0.005600, 0.006353},
	{0.012800, 0.003812, 0.005600, 0.005083}, {0.012800, 0.003812, 0.004800, 0.007624},
	{0.012800, 0.003812, 0.004800, 0.006353}, {0.012800, 0.003812, 0.004800, 0.005083},
	{0.012800, 0.003812, 0.004000, 0.008895}, {0.012800, 0.003812, 0.004000, 0.007624},
	{0.012800, 0.003812, 0.004000, 0.006353}, {0.012800, 0.003812, 0.004000, 0.005083},
	{0.012800, 0.003812, 0.003200, 0.007624}, {0.012800, 0.003812, 0.003200, 0.006353},
	{0.012800, 0.003812, 0.003200, 0.005083}, {0.012800, 0.002541, 0.012000, 0.005083},
	{0.012800, 0.002541, 0.011200, 0.005083}, {0.012800, 0.002541, 0.010400, 0.005083},
	{0.012800, 0.002541, 0.009600, 0.005083}, {0.012800, 0.002541, 0.008800, 0.005083},
	{0.012800, 0.002541, 0.008000, 0.005083}, {0.012800, 0.002541, 0.007200, 0.006353},
	{0.012800, 0.002541, 0.007200, 0.005083}, {0.012800, 0.002541, 0.006400, 0.006353},
	{0.012800, 0.002541, 0.006400, 0.005083}, {0.012800, 0.002541, 0.005600, 0.007624},
	{0.012800, 0.002541, 0.005600, 0.006353}, {0.012800, 0.002541, 0.005600, 0.005083},
	{0.012800, 0.002541, 0.004800, 0.006353}, {0.012800, 0.002541, 0.004800, 0.005083},
	{0.012800, 0.002541, 0.004000, 0.007624}, {0.012800, 0.002541, 0.004000, 0.006353},
	{0.012800, 0.002541, 0.004000, 0.005083}, {0.012800, 0.002541, 0.003200, 0.007624},
	{0.012800, 0.002541, 0.003200, 0.006353}, {0.012800, 0.002541, 0.003200, 0.005083},
	{0.012000, 0.003812, 0.006400, 0.007624}, {0.012000, 0.003812, 0.005600, 0.007624},
	{0.012000, 0.003812, 0.004800, 0.007624}, {0.012000, 0.003812, 0.004000, 0.008895},
	{0.012000, 0.003812, 0.004000, 0.007624}, {0.012000, 0.002541, 0.011200, 0.005083},
	{0.012000, 0.002541, 0.010400, 0.005083}, {0.012000, 0.002541, 0.009600, 0.005083},
	{0.012000, 0.002541, 0.008800, 0.006353}, {0.012000, 0.002541, 0.008800, 0.005083},
	{0.012000, 0.002541, 0.008000, 0.006353}, {0.012000, 0.002541, 0.008000, 0.005083},
	{0.012000, 0.002541, 0.007200, 0.006353}, {0.012000, 0.002541, 0.007200, 0.005083},
	{0.012000, 0.002541, 0.006400, 0.007624}, {0.012000, 0.002541, 0.006400, 0.006353},
	{0.012000, 0.002541, 0.006400, 0.005083}, {0.012000, 0.002541, 0.005600, 0.007624},
	{0.012000, 0.002541, 0.005600, 0.006353}, {0.012000, 0.002541, 0.005600, 0.005083},
	{0.012000, 0.002541, 0.004800, 0.007624}, {0.012000, 0.002541, 0.004800, 0.006353},
	{0.012000, 0.002541, 0.004800, 0.005083}, {0.012000, 0.002541, 0.004000, 0.008895},
	{0.012000, 0.002541, 0.004000, 0.007624}, {0.012000, 0.002541, 0.004000, 0.006353},
	{0.012000, 0.002541, 0.004000, 0.005083}, {0.011200, 0.003812, 0.008000, 0.006353},
	{0.011200, 0.003812, 0.007200, 0.006353}, {0.011200, 0.003812, 0.006400, 0.006353},
	{0.011200, 0.003812, 0.005600, 0.007624}, {0.011200, 0.003812, 0.005600, 0.006353},
	{0.011200, 0.003812, 0.004800, 0.007624}, {0.011200, 0.003812, 0.004800, 0.006353},
	{0.011200, 0.003812, 0.004000, 0.007624}, {0.011200, 0.003812, 0.004000, 0.006353},
	{0.011200, 0.003812, 0.003200, 0.007624}, {0.011200, 0.002541, 0.010400, 0.005083},
	{0.011200, 0.002541, 0.010400, 0.003812}, {0.011200, 0.002541, 0.009600, 0.005083},
	{0.011200, 0.002541, 0.009600, 0.003812}, {0.011200, 0.002541, 0.008800, 0.005083},
	{0.011200, 0.002541, 0.008800, 0.003812}, {0.011200, 0.002541, 0.008000, 0.005083},
	{0.011200, 0.002541, 0.008000, 0.003812}, {0.011200, 0.002541, 0.007200, 0.005083},
	{0.011200, 0.002541, 0.007200, 0.003812}, {0.011200, 0.002541, 0.006400, 0.005083},
	{0.011200, 0.002541, 0.006400, 0.003812}, {0.011200, 0.002541, 0.005600, 0.006353},
	{0.011200, 0.002541, 0.005600, 0.005083}, {0.011200, 0.002541, 0.005600, 0.003812},
	{0.011200, 0.002541, 0.004800, 0.006353}, {0.011200, 0.002541, 0.004800, 0.005083},
	{0.011200, 0.002541, 0.004800, 0.003812}, {0.011200, 0.002541, 0.004000, 0.006353},
	{0.011200, 0.002541, 0.004000, 0.005083}, {0.011200, 0.002541, 0.004000, 0.003812},
	{0.010400, 0.003812, 0.007200, 0.006353}, {0.010400, 0.003812, 0.006400, 0.006353},
	{0.010400, 0.003812, 0.005600, 0.006353}, {0.010400, 0.003812, 0.004800, 0.007624},
	{0.010400, 0.003812, 0.004000, 0.007624}, {0.010400, 0.003812, 0.003200, 0.007624},
	{0.010400, 0.002541, 0.009600, 0.005083}, {0.010400, 0.002541, 0.009600, 0.003812},
	{0.010400, 0.002541, 0.008800, 0.005083}, {0.010400, 0.002541, 0.008800, 0.003812},
	{0.010400, 0.002541, 0.008000, 0.005083}, {0.010400, 0.002541, 0.008000, 0.003812},
	{0.010400, 0.002541, 0.007200, 0.005083}, {0.010400, 0.002541, 0.007200, 0.003812},
	{0.010400, 0.002541, 0.006400, 0.006353}, {0.010400, 0.002541, 0.006400, 0.005083},
	{0.010400, 0.002541, 0.006400, 0.003812}, {0.010400, 0.002541, 0.005600, 0.006353},
	{0.010400, 0.002541, 0.005600, 0.005083}, {0.010400, 0.002541, 0.005600, 0.003812},
	{0.010400, 0.002541, 0.004800, 0.006353}, {0.010400, 0.002541, 0.004000, 0.007624},
	{0.010400, 0.002541, 0.004000, 0.006353}, {0.010400, 0.002541, 0.003200, 0.007624},
	{0.010400, 0.002541, 0.003200, 0.006353}, {0.009600, 0.003812, 0.008800, 0.005083},
	{0.009600, 0.003812, 0.008800, 0.003812}, {0.009600, 0.003812, 0.008000, 0.005083},
	{0.009600, 0.003812, 0.008000, 0.003812}, {0.009600, 0.003812, 0.007200, 0.005083},
	{0.009600, 0.003812, 0.007200, 0.003812}, {0.009600, 0.003812, 0.006400, 0.006353},
	{0.009600, 0.003812, 0.006400, 0.005083}, {0.009600, 0.003812, 0.006400, 0.003812},
	{0.009600, 0.003812, 0.005600, 0.006353}, {0.009600, 0.003812, 0.005600, 0.005083},
	{0.009600, 0.003812, 0.005600, 0.003812}, {0.009600, 0.003812, 0.004800, 0.007624},
	{0.009600, 0.003812, 0.004800, 0.006353}, {0.009600, 0.003812, 0.004800, 0.005083},
	{0.009600, 0.003812, 0.004800, 0.003812}, {0.009600, 0.003812, 0.004000, 0.007624},
	{0.009600, 0.003812, 0.004000, 0.006353}, {0.009600, 0.003812, 0.004000, 0.005083},
	{0.009600, 0.003812, 0.004000, 0.003812}, {0.009600, 0.003812, 0.003200, 0.007624},
	{0.009600, 0.003812, 0.003200, 0.006353}, {0.009600, 0.003812, 0.003200, 0.005083},
	{0.009600, 0.003812, 0.003200, 0.003812}, {0.009600, 0.002541, 0.008800, 0.005083},
	{0.009600, 0.002541, 0.008800, 0.003812}, {0.009600, 0.002541, 0.008000, 0.005083},
	{0.009600, 0.002541, 0.008000, 0.003812}, {0.009600, 0.002541, 0.007200, 0.005083},
	{0.009600, 0.002541, 0.007200, 0.003812}, {0.009600, 0.002541, 0.006400, 0.005083},
	{0.009600, 0.002541, 0.006400, 0.003812}, {0.009600, 0.002541, 0.005600, 0.006353},
	{0.009600, 0.002541, 0.005600, 0.005083}, {0.009600, 0.002541, 0.005600, 0.003812},
	{0.009600, 0.002541, 0.004800, 0.006353}, {0.009600, 0.002541, 0.004800, 0.005083},
	{0.009600, 0.002541, 0.004800, 0.003812}, {0.009600, 0.002541, 0.004000, 0.007624},
	{0.009600, 0.002541, 0.004000, 0.006353}, {0.009600, 0.002541, 0.004000, 0.005083},
	{0.009600, 0.002541, 0.004000, 0.003812}, {0.009600, 0.002541, 0.003200, 0.007624},
	{0.009600, 0.002541, 0.003200, 0.006353}, {0.009600, 0.002541, 0.003200, 0.005083},
	{0.009600, 0.002541, 0.003200, 0.003812}, {0.008800, 0.003812, 0.008800, 0.002541},
	{0.008800, 0.003812, 0.008000, 0.005083}, {0.008800, 0.003812, 0.008000, 0.003812},
	{0.008800, 0.003812, 0.007200, 0.005083}, {0.008800, 0.003812, 0.007200, 0.003812},
	{0.008800, 0.003812, 0.007200, 0.002541}, {0.008800, 0.003812, 0.006400, 0.005083},
	{0.008800, 0.003812, 0.006400, 0.003812}, {0.008800, 0.003812, 0.005600, 0.006353},
	{0.008800, 0.003812, 0.005600, 0.005083}, {0.008800, 0.003812, 0.005600, 0.003812},
	{0.008800, 0.003812, 0.005600, 0.002541}, {0.008800, 0.003812, 0.004800, 0.006353},
	{0.008800, 0.003812, 0.004800, 0.005083}, {0.008800, 0.003812, 0.004800, 0.003812},
	{0.008800, 0.003812, 0.004000, 0.007624}, {0.008800, 0.003812, 0.004000, 0.006353},
	{0.008800, 0.003812, 0.004000, 0.005083}, {0.008800, 0.003812, 0.004000, 0.003812},
	{0.008800, 0.003812, 0.004000, 0.002541}, {0.008800, 0.003812, 0.003200, 0.007624},
	{0.008800, 0.003812, 0.003200, 0.006353}, {0.008800, 0.003812, 0.003200, 0.005083},
	{0.008800, 0.003812, 0.003200, 0.003812}, {0.008800, 0.002541, 0.008000, 0.005083},
	{0.008800, 0.002541, 0.008000, 0.003812}, {0.008800, 0.002541, 0.007200, 0.005083},
	{0.008800, 0.002541, 0.007200, 0.003812}, {0.008800, 0.002541, 0.007200, 0.002541},
	{0.008800, 0.002541, 0.006400, 0.005083}, {0.008800, 0.002541, 0.006400, 0.003812},
	{0.008800, 0.002541, 0.005600, 0.005083}, {0.008800, 0.002541, 0.005600, 0.003812},
	{0.008800, 0.002541, 0.005600, 0.002541}, {0.008800, 0.002541, 0.004800, 0.005083},
	{0.008800, 0.002541, 0.004800, 0.003812}, {0.008800, 0.002541, 0.004000, 0.006353},
	{0.008800, 0.002541, 0.004000, 0.005083}, {0.008800, 0.002541, 0.004000, 0.003812},
	{0.008800, 0.002541, 0.004000, 0.002541}, {0.008800, 0.002541, 0.003200, 0.006353},
	{0.008800, 0.002541, 0.003200, 0.005083}, {0.008800, 0.002541, 0.003200, 0.003812},
	{0.008000, 0.003812, 0.008800, 0.002541}, {0.008000, 0.003812, 0.007200, 0.005083},
	{0.008000, 0.003812, 0.007200, 0.003812}, {0.008000, 0.003812, 0.007200, 0.002541},
	{0.008000, 0.003812, 0.006400, 0.005083}, {0.008000, 0.003812, 0.006400, 0.003812},
	{0.008000, 0.003812, 0.005600, 0.006353}, {0.008000, 0.003812, 0.005600, 0.005083},
	{0.008000, 0.003812, 0.005600, 0.003812}, {0.008000, 0.003812, 0.005600, 0.002541},
	{0.008000, 0.003812, 0.004800, 0.006353}, {0.008000, 0.003812, 0.004800, 0.005083},
	{0.008000, 0.003812, 0.004800, 0.003812}, {0.008000, 0.003812, 0.004000, 0.007624},
	{0.008000, 0.003812, 0.004000, 0.006353}, {0.008000, 0.003812, 0.004000, 0.005083},
	{0.008000, 0.003812, 0.004000, 0.003812}, {0.008000, 0.003812, 0.004000, 0.002541},
	{0.008000, 0.003812, 0.003200, 0.007624}, {0.008000, 0.003812, 0.003200, 0.006353},
	{0.008000, 0.003812, 0.003200, 0.005083}, {0.008000, 0.003812, 0.003200, 0.003812},
	{0.008000, 0.002541, 0.007200, 0.005083}, {0.008000, 0.002541, 0.007200, 0.003812},
	{0.008000, 0.002541, 0.007200, 0.002541}, {0.008000, 0.002541, 0.006400, 0.005083},
	{0.008000, 0.002541, 0.006400, 0.003812}, {0.008000, 0.002541, 0.005600, 0.005083},
	{0.008000, 0.002541, 0.005600, 0.003812}, {0.008000, 0.002541, 0.005600, 0.002541},
	{0.008000, 0.002541, 0.004800, 0.005083}, {0.008000, 0.002541, 0.004800, 0.003812},
	{0.008000, 0.002541, 0.004000, 0.006353}, {0.008000, 0.002541, 0.004000, 0.005083},
	{0.008000, 0.002541, 0.004000, 0.003812}, {0.008000, 0.002541, 0.004000, 0.002541},
	{0.008000, 0.002541, 0.003200, 0.006353}, {0.008000, 0.002541, 0.003200, 0.005083},
	{0.008000, 0.002541, 0.003200, 0.003812}, {0.007200, 0.003812, 0.008000, 0.002541},
	{0.007200, 0.003812, 0.007200, 0.002541}, {0.007200, 0.003812, 0.006400, 0.005083},
	{0.007200, 0.003812, 0.006400, 0.003812}, {0.007200, 0.003812, 0.006400, 0.002541},
	{0.007200, 0.003812, 0.005600, 0.005083}, {0.007200, 0.003812, 0.005600, 0.003812},
	{0.007200, 0.003812, 0.005600, 0.002541}, {0.007200, 0.003812, 0.004800, 0.005083},
	{0.007200, 0.003812, 0.004800, 0.003812}, {0.007200, 0.003812, 0.004800, 0.002541},
	{0.007200, 0.003812, 0.004000, 0.006353}, {0.007200, 0.003812, 0.004000, 0.005083},
	{0.007200, 0.003812, 0.004000, 0.003812}, {0.007200, 0.003812, 0.004000, 0.002541},
	{0.007200, 0.003812, 0.003200, 0.007624}, {0.007200, 0.003812, 0.003200, 0.006353},
	{0.007200, 0.003812, 0.003200, 0.005083}, {0.007200, 0.003812, 0.003200, 0.003812},
	{0.007200, 0.003812, 0.003200, 0.002541}, {0.007200, 0.002541, 0.006400, 0.005083},
	{0.007200, 0.002541, 0.006400, 0.003812}, {0.007200, 0.002541, 0.006400, 0.002541},
	{0.007200, 0.002541, 0.005600, 0.005083}, {0.007200, 0.002541, 0.005600, 0.003812},
	{0.007200, 0.002541, 0.005600, 0.002541}, {0.007200, 0.002541, 0.004800, 0.005083},
	{0.007200, 0.002541, 0.004800, 0.003812}, {0.007200, 0.002541, 0.004800, 0.002541},
	{0.007200, 0.002541, 0.004000, 0.006353}, {0.007200, 0.002541, 0.004000, 0.005083},
	{0.007200, 0.002541, 0.004000, 0.003812}, {0.007200, 0.002541, 0.004000, 0.002541},
	{0.007200, 0.002541, 0.003200, 0.007624}, {0.007200, 0.002541, 0.003200, 0.006353},
	{0.007200, 0.002541, 0.003200, 0.005083}, {0.007200, 0.002541, 0.003200, 0.003812},
	{0.007200, 0.002541, 0.003200, 0.002541}, {0.006400, 0.003812, 0.007200, 0.002541},
	{0.006400, 0.003812, 0.006400, 0.002541}, {0.006400, 0.003812, 0.005600, 0.005083},
	{0.006400, 0.003812, 0.005600, 0.003812}, {0.006400, 0.003812, 0.005600, 0.002541},
	{0.006400, 0.003812, 0.004800, 0.005083}, {0.006400, 0.003812, 0.004800, 0.003812},
	{0.006400, 0.003812, 0.004800, 0.002541}, {0.006400, 0.003812, 0.004000, 0.005083},
	{0.006400, 0.003812, 0.004000, 0.003812}, {0.006400, 0.003812, 0.004000, 0.002541},
	{0.006400, 0.003812, 0.003200, 0.007624}, {0.006400, 0.003812, 0.003200, 0.006353},
	{0.006400, 0.003812, 0.003200, 0.005083}, {0.006400, 0.003812, 0.003200, 0.003812},
	{0.006400, 0.003812, 0.003200, 0.002541}, {0.006400, 0.002541, 0.005600, 0.005083},
	{0.006400, 0.002541, 0.005600, 0.003812}, {0.006400, 0.002541, 0.005600, 0.002541},
	{0.006400, 0.002541, 0.004800, 0.005083}, {0.006400, 0.002541, 0.004800, 0.003812},
	{0.006400, 0.002541, 0.004800, 0.002541}, {0.006400, 0.002541, 0.004000, 0.005083},
	{0.006400, 0.002541, 0.004000, 0.003812}, {0.006400, 0.002541, 0.004000, 0.002541},
	{0.006400, 0.002541, 0.003200, 0.006353}, {0.006400, 0.002541, 0.003200, 0.005083},
	{0.006400, 0.002541, 0.003200, 0.003812}, {0.006400, 0.002541, 0.003200, 0.002541},
	{0.005600, 0.003812, 0.006400, 0.002541}, {0.005600, 0.003812, 0.005600, 0.002541},
	{0.005600, 0.003812, 0.004800, 0.005083}, {0.005600, 0.003812, 0.004800, 0.003812},
	{0.005600, 0.003812, 0.004800, 0.002541}, {0.005600, 0.003812, 0.004000, 0.005083},
	{0.005600, 0.003812, 0.004000, 0.003812}, {0.005600, 0.003812, 0.004000, 0.002541},
	{0.005600, 0.003812, 0.003200, 0.005083}, {0.005600, 0.003812, 0.003200, 0.003812},
	{0.005600, 0.003812, 0.003200, 0.002541}, {0.005600, 0.002541, 0.004800, 0.005083},
	{0.005600, 0.002541, 0.004800, 0.003812}, {0.005600, 0.002541, 0.004800, 0.002541},
	{0.005600, 0.002541, 0.004000, 0.005083}, {0.005600, 0.002541, 0.004000, 0.003812},
	{0.005600, 0.002541, 0.004000, 0.002541}, {0.005600, 0.002541, 0.003200, 0.005083},
	{0.005600, 0.002541, 0.003200, 0.003812}, {0.005600, 0.002541, 0.003200, 0.002541},
	{0.003200, 0.002541, 0.002400, 0.003812},
}

func tryToFindCoeffForNewId(img gocv.Mat) {
	//takes 15 min
	maxW := 19
	maxH := 19
	maxW2 := 19
	maxH2 := 19
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

					if !checkRegionsNewID2(regions) {
						continue
					}

					original2 := img.Clone()
					for _, v := range regions {
						rect := gocv.BoundingRect(v)
						gocv.Rectangle(original2, rect, color.RGBA{255, 0, 0, 255}, 1)
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

	end := time.Now()
	diff := end.Sub(start)
	fmt.Println(diff)
}

func testCoefficientsForNewId(img gocv.Mat) {
	for ix, fc := range newIdCoeffArr {
		w1c := float64(fc.w1) / float64(img.Cols())
		h1c := float64(fc.h1) / float64(img.Rows())
		w2c := float64(fc.w2) / float64(img.Cols())
		h2c := float64(fc.h2) / float64(img.Rows())

		regions, err := textRegionsInternal(img, fc)
		if err != nil {
			return
		}

		original2 := img.Clone()
		for _, v := range regions {
			rect := gocv.BoundingRect(v)
			gocv.Rectangle(original2, rect, color.RGBA{255, 0, 0, 255}, 1)
		}

		fmt.Printf("{%f, %f, %f, %f}, ", w1c, h1c, w2c, h2c)
		if ix%2 == 0 {
			fmt.Printf("\n")
		}
		// utils.ShowImageInNamedWindow(original2, fmt.Sprintf("text regions: connected."))
		original2.Close()
	}
}

func textRegionsInternal(img gocv.Mat, fc extractTextRegionIntCoeff) ([][]image.Point, error) {
	// We have to get these values from JSON or somehow from document
	symbolWidth := int(float64(img.Cols()) * symbolWidthCoeff)
	symbolHeight := int(float64(img.Rows()) * symbolHeightCoeff)

	if symbolWidth < 2 || symbolHeight < 2 {
		return nil, errors.New("Symbol's size too small. Something wrong with region at all")
	}

	original := img.Clone()
	defer original.Close()
	gray := gocv.NewMat()
	defer gray.Close()
	grad := gocv.NewMat()
	defer grad.Close()
	binarization := gocv.NewMat()
	defer binarization.Close()
	opening := gocv.NewMat()
	defer opening.Close()
	connected := gocv.NewMat()
	defer connected.Close()
	kernel := gocv.GetStructuringElement(gocv.MorphEllipse, image.Point{fc.w1, fc.h1})
	defer kernel.Close()

	gocv.CvtColor(original, gray, gocv.ColorBGRToGray)
	gocv.MorphologyEx(gray, grad, gocv.MorphGradient, kernel)

	gocv.Threshold(grad, binarization, 0.0, 255.0, gocv.ThresholdBinary|gocv.ThresholdOtsu)

	kernel = gocv.GetStructuringElement(gocv.MorphRect, image.Point{fc.w2, fc.h2})
	gocv.MorphologyEx(binarization, opening, gocv.MorphOpen, kernel)

	kernel = gocv.GetStructuringElement(gocv.MorphRect, image.Point{symbolWidth, 1})
	gocv.MorphologyEx(opening, connected, gocv.MorphClose, kernel)

	regions := gocv.FindContours(connected, gocv.RetrievalCComp, gocv.ChainApproxSimple)
	return regions, nil
}

//TextRegions returns text regions on image
func TextRegions(img gocv.Mat) ([][]image.Point, error) {
	// We have to get these values from JSON or somehow from document
	// tryToFindCoeffForNewId(img)
	testCoefficientsForNewId(img)
	return textRegionsInternal(img, extractTextRegionIntCoeff{7, 3, 5, 4})
}

//RecognizeRegions sends found regions to tesseract ocr
func RecognizeRegions(img gocv.Mat, regions [][]image.Point, preview string) (result []block, path string) {
	//We have to get these values from JSON or somehow from document

	symbolWidth := int(float64(img.Cols()) * symbolWidthCoeff)
	symbolHeight := int(float64(img.Rows()) * symbolHeightCoeff)

	client := gosseract.NewClient()
	defer client.Close()

	client.SetLanguage("rus", "eng")

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

		// text := "hoho"
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
