package ocr

import (
	"fmt"
	"image"
	"os"

	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/ocr/preprocessing"
	"github.com/maddevsio/go-idmatch/ocr/processing"
	"github.com/maddevsio/go-idmatch/templates"
	"github.com/maddevsio/go-idmatch/utils"
	"gocv.io/x/gocv"
)

func Recognize(file, template, preview string) (map[string]interface{}, string) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		log.Print(log.ErrorLevel, err.Error())
		return nil, ""
	}
	cards, err := templates.Load(template)
	if err != nil {
		log.Print(log.ErrorLevel, "Failed to load \""+template+"\" template")
		os.Exit(1)
	}

	img := gocv.IMRead(file, gocv.IMReadColor)
	k := 1000.0 / float64(img.Rows())
	gocv.Resize(img, &img, image.Point{0, 0}, k, k, gocv.InterpolationCubic)

	var max int
	var card templates.Card
	var match []preprocessing.MatchPoint
	sample := gocv.NewMat()
	defer sample.Close()
	for _, v := range cards {
		// Need to change template usage approach
		if len(v.Sample) == 0 {
			log.Print(log.ErrorLevel, "Required \"Sample\" path is missing in template json")
			continue
		}
		if _, err := os.Stat(v.Sample); os.IsNotExist(err) {
			log.Print(log.ErrorLevel, "Document sample file is missing")
			continue
		}
		s := gocv.IMRead(v.Sample, gocv.IMReadGrayScale)
		m := preprocessing.Match(img, s)
		fmt.Printf("%s: %d\n", v.Type, len(m))
		if len(m) > max {
			max = len(m)
			card = v
			sample = s
			match = m
		}
	}

	img, err = preprocessing.Contour(img, match, card.AspectRatio, float64(sample.Cols()))
	if err != nil {
		log.Print(log.ErrorLevel, err.Error())
		return nil, ""
	}

	regions, err := processing.TextRegions(img, card)
	if err != nil {
		log.Print(log.ErrorLevel, "Failed to find text regions")
		return nil, ""
	}

	blocks, path := processing.RecognizeRegions(img, card, regions, preview)
	output, err := processing.MatchBlocks(blocks, card)
	if err != nil {
		log.Print(log.ErrorLevel, err.Error())
		return nil, ""
	}

	utils.Sanitize(output, card)
	return output, path
}
