package preprocessing

import (
	"fmt"
	"image"
	"math"
	"math/rand"

	"github.com/MaxHalford/gago"
	"github.com/maddevsio/go-idmatch/log"
	"gocv.io/x/gocv"
)

type Factors struct {
	Blur    []float64
	CannyT1 []float64
	CannyT2 []float64
}

var Rect image.Rectangle

func (X Factors) Evaluate() (float64, error) {
	var res float64
	child := gocv.NewMat()
	defer child.Close()

	a, b := float32(math.Abs(X.CannyT1[0])), float32(math.Abs(X.CannyT2[0]))
	c := int(X.Blur[0])
	if c%2 != 1 {
		c++
	}
	if c <= 0 {
		c = 1
	} else if c > 21 {
		c = 21
	}
	log.Print(log.DebugLevel, fmt.Sprintf("Blur: %v, canny min: %v, canny max: %v\n", c, a, b))

	gocv.MedianBlur(Img, child, c)
	gocv.Canny(child, child, a, b)

	var maxArea int
	var maxRect image.Rectangle
	contours := gocv.FindContours(child, gocv.RetrievalTree, gocv.ChainApproxSimple)
	for _, v := range contours {
		rect := gocv.BoundingRect(v)
		if area := rect.Dx() * rect.Dy(); area > maxArea {
			maxArea = area
			maxRect = rect
		}
	}

	ratio := float64(maxRect.Dx()) / float64(maxRect.Dy())
	deltaArea := OriginalArea - maxArea
	deltaRatio := math.Abs(OriginalRatio - ratio)
	res = float64(deltaArea)
	if deltaRatio < 0.1 {
		res *= deltaRatio
	}

	log.Print(log.DebugLevel, fmt.Sprintf("rect: %v, dRatio: %v, dArea: %v, result: %v\n", maxRect, deltaRatio, deltaArea, res))
	Rect = maxRect

	return float64(res), nil
}

func (X Factors) Mutate(rng *rand.Rand) {}

func (X Factors) Crossover(Y gago.Genome, rng *rand.Rand) {}

func (X Factors) Clone() gago.Genome {
	return Factors{
		Blur:    X.Blur,
		CannyT1: X.CannyT1,
		CannyT2: X.CannyT2,
	}
}

func InitValues(rng *rand.Rand) gago.Genome {
	return Factors{
		Blur:    gago.InitUnifFloat64(1, 1, 21, rng),
		CannyT1: gago.InitUnifFloat64(1, 1, 30, rng),
		CannyT2: gago.InitUnifFloat64(1, 30, 180, rng),
	}
}

func GetFactors() {
	var ga = gago.Generational(InitValues)
	ga.ParallelEval = true
	if err := ga.Initialize(); err != nil {
		log.Print(log.WarnLevel, err.Error())
	}

	for i := 1; i < 4; i++ {
		if err := ga.Evolve(); err != nil {
			log.Print(log.WarnLevel, err.Error())
		}
		fmt.Printf("Best fitness at generation %d: %f (%v)\n", i, ga.HallOfFame[0].Fitness, ga.HallOfFame[0].Genome)
	}
}
