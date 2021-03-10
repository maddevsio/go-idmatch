package main

import (
	"fmt"
	"os"

	"github.com/LibertusDio/go-idmatch/log"
	"github.com/LibertusDio/go-idmatch/ocr"
	"github.com/LibertusDio/go-idmatch/ocr/preprocessing"
	"github.com/LibertusDio/go-idmatch/web"
	"github.com/urfave/cli"
)

const flagNameCSFolderPath = "folder"
const flagNameTemplate = "template"
const flagOldKgId = "KG idcard old"

func main() {
	app := cli.NewApp()
	app.Name = "go-idmatch"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "LibertusDio",
			Email: "rock@maddevs.io",
		},
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "debug, d"},
	}

	app.Before = func(c *cli.Context) error {
		if c.Bool("d") {
			log.SetLevel(log.DebugLevel)
		}
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name: "service",
			Action: func(c *cli.Context) error {
				preprocessing.InitCache()
				web.Service()
				return nil
			},
		},
		{
			Name: "ocr",
			Subcommands: []cli.Command{
				{
					Name:  "image",
					Usage: "send the image to ocr recognition",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "front", Usage: "document frontside image"},
						cli.StringFlag{Name: "back", Usage: "document backside image"},
						cli.StringFlag{Name: "template", Usage: "document template to use"},
						cli.StringFlag{Name: "preview", Usage: "path to export preview image"}},
					Action: func(c *cli.Context) error {
						result, _ := ocr.Recognize(c.String("front"), c.String("back"), c.String("template"), c.String("preview"))
						for k, v := range result {
							fmt.Printf("%s: %s\n", k, v)
						}
						return nil
					},
				},
			},
		},
		{
			Name:  "check_solution",
			Usage: "Send folder with images and related json files to ocr recognition with match percentage calculating",
			Flags: []cli.Flag{
				cli.StringFlag{Name: flagNameCSFolderPath, Usage: "Path to data folder"},
				cli.StringFlag{Name: flagNameTemplate, Value: flagOldKgId, Usage: "document template to use"},
			},
			Action: func(c *cli.Context) error {
				dir := c.String(flagNameCSFolderPath)
				if len(dir) == 0 {
					fmt.Println("Provide \"--folder\" flag please")
					return nil
				}
				return ocr.CheckSolution(dir, c.String(flagNameTemplate))
			},
		},
	}

	app.Run(os.Args)
}
