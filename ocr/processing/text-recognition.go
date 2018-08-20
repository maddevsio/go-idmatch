package processing

import (
	"image"
	"image/color"
	"math"

	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/templates"
	"github.com/otiai10/gosseract"
	"gocv.io/x/gocv"
)

func MatchBlocks(blocks []Block, t templates.Side, img gocv.Mat) {
	for i, field := range t.Structure {
		min := 100.0
		for _, item := range blocks {
			gocv.Circle(&t.Img, image.Point{int(float64(img.Cols()) * field.Position.X), int(float64(img.Rows()) * field.Position.Y)}, 4, color.RGBA{0, 255, 0, 255}, 2)
			gocv.PutText(&t.Img, field.Name, image.Point{int(float64(img.Cols()) * field.Position.X), int(float64(img.Rows()) * field.Position.Y)}, gocv.FontHersheyPlain, 1, color.RGBA{0, 0, 0, 255}, 1)
			if d := math.Sqrt(float64((item.x-field.Position.X)*(item.x-field.Position.X)) + float64((item.y-field.Position.Y)*(item.y-field.Position.Y))); d < min && d < 0.1 {
				min = d
				t.Structure[i].Raw = item.raw
				gocv.PutText(&img, field.Name, image.Point{int(float64(img.Cols()) * item.x), int(float64(img.Rows()) * item.y)}, gocv.FontHersheyPlain, 1, color.RGBA{0, 0, 0, 255}, 1)
			}
		}
	}
}

func RecognizeText(fields []templates.Field) {
	client := gosseract.NewClient()
	defer client.Close()
	for k, field := range fields {
		if len(field.Language) == 0 {
			field.Language = "rus"
		}
		if err := client.SetLanguage(field.Language); err != nil {
			log.Print(log.ErrorLevel, "Set language: "+err.Error())
			continue
		}
		if err := client.SetImageFromBytes(field.Raw); err != nil {
			log.Print(log.ErrorLevel, "Text block \""+field.Name+"\": "+err.Error())
			continue
		}
		text, err := client.Text()
		if err != nil {
			log.Print(log.WarnLevel, "Get text: "+err.Error())
			continue
		}
		log.Print(log.DebugLevel, text)
		fields[k].Text = text
	}
}
