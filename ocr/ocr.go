package ocr

import (
	"log"

	"github.com/maddevsio/go-idmatch/ocr/preprocessing"
	"github.com/maddevsio/go-idmatch/ocr/processing"
)

func Recognize(file, template, preview string) (string, string) {
	var output []byte
	roi := preprocessing.Contours(file)
	regions := processing.TextRegions(roi)
	blocks, path := processing.RecognizeRegions(roi, regions, preview)
	output, err := processing.MatchBlocks(blocks, template)
	if err != nil {
		log.Fatal(err)
	}
	return string(output), path
}
