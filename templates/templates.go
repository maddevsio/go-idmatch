package templates

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/maddevsio/go-idmatch/config"
)

type TemplateFile struct {
	Template []Card `json:"card"`
}

type Card struct {
	Type      string `json:"type"`
	Structure []struct {
		Field string  `json:"field"`
		X     float64 `json:"x"`
		Y     float64 `json:"y"`
	} `json:"structure"`
}

func Load(name string) (template Card, err error) {
	var f TemplateFile
	// Need to add config package with paths
	dir, err := ioutil.ReadDir(config.Template.Path)
	if err != nil {
		return template, err
	}

	for _, file := range dir {
		jsonFile, err := ioutil.ReadFile(config.Template.Path + file.Name())
		if err != nil {
			return template, err
		}
		err = json.Unmarshal(jsonFile, &f)
		if err != nil {
			return template, err
		}
		// JSON file may contain multiple templates. Needs to be reviewed.
		for _, template := range f.Template {
			if template.Type == name {
				return template, nil
			}
		}
	}
	return template, errors.New("Template \"" + name + "\" not found")
}
