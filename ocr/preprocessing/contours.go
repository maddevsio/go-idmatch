package preprocessing

import (
	"crypto/md5"
	"errors"
	"fmt"
	"image"
	"math"
	"math/rand"
	"sync"

	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"
)

type point struct {
	descriptor []float64
	keypoint   gocv.KeyPoint
}

type MatchPoint struct {
	a        point
	b        point
	distance float64
}

type Cache map[string]*[]point

var (
	cache Cache
	lock  sync.Mutex
)

func InitCache() {
	cache = make(map[string]*[]point)
}

func descriptorArr(gray gocv.Mat) (p []point) {
	sift := contrib.NewSIFT()
	defer sift.Close()

	mask := gocv.NewMat()
	defer mask.Close()

	k, d := sift.DetectAndCompute(gray, mask)
	defer d.Close()

	for i, v := range k {
		var tmp []float64
		for j := 0; j < d.Cols(); j++ {
			tmp = append(tmp, float64(d.GetFloatAt(i, j)))
		}
		p = append(p, point{keypoint: v, descriptor: tmp})
	}
	return
}

func arrayDistance(a, b []float64) float64 {
	var d float64
	for i := range a {
		d += (b[i] - a[i]) * (b[i] - a[i])
	}
	return math.Sqrt(d)
}

func pointDistance(a, b image.Point) float64 {
	return math.Sqrt(float64((b.X-a.X)*(b.X-a.X)) + float64((b.Y-a.Y)*(b.Y-a.Y)))
}

func triangleAngles(a, b, c float64) (float64, float64, float64) {
	cc := (math.Pow(b, 2) + math.Pow(c, 2) - math.Pow(a, 2)) / (2 * b * c)
	aa := (math.Pow(c, 2) + math.Pow(a, 2) - math.Pow(b, 2)) / (2 * c * a)
	bb := (math.Pow(a, 2) + math.Pow(b, 2) - math.Pow(c, 2)) / (2 * a * b)

	return aa, bb, cc
}

func anglesByVertex(p1, p2, p3 image.Point) (float64, float64, float64) {
	return triangleAngles(pointDistance(p1, p2), pointDistance(p2, p3), pointDistance(p3, p1))
}

func matchDescriptors(a, b []point) []MatchPoint {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}
	mask := make([]bool, len(b))
	var match []MatchPoint
	for _, aa := range a {
		mp := MatchPoint{distance: 10000}
		index := 0
		for i, bb := range b {
			if mask[i] {
				continue
			}
			if d := arrayDistance(aa.descriptor, bb.descriptor); d < mp.distance {
				mp = MatchPoint{a: aa, b: bb, distance: d}
				index = i
			}
		}
		// fmt.Printf("matching descriptors: %d/%d\r", len(a), k)
		match = append(match, mp)
		mask[index] = true
	}
	return match
}

func filterGoodMatch(match []MatchPoint) []MatchPoint {
	var goodMatch []MatchPoint
	for _, v := range match {
		if v.distance > 200 {
			continue
		}
		goodMatch = append(goodMatch, v)
	}
	return goodMatch
}

func matchTriangles(goodMatch []MatchPoint, threshold int) []MatchPoint {
	if len(goodMatch) == 0 {
		return nil
	}
	var counter int
	var equals []MatchPoint
	for i := 0; i < 100000 && counter < threshold; i++ {
		// Getting set of three random descriptors to form a triangle
		rand1 := rand.Intn(len(goodMatch))
		rand2 := rand.Intn(len(goodMatch))
		rand3 := rand.Intn(len(goodMatch))
		// Need to use triangle struc
		p1 := image.Point{int(goodMatch[rand1].a.keypoint.X), int(goodMatch[rand1].a.keypoint.Y)}
		p2 := image.Point{int(goodMatch[rand2].a.keypoint.X), int(goodMatch[rand2].a.keypoint.Y)}
		p3 := image.Point{int(goodMatch[rand3].a.keypoint.X), int(goodMatch[rand3].a.keypoint.Y)}
		pp1 := image.Point{int(goodMatch[rand1].b.keypoint.X), int(goodMatch[rand1].b.keypoint.Y)}
		pp2 := image.Point{int(goodMatch[rand2].b.keypoint.X), int(goodMatch[rand2].b.keypoint.Y)}
		pp3 := image.Point{int(goodMatch[rand3].b.keypoint.X), int(goodMatch[rand3].b.keypoint.Y)}
		a, b, c := anglesByVertex(p1, p2, p3)
		aa, bb, cc := anglesByVertex(pp1, pp2, pp3)

		// "Line shaped triangles" gives too much false match
		if math.Abs(a) > 0.9999 || math.Abs(b) > 0.9999 || math.Abs(c) > 0.9999 {
			continue
		}

		// Allowed rate of similarity
		if math.Abs(a-aa) < 0.0001 && math.Abs(b-bb) < 0.0001 && math.Abs(c-cc) < 0.0001 {
			equals = append(equals, goodMatch[rand1])
			equals = append(equals, goodMatch[rand2])
			equals = append(equals, goodMatch[rand3])
			counter++
		}
	}
	return equals
}

func get(key string) (*[]point, bool) {
	value, ok := cache[key]
	return value, ok
}

func set(key string, value *[]point) {
	if cache != nil {
		lock.Lock()
		defer lock.Unlock()
		cache[key] = value
	}
}

func Match(img, sample gocv.Mat) []MatchPoint {
	hash := fmt.Sprintf("%x", md5.Sum(sample.ToBytes()))

	a, hit := get(hash)
	if !hit {
		aa := descriptorArr(sample)
		a = &aa
		set(hash, a)
	}

	gray := gocv.NewMat()
	defer gray.Close()

	if img.Empty() {
		return []MatchPoint{}
	}

	gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	b := descriptorArr(gray)

	// for _, v := range *a {
	// 	gocv.Circle(&sample, image.Point{int(v.keypoint.X), int(v.keypoint.Y)}, int(v.keypoint.Size), color.RGBA{255, 0, 0, 255}, 1)
	// }

	// utils.ShowImage(sample)

	return filterGoodMatch(matchDescriptors(*a, b))
}

func Contour(img, sample gocv.Mat, goodMatch []MatchPoint, ratio, sampleWidth float64) (gocv.Mat, error) {
	// Two steps matching: finding similar triangles and matching their position
	var equals []MatchPoint
	miss := true
	for set := 0; set < 10 && miss; set++ {
		// Getting 3 similar triangles out of most equal descriptors
		equals = matchTriangles(goodMatch, 3)

		if len(equals)/3 < 3 {
			fmt.Printf("Not enough similar triangles in set %d (need 3, got %d)\n", set, len(equals)/3)
			continue
		}

		// Matching trianlges positions
		fmt.Printf("testing %d set of triangles\n", set+1)
		for i := 0; i < len(equals)/3; i++ {
			// Need to use triangle struct
			p1 := image.Point{X: int(equals[i].a.keypoint.X), Y: int(equals[i].a.keypoint.Y)}
			p2 := image.Point{X: int(equals[i+3].a.keypoint.X), Y: int(equals[i+3].a.keypoint.Y)}
			p3 := image.Point{X: int(equals[i+6].a.keypoint.X), Y: int(equals[i+6].a.keypoint.Y)}
			pp1 := image.Point{X: int(equals[i].b.keypoint.X), Y: int(equals[i].b.keypoint.Y)}
			pp2 := image.Point{X: int(equals[i+3].b.keypoint.X), Y: int(equals[i+3].b.keypoint.Y)}
			pp3 := image.Point{X: int(equals[i+6].b.keypoint.X), Y: int(equals[i+6].b.keypoint.Y)}
			a, b, c := anglesByVertex(p1, p2, p3)
			aa, bb, cc := anglesByVertex(pp1, pp2, pp3)
			miss = false
			if math.Abs(a-aa) > 0.1 || math.Abs(b-bb) > 0.1 || math.Abs(c-cc) > 0.1 {
				miss = true
				break
			}
		}
		if !miss {
			break
		}
	}

	if miss {
		return img, errors.New("Cannot find equaly located similar triangles")
	}

	theta1 := math.Atan2(equals[1].a.keypoint.Y-equals[0].a.keypoint.Y, equals[1].a.keypoint.X-equals[0].a.keypoint.X)
	theta2 := math.Atan2(equals[1].b.keypoint.Y-equals[0].b.keypoint.Y, equals[1].b.keypoint.X-equals[0].b.keypoint.X)
	theta := (theta2 - theta1) * 180 / math.Pi

	fmt.Printf("rotation angle: %v\n", theta)

	// Need to fix this ugly "*2" workaround with proper bounding rotation not to cut off edges
	rotation := gocv.GetRotationMatrix2D(image.Point{int(equals[1].b.keypoint.X), int(equals[1].b.keypoint.Y)}, theta, 1)
	gocv.WarpAffine(img, &img, rotation, image.Point{img.Cols() * 2, img.Rows() * 2})

	d1 := math.Sqrt(math.Pow(equals[1].a.keypoint.X-equals[0].a.keypoint.X, 2) + math.Pow(equals[1].a.keypoint.Y-equals[0].a.keypoint.Y, 2))
	d2 := math.Sqrt(math.Pow(equals[1].b.keypoint.X-equals[0].b.keypoint.X, 2) + math.Pow(equals[1].b.keypoint.Y-equals[0].b.keypoint.Y, 2))
	scale := d2 / d1

	fmt.Printf("scale rate: %v\n", scale)

	left := equals[1].b.keypoint.X - equals[1].a.keypoint.X*scale
	right := equals[1].b.keypoint.X + (sampleWidth-equals[1].a.keypoint.X)*scale
	top := equals[1].b.keypoint.Y - equals[1].a.keypoint.Y*scale
	bottom := top + (right-left)/ratio

	if left < 0 {
		left = 0
	}
	if top < 0 {
		top = 0
	}
	if right > float64(img.Cols()) {
		right = float64(img.Cols())
	}
	if bottom > float64(img.Rows()) {
		bottom = float64(img.Rows())
	}

	// gocv.Rectangle(&img, image.Rect(int(left), int(top), int(right), int(bottom)), color.RGBA{0, 255, 0, 255}, 2)
	// utils.ShowImage(img)

	img = img.Region(image.Rect(int(left), int(top), int(right), int(bottom)))

	return img, nil
}
