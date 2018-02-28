package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"path/filepath"

	"github.com/maddevsio/go-idmatch/log"
	"github.com/maddevsio/go-idmatch/ocr"
	"github.com/maddevsio/go-idmatch/web"
	"github.com/urfave/cli"
)

const flagNameCSFolderPath = "folder"
const flagNameTemplate = "template"

//feel free to add something here
var imageFormats = [...]string{
	".jpg",
	".jpeg",
	".png",
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func isImage(ext string) bool {
	for _, f := range imageFormats {
		if strings.Compare(ext, f) == 0 {
			return true
		}
	}
	return false
}

type oldIdCardInfo struct {
	Surname     string `json:"surname"`
	Firstname   string `json:"firstname"`
	Middlename  string `json:"middlename"`
	Gender      string `json:"gender"`
	Inn         string `json:"inn"`
	Birthday    string `json:"birthday"`
	Nationality string `json:"nationality"`
	Serial      string `json:"serial"`
}

func checkSolution(c *cli.Context) error {
	folderPath := c.String(flagNameCSFolderPath)
	fmt.Println(folderPath)
	filepath.Walk(folderPath,
		func(path string, info os.FileInfo, err error) error {
			ext := filepath.Ext(path)
			if info.IsDir() || !isImage(ext) {
				return nil
			}

			fileNameWOExt := strings.TrimSuffix(info.Name(), ext)
			jsonFileName := folderPath + fileNameWOExt + ".json"

			if !fileExists(jsonFileName) {
				return nil
			}

			jsonFileData, err := ioutil.ReadFile(jsonFileName)
			if err != nil {
				fmt.Println(err.Error())
				return nil
			}
			var jsonInfo oldIdCardInfo
			err = json.Unmarshal(jsonFileData, &jsonInfo)
			if err != nil {
				fmt.Println(err.Error())
				return nil
			}

			ocrRes, _ := ocr.Recognize(path, c.String(flagNameTemplate), "")

			var ocrInfo oldIdCardInfo
			str, err := json.MarshalIndent(ocrRes, "", "	")
			if err != nil {
				fmt.Println(err.Error())
				return nil
			}

			fmt.Println(ocrRes)
			fmt.Println("***********")
			fmt.Println(str)
			err = json.Unmarshal([]byte(str), &ocrRes)
			if err != nil {
				fmt.Println(err.Error())
				return nil
			}

			fmt.Println(ocrInfo.Firstname)
			fmt.Println(jsonInfo.Firstname)

			return nil
		})
	return nil
}

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
		{
			Name:  "check_solution",
			Usage: "Send folder with images and related json files to ocr recognition with match percentage calculating",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: flagNameCSFolderPath, Usage: "Path to data folder",
				},
				cli.StringFlag{
					Name: flagNameTemplate, Value: "KG idcard old", Usage: "document template to use",
				},
			},
			Action: checkSolution,
		},
	}

	app.Run(os.Args)
}
