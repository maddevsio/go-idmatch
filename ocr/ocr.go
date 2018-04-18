package ocr

import (
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
	card, err := templates.Load(template)
	if err != nil {
		log.Print(log.ErrorLevel, "Failed to load \""+template+"\" template")
		os.Exit(1)
	}

	// Need to change template usage approach
	if len(card.Sample) == 0 {
		log.Print(log.ErrorLevel, "\"Sample\" is missing in template json")
		os.Exit(1)
	}

	if _, err := os.Stat(card.Sample); os.IsNotExist(err) {
		log.Print(log.ErrorLevel, "Document sample file is missing")
		os.Exit(1)
	}

	img := gocv.IMRead(file, gocv.IMReadColor)

	img, err = preprocessing.Contour(img, card.Sample, card.AspectRatio)
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
