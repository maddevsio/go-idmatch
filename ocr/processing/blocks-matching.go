package processing

import (
	"encoding/json"
	"log"

	"github.com/tzununbekov/go-idmatch/templates"
)

func MatchBlocks(blocks []block, template string) ([]byte, error) {
	data := make(map[string]interface{})
	t, err := templates.Load(template)
	if err != nil {
		log.Fatal(err)
	}
	// Need to revise this logic
	for _, field := range t.Structure {
		for _, item := range blocks {
			if field.X >= item.x*0.95 && field.X <= (item.x+item.w)*1.1 {
				if field.Y >= item.y*0.95 && field.Y <= (item.y+item.h)*1.1 {
					data[field.Field] = item.text
				}
			}
		}
	}
	return json.MarshalIndent(data, "", "	")
}
