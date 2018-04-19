package templates

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/maddevsio/go-idmatch/config"
)

type TemplateFile struct {
	Template []Card `json:"card"`
}

type MaxQualitySizesT struct {
	MaxQualitySymWidth  float64 `json:"maxQualitySymWidth"`
	MaxQualityWidth     float64 `json:"maxQualityWidth"`
	MaxQualitySymHeight float64 `json:"maxQualitySymHeight"`
	MaxQualityHeight    float64 `json:"maxQualityHeight"`
}

type TextRegionFilterCoefficientsT struct {
	W1 float64 `json:"w1"`
	H1 float64 `json:"h1"`
	W2 float64 `json:"w2"`
	H2 float64 `json:"h2"`
}

type Card struct {
	Type                         string                        `json:"type"`
	AspectRatio                  float64                       `json:"aspectRatio"`
	TextBlocksThreshold          float64                       `json:"textBlocksThreshold"`
	Sample                       string                        `json:"sample"`
	TextRegionFilterCoefficients TextRegionFilterCoefficientsT `json:"textRegionFilterCoefficients"`
	MaxQualitySizes              MaxQualitySizesT              `json:"maxQualitySizes"`

	Structure []struct {
		Field string  `json:"field"`
		X     float64 `json:"x"`
		Y     float64 `json:"y"`
		Type  string  `json:"type"`
	} `json:"structure"`
}

func Load(name string) (list []Card, err error) {
	var f TemplateFile
	// Need to add config package with paths
	dir, err := ioutil.ReadDir(config.Template.Path)
	if err != nil {
		return list, err
	}

	for _, file := range dir {
		jsonFile, err := ioutil.ReadFile(config.Template.Path + file.Name())
		if err != nil {
			fmt.Println(err.Error())
			return list, err
		}
		err = json.Unmarshal(jsonFile, &f)
		if err != nil {
			return list, err
		}

		for _, template := range f.Template {
			if template.Type == name || len(name) == 0 {
				list = append(list, template)
			}
		}
	}
	if len(list) == 0 {
		return list, errors.New("Template \"" + name + "\" not found")
	}
	return list, nil
}
