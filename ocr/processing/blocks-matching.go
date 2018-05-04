package processing

import (
	"math"

	"github.com/maddevsio/go-idmatch/templates"
)

func MatchBlocks(blocks []block, t templates.Card) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	for _, field := range t.Structure {
		min := 100.0
		for _, item := range blocks {
			if d := math.Sqrt(float64((item.x-field.X)*(item.x-field.X)) + float64((item.y-field.Y)*(item.y-field.Y))); d < min && d < 0.05 {
				min = d
				data[field.Field] = item.text
				// fmt.Println(field.Field, item.text, min)
			}
		}
	}
	return data, nil
}
