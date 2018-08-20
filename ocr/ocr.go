package ocr

import (
	"fmt"
	"os"
	"sync"

	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/ocr/preprocessing"
	"github.com/maddevsio/go-idmatch/ocr/processing"
	"github.com/maddevsio/go-idmatch/templates"
	"gocv.io/x/gocv"
)

func readAndScale(path string) (gocv.Mat, error) {
	if _, err := os.Stat(path); err != nil {
		return gocv.Mat{}, err
	}
	img := gocv.IMRead(path, gocv.IMReadColor)
	// k := 1000.0 / float64(img.Rows())
	// gocv.Resize(img, &img, image.Point{0, 0}, k, k, gocv.InterpolationCubic)
	return img, nil
}

func contour(side templates.Side, ratio float64) (gocv.Mat, error) {
	sampleMap := gocv.IMRead(side.Sample, gocv.IMReadGrayScale)
	defer sampleMap.Close()
	return preprocessing.Contour(side.Img, sampleMap, side.Match, ratio, side.Cols)
}

func regions(img gocv.Mat, c templates.Card) ([]processing.Block, error) {
	regions, err := processing.TextRegions(img, c)
	if err != nil {
		return nil, err
	}
	blocks, _ := processing.RecognizeRegions(img, c, regions)
	return blocks, nil
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
	defer frontside.Close()
	defer backside.Close()

	var wg sync.WaitGroup
	result := make(chan templates.Card, 5)

	for _, v := range cards {
		wg.Add(1)
		go func(v templates.Card) {
			defer wg.Done()
			frontSample := gocv.IMRead(v.Front.Sample, gocv.IMReadGrayScale)
			backSample := gocv.IMRead(v.Back.Sample, gocv.IMReadGrayScale)
			defer frontSample.Close()
			defer backSample.Close()
			t := v
			t.Front.Cols = float64(frontSample.Cols())
			t.Back.Cols = float64(backSample.Cols())
			if _, err := os.Stat(v.Front.Sample); ferr == nil && err == nil {
				t.Front.Match = preprocessing.Match(frontside, frontSample)
			}
			if _, err := os.Stat(v.Back.Sample); berr == nil && err == nil {
				t.Back.Match = preprocessing.Match(backside, backSample)
			}
			fmt.Printf("%s: frontside %d, backside %d\n", v.Type, len(t.Front.Match), len(t.Back.Match))
			result <- t
		}(v)
	}
	wg.Wait()
	close(result)

	var match templates.Card
	for v := range result {
		if len(v.Front.Match)+len(v.Back.Match) > len(match.Front.Match)+len(match.Back.Match) {
			match = v
		}
	}

	match.Front.Img = frontside
	match.Back.Img = backside
	output := make(chan []templates.Field, 10)

	for _, v := range []templates.Side{match.Front, match.Back} {
		wg.Add(1)
		go func(v templates.Side) {
			defer wg.Done()
			if len(v.Match) != 0 {
				img, err := contour(v, match.AspectRatio)
				if err != nil {
					log.Print(log.ErrorLevel, err.Error())
					return
				}
				blocks, err := regions(img, match)
				if err != nil {
					log.Print(log.ErrorLevel, err.Error())
					return
				}
				processing.MatchBlocks(blocks, v, v.Img)
				processing.RecognizeText(v.Structure)
				fmt.Println(v.Structure)
				output <- v.Structure
			}
		}(v)
	}
	wg.Wait()
	close(output)

	var data []templates.Field
	for m := range output {
		data = append(m, data...)
	}

	r := make(map[string]interface{})
	for _, v := range data {
		r[v.Name] = v.Text
	}
	return r, ""
}
