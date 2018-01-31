package ocr

import (
	"fmt"
	"log"

	"github.com/tzununbekov/go-idmatch/ocr/preprocessing"
	"github.com/tzununbekov/go-idmatch/ocr/processing"
)

func Recognize(file, template string) {
	var output []byte
	roi := preprocessing.Contours(file)
	regions := processing.TextRegions(roi)
	blocks := processing.RecognizeRegions(roi, regions)
	output, err := processing.MatchBlocks(blocks, template)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(output))
}
