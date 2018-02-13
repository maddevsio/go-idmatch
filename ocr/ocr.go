package ocr

import (
	"os"

	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/ocr/preprocessing"
	"github.com/maddevsio/go-idmatch/ocr/processing"
)

func Recognize(file, template, preview string) (map[string]interface{}, string) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		log.Print(log.ErrorLevel, err.Error())
		return nil, ""
	}
	roi := preprocessing.Contours(file)
	regions := processing.TextRegions(roi)
	blocks, path := processing.RecognizeRegions(roi, regions, preview)
	output, err := processing.MatchBlocks(blocks, template)
	if err != nil {
		log.Print(log.ErrorLevel, err.Error())
		return nil, ""
	}
	return output, path
}
