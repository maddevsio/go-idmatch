package ocr

import (
	"fmt"
	"image"
	"os"
	"sync"

	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/ocr/preprocessing"
	"github.com/maddevsio/go-idmatch/ocr/processing"
	"github.com/maddevsio/go-idmatch/templates"
	"github.com/maddevsio/go-idmatch/utils"
	"gocv.io/x/gocv"
)

type t struct {
	matchpoint []preprocessing.MatchPoint
	card       templates.Card
	cols       int
}

func Recognize(file, template, preview string) (map[string]interface{}, string) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		log.Print(log.ErrorLevel, err.Error())
		return nil, ""
	}
	cards, err := templates.Load(template)
	if err != nil {
		log.Print(log.ErrorLevel, "Failed to load \""+template+"\" template")
		return nil, ""
	}

	img := gocv.IMRead(file, gocv.IMReadColor)
	k := 1000.0 / float64(img.Rows())
	gocv.Resize(img, &img, image.Point{0, 0}, k, k, gocv.InterpolationCubic)

	var wg sync.WaitGroup
	result := make(chan t, 5)

	for _, v := range cards {
		wg.Add(1)
		go func(v templates.Card) {
			defer wg.Done()
			if len(v.Sample) == 0 {
				log.Print(log.ErrorLevel, "Required \"Sample\" path is missing in template json")
				return
			}
			if _, err = os.Stat(v.Sample); os.IsNotExist(err) {
				log.Print(log.ErrorLevel, "Document sample file is missing")
				return
			}
			s := gocv.IMRead(v.Sample, gocv.IMReadColor)
			defer s.Close()
			m := preprocessing.Match(img, s)
			fmt.Printf("%s: %d\n", v.Type, len(m))

			result <- t{card: v, cols: s.Cols(), matchpoint: m}
		}(v)
	}
	wg.Wait()
	close(result)

	var match t
	for v := range result {
		if len(v.matchpoint) > len(match.matchpoint) {
			match = v
		}
	}

	sample := gocv.IMRead(match.card.Sample, gocv.IMReadGrayScale)
	defer sample.Close()

	img, err = preprocessing.Contour(img, sample, match.matchpoint, match.card.AspectRatio, float64(match.cols))
	if err != nil {
		log.Print(log.ErrorLevel, err.Error())
		return nil, ""
	}

	regions, err := processing.TextRegions(img, match.card)
	if err != nil {
		log.Print(log.ErrorLevel, "Failed to find text regions")
		return nil, ""
	}

	blocks, path := processing.RecognizeRegions(img, match.card, regions, preview)
	output := processing.MatchBlocks(blocks, match.card)

	img.Close()
	utils.Sanitize(output, match.card)
	return output, path
}
