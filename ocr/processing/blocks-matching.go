package processing

import (
	"image"
	"image/color"
	"math"

	"github.com/maddevsio/go-idmatch/templates"
	"github.com/maddevsio/go-idmatch/utils"
	"gocv.io/x/gocv"
)

func MatchBlocks(blocks []block, t templates.Side, img gocv.Mat) map[string]interface{} {
	data := make(map[string]interface{})
	for _, field := range t.Structure {
		min := 100.0
		for _, item := range blocks {
			gocv.Circle(&img, image.Point{int(float64(img.Cols()) * field.X), int(float64(img.Rows()) * field.Y)}, 4, color.RGBA{0, 255, 0, 255}, 2)
			gocv.PutText(&img, field.Field, image.Point{int(float64(img.Cols()) * field.X), int(float64(img.Rows()) * field.Y)}, gocv.FontHersheyPlain, 1, color.RGBA{0, 0, 0, 255}, 1)
			if d := math.Sqrt(float64((item.x-field.X)*(item.x-field.X)) + float64((item.y-field.Y)*(item.y-field.Y))); d < min && d < 0.1 {
				min = d
				data[field.Field] = item.text
				gocv.PutText(&img, field.Field, image.Point{int(float64(img.Cols()) * item.x), int(float64(img.Rows()) * item.y)}, gocv.FontHersheyPlain, 1, color.RGBA{0, 0, 0, 255}, 1)
			}
		}
	}
	utils.ShowImage(img)
	return data
}
