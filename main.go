package main

import (
	"os"

	"github.com/tzununbekov/go-idmatch/ocr"
	"github.com/tzununbekov/go-idmatch/web"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "go-idmatch"
	app.Commands = []cli.Command{
		{
			Name: "service",
			Action: func(c *cli.Context) error {
				web.Service()
				return nil
			},
		},
		{
			Name: "image",
			Action: func(c *cli.Context) error {
				ocr.Recognize(c.Args().First(), "KG idcard old")
				return nil
			},
		},
	}

	app.Run(os.Args)
}
