package config

import (
	"os"

	"github.com/maddevsio/go-idmatch/log"

	gcfg "gopkg.in/gcfg.v1"
)

type preprocessing struct {
	HoughThreshold      int
	BorderStep          int
	ErodeLength         int
	DilateThickness     int
	MinAreaRatio        float64
	MaxAreaRatio        float64
	MaxAspectDelta      float64
	CleanCannyT1        float32
	CleanCannyT2        float32
	CleanCannyBlurSize  int
	CleanCannyBlurSigma float64
	CannyT1             float32
	CannyT2             float32
	CannyBlurSize       int
	CannyBlurSigma      float64
}

type processing struct {
}

type template struct {
	Path string
}

type web struct {
	Uploads   string
	Preview   string
	Static    string
	Templates string
}

type configFile struct {
	Preprocessing preprocessing
	Processing    processing
	Template      template
	Web           web
}

const defaultConfig = `
	[preprocessing]
	houghThreshold = 200
	borderStep = 5
	erodeLength = 15
	dilateThickness = 2
	minAreaRatio = 0.33
	maxAreaRatio = 0.97
	maxAspectDelta = 0.1
	cleanCannyT1 = 30
	cleanCannyT2 = 170
	cleanCannyBlurSize = 7
	cleanCannyBlurSigma = 10
	cannyT1 = 10
	cannyT2 = 50
	cannyBlurSize = 3
	cannyBlurSigma = 5

	[processing]

	[template]
	path = "templates/json/"
	
	[web]
	static = "web/static/"
	uploads = "web/uploads/"
	preview = "web/preview/"
	templates = "web/templates/"
`

var (
	config        configFile
	Preprocessing preprocessing
	Processing    processing
	Template      template
	Web           web
)

func init() {
	if err := gcfg.ReadStringInto(&config, defaultConfig); err != nil {
		log.Print(log.ErrorLevel, err.Error())
		os.Exit(1)
	}
	if err := gcfg.ReadFileInto(&config, "config.gcfg"); err != nil {
		log.Print(log.DebugLevel, "Unable to read config.gcfg, loading default config")
	}
	Preprocessing = config.Preprocessing
	Processing = config.Processing
	Template = config.Template
	Web = config.Web
}
