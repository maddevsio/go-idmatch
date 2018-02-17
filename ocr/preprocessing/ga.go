package preprocessing

import (
	"fmt"
	"math/rand"
	"reflect"

	"github.com/MaxHalford/gago"
	"github.com/maddevsio/go-idmatch/log"
	"gocv.io/x/gocv"
)

type Factors struct {
	Blur    []float64
	CannyT1 []float64
	CannyT2 []float64
}

func (X Factors) Evaluate() (float64, error) {
	// var ratio float64

	child := gocv.NewMat()
	defer child.Close()

	a, b := float32(X.CannyT1[0]), float32(X.CannyT2[0])
	c := int(X.Blur[0])
	if c%2 != 1 {
		c++
	} else if c <= 0 {
		c = 1
	}

	X.Blur = []float64{float64(int(c))}

	log.Print(log.DebugLevel, fmt.Sprintf("Blur: %v, canny min: %v, canny max: %v\n", c, a, b))

	gocv.MedianBlur(Img, child, c)
	gocv.Canny(child, child, a, b)

	contours := gocv.FindContours(child, gocv.RetrievalTree, gocv.ChainApproxSimple)
	area := 100000
	if len(contours) > 0 {
		rect := gocv.BoundingRect(contours[0])
		area = rect.Dx() * rect.Dy()
	}
	// maxArea = area
	// ratio = float64(rect.Dx()) / float64(rect.Dy())
	// fmt.Println(OriginalRatio-ratio, 1-float64(area)/float64(OriginalArea))
	// gocv.Rectangle(img3, rect, color.RGBA{255, 0, 0, 255}, 2)
	return float64(OriginalArea - area), nil
}

func (X Factors) Mutate(rng *rand.Rand) {
	// gago.MutNormalFloat64(X.Blur, 0.1, rng)
	// gago.MutNormalFloat64(X.CannyT1, 0.01, rng)
	// gago.MutNormalFloat64(X.CannyT2, 0.01, rng)
}

func (X Factors) Crossover(Y gago.Genome, rng *rand.Rand) {
	gago.CrossUniformFloat64(X.Blur, Y.(Factors).Blur, rng)
	gago.CrossUniformFloat64(X.CannyT1, Y.(Factors).CannyT1, rng)
	gago.CrossUniformFloat64(X.CannyT2, Y.(Factors).CannyT2, rng)
}

func (X Factors) Clone() gago.Genome {
	Y := Factors{
		Blur:    X.Blur,
		CannyT1: X.CannyT1,
		CannyT2: X.CannyT2,
	}
	return Y
}

func InitValues(rng *rand.Rand) gago.Genome {
	var X Factors
	X.Blur = gago.InitUnifFloat64(1, 1, 17, rng)
	X.CannyT1 = gago.InitUnifFloat64(1, 1, 50, rng)
	X.CannyT2 = gago.InitUnifFloat64(1, 50, 200, rng)
	return X
}

func GetFactors() (res []float64) {
	var ga = gago.Generational(InitValues)
	if err := ga.Initialize(); err != nil {
		log.Print(log.WarnLevel, err.Error())
	}

	for i := 1; i < 7; i++ {
		if err := ga.Evolve(); err != nil {
			log.Print(log.WarnLevel, err.Error())
		}
		fmt.Printf("Best fitness at generation %d: %f (%v)\n", i, ga.HallOfFame[0].Fitness, ga.HallOfFame[0].Genome)
	}

	t := reflect.ValueOf(ga.HallOfFame[0].Genome)
	for i := 0; i < t.NumField(); i++ {
		slice, ok := t.Field(i).Interface().([]float64)
		if !ok {
			log.Print(log.ErrorLevel, "Error decoding genome value")
		}
		res = append(res, slice[0])
	}
	return res
}
