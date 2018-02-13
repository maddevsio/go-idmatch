package main

import (
	"fmt"
	"os"

	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/ocr"
	"github.com/maddevsio/go-idmatch/web"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "go-idmatch"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Maddevsio",
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
						cli.StringFlag{Name: "template", Value: "KG idcard old", Usage: "document template to use"},
						cli.StringFlag{Name: "preview", Usage: "path to export preview image"}},
					Action: func(c *cli.Context) error {
						result, path := ocr.Recognize(c.Args().Get(0), c.String("template"), c.String("preview"))
						for k, v := range result {
							fmt.Printf("%s: %s\n", k, v)
						}
						fmt.Println(path)
						return nil
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
