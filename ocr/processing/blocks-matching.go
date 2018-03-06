package processing

import (
	"github.com/maddevsio/go-idmatch/templates"
)

func MatchBlocks(blocks []block, t templates.Card) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	heightThreshold := 1 - t.TextBlocksThreshold
	widthThreshold := 1 + t.TextBlocksThreshold
	// Need to revise this logic
	for _, field := range t.Structure {
		for _, item := range blocks {
			if field.X >= item.x*heightThreshold && field.X <= (item.x+item.w)*widthThreshold {
				if field.Y >= item.y*heightThreshold && field.Y <= (item.y+item.h)*widthThreshold {
					data[field.Field] = item.text
				}
			}
		}
	}
	return data, nil
}
