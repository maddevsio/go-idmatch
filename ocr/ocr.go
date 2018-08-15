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
	frontMatch []preprocessing.MatchPoint
	backMatch  []preprocessing.MatchPoint
	card       templates.Card
	cols       int
}

func readAndScale(path string) (gocv.Mat, error) {
	if _, err := os.Stat(path); err != nil {
		return gocv.Mat{}, err
	}
	img := gocv.IMRead(path, gocv.IMReadColor)
	k := 1000.0 / float64(img.Rows())
	gocv.Resize(img, &img, image.Point{0, 0}, k, k, gocv.InterpolationCubic)
	return img, nil
}

func cropAndAnalyze(img gocv.Mat, samplePath string, mp []preprocessing.MatchPoint, ratio, cols float64, card templates.Card) (map[string]interface{}, gocv.Mat) {
	sample := gocv.IMRead(samplePath, gocv.IMReadGrayScale)
	defer sample.Close()
	aligned, err := preprocessing.Contour(img, sample, mp, ratio, cols)
	if err != nil {
		log.Print(log.ErrorLevel, err.Error())
		return nil, gocv.Mat{}
	}
	regions, err := processing.TextRegions(aligned, card)
	if err != nil {
		log.Print(log.ErrorLevel, "Failed to find text regions")
		return nil, gocv.Mat{}
	}
	blocks, preview := processing.RecognizeRegions(aligned, card, regions)
	output := processing.MatchBlocks(blocks, match.card.Front, aligned)
	return output, preview
}

func Recognize(front, back, template, preview string) (map[string]interface{}, string) {
	cards, err := templates.Load(template)
	if err != nil {
		log.Print(log.ErrorLevel, "Failed to load \""+template+"\" template")
		return nil, ""
	}

	if len(front) == 0 && len(back) == 0 {
		log.Print(log.ErrorLevel, "Please provide at least one side image of document")
		return nil, ""
	}

	// A bit more uglyness
	frontside, ferr := readAndScale(front)
	backside, berr := readAndScale(back)

	var wg sync.WaitGroup
	var f, b []preprocessing.MatchPoint
	result := make(chan t, 5)

	for _, v := range cards {
		wg.Add(1)
		go func(v templates.Card) {
			defer wg.Done()
			frontSample := gocv.IMRead(v.Front.Sample, gocv.IMReadGrayScale)
			backSample := gocv.IMRead(v.Back.Sample, gocv.IMReadGrayScale)
			defer frontSample.Close()
			defer backSample.Close()
			if _, err := os.Stat(v.Front.Sample); ferr == nil && err == nil {
				f = preprocessing.Match(frontside, frontSample)
			}
			if _, err := os.Stat(v.Back.Sample); berr == nil && err == nil {
				b = preprocessing.Match(backside, backSample)
			}
			fmt.Printf("%s: frontside %d, backside %d\n", v.Type, len(f), len(b))

			result <- t{card: v, cols: frontSample.Cols(), frontMatch: f, backMatch: b}
		}(v)
	}
	wg.Wait()
	close(result)

	var match t
	for v := range result {
		if len(v.frontMatch) > len(match.frontMatch) {
			match = v
		}
	}

	// output := make(map[string]interface{})
	if len(match.frontMatch) != 0 {
		frontSample := gocv.IMRead(match.card.Front.Sample, gocv.IMReadGrayScale)
		defer frontSample.Close()
		frontside, err = preprocessing.Contour(frontside, frontSample, match.frontMatch, match.card.AspectRatio, float64(match.cols))
		if err != nil {
			log.Print(log.ErrorLevel, err.Error())
			return nil, ""
		}
		regions, err := processing.TextRegions(frontside, match.card)
		if err != nil {
			log.Print(log.ErrorLevel, "Failed to find text regions")
			return nil, ""
		}
		blocks, _ := processing.RecognizeRegions(frontside, match.card, regions, preview)
		for k, v := range processing.MatchBlocks(blocks, match.card.Front, frontside) {
			output[k] = v
		}
	}

	if len(match.backMatch) != 0 {
		backSample := gocv.IMRead(match.card.Back.Sample, gocv.IMReadGrayScale)
		defer backSample.Close()
		backside, err = preprocessing.Contour(backside, backSample, match.backMatch, match.card.AspectRatio, float64(match.cols))
		if err != nil {
			log.Print(log.ErrorLevel, err.Error())
			return nil, ""
		}
		regions, err := processing.TextRegions(backside, match.card)
		if err != nil {
			log.Print(log.ErrorLevel, "Failed to find text regions")
			return nil, ""
		}
		blocks, _ := processing.RecognizeRegions(backside, match.card, regions, preview)
		for k, v := range processing.MatchBlocks(blocks, match.card.Back, backside) {
			output[k] = v
		}
	}

	frontside.Close()
	backside.Close()
	utils.Sanitize(output, match.card)
	return output, ""
}
