package templates

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/maddevsio/go-idmatch/config"
	"github.com/maddevsio/go-idmatch/ocr/preprocessing"
	"gocv.io/x/gocv"
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

type Field struct {
	Name     string
	Text     string
	Raw      []byte
	Type     string
	Language string
	Prefix   string
	Length   int
	Position struct {
		X float64
		Y float64
	}
}

type Side struct {
	Img       gocv.Mat
	Sample    string
	Match     []preprocessing.MatchPoint
	Cols      float64
	Structure []Field
}

type Card struct {
	Type                         string
	AspectRatio                  float64
	TextBlocksThreshold          float64
	TextRegionFilterCoefficients TextRegionFilterCoefficientsT
	MaxQualitySizes              MaxQualitySizesT
	Front, Back                  Side
}

func Load(name string) (list []Card, err error) {
	dir, err := ioutil.ReadDir(config.Template.Path)
	if err != nil {
		return list, err
	}

	for _, file := range dir {
		var f TemplateFile
		jsonFile, err := ioutil.ReadFile(config.Template.Path + file.Name())
		if err != nil {
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
