package config

import (
	"os"

	"github.com/maddevsio/go-idmatch/log"

	gcfg "gopkg.in/gcfg.v1"
)

type preprocessing struct {
}

type processing struct {
}

type template struct {
	Path string
}

type web struct {
	UploadLimit string
	Uploads     string
	Preview     string
	Static      string
	Templates   string
}

type configFile struct {
	Preprocessing preprocessing
	Processing    processing
	Template      template
	Web           web
}

const defaultConfig = `
	[preprocessing]

	[processing]

	[template]
	path = "templates/json/"
	
	[web]
	uploadLimit = "10M"
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
