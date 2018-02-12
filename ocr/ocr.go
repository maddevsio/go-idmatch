package ocr

import (
	"log"
	"os"

	"github.com/maddevsio/go-idmatch/ocr/preprocessing"
	"github.com/maddevsio/go-idmatch/ocr/processing"
)

func Recognize(file, template, preview string) (map[string]interface{}, string) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		log.Fatal(err)
	}
	roi := preprocessing.Contours(file)
	regions := processing.TextRegions(roi)
	blocks, path := processing.RecognizeRegions(roi, regions, preview)
	output, err := processing.MatchBlocks(blocks, template)
	if err != nil {
		log.Fatal(err)
	}
	return output, path
}
