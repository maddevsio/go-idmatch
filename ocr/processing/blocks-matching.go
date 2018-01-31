package processing

import (
	"log"

	"github.com/tzununbekov/go-idmatch/templates"
)

func normalizePosition() {

}

func IdentifyBlocks(text []block, template string) {
	_, err := templates.Load(template)
	if err != nil {
		log.Fatal(err)
	}
}
