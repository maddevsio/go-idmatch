package processing

import (
	"math"

	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/templates"
	"github.com/otiai10/gosseract"
	"gocv.io/x/gocv"
)

func MatchBlocks(blocks []Block, t templates.Side, img gocv.Mat) {
	// var data []templates.Field
	for i, field := range t.Structure {
		min := 100.0
		for _, item := range blocks {
			// gocv.Circle(&img, image.Point{int(float64(img.Cols()) * field.Postion.X), int(float64(img.Rows()) * field.Position.Y)}, 4, color.RGBA{0, 255, 0, 255}, 2)
			// gocv.PutText(&img, field.Name, image.Point{int(float64(img.Cols()) * field.Position.X), int(float64(img.Rows()) * field.Position.Y)}, gocv.FontHersheyPlain, 1, color.RGBA{0, 0, 0, 255}, 1)
			if d := math.Sqrt(float64((item.x-field.Position.X)*(item.x-field.Position.X)) + float64((item.y-field.Position.Y)*(item.y-field.Position.Y))); d < min && d < 0.1 {
				min = d
				t.Structure[i].Raw = item.raw
				// gocv.PutText(&img, field.Field, image.Point{int(float64(img.Cols()) * item.x), int(float64(img.Rows()) * item.y)}, gocv.FontHersheyPlain, 1, color.RGBA{0, 0, 0, 255}, 1)
			}
		}
		// data = append(data, field)
	}
	// fmt.Println(data)
	// return t.Structure
}

func RecognizeText(f []templates.Field) {
	client := gosseract.NewClient()
	defer client.Close()
	for k, v := range f {
		if len(v.Language) == 0 {
			v.Language = "rus"
		}
		if err := client.SetLanguage(v.Language); err != nil {
			log.Print(log.ErrorLevel, "Set language: "+err.Error())
			continue
		}
		if err := client.SetImageFromBytes(v.Raw); err != nil {
			log.Print(log.ErrorLevel, "Set image: "+err.Error())
			continue
		}
		text, err := client.Text()
		if err != nil {
			log.Print(log.WarnLevel, "Get text: "+err.Error())
			continue
		}
		f[k].Text = text
	}
}
