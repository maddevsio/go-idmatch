package ocr

import (
	"os"

	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/ocr/preprocessing"
	"github.com/maddevsio/go-idmatch/ocr/processing"
	"github.com/maddevsio/go-idmatch/templates"
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
	roi := preprocessing.Contours(file, card)
	regions := processing.TextRegions(roi)
	blocks, path := processing.RecognizeRegions(roi, regions, preview)
	output, err := processing.MatchBlocks(blocks, card)
	if err != nil {
		log.Print(log.ErrorLevel, err.Error())
		return nil, ""
	}
	return output, path
}
