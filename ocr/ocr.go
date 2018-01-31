package ocr

import (
	"github.com/tzununbekov/go-idmatch/ocr/preprocessing"
	"github.com/tzununbekov/go-idmatch/ocr/processing"
)

func Recognize(file, template string) {
	roi := preprocessing.Contours(file)
	regions := processing.TextRegions(roi)
	text := processing.RecognizeRegions(roi, regions)
	processing.IdentifyBlocks(text, template)
}
