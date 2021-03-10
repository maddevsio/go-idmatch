package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/LibertusDio/go-idmatch/config"
	"github.com/LibertusDio/go-idmatch/ocr"
	"github.com/LibertusDio/go-idmatch/templates"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func saveFile(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	buff := make([]byte, 512)
	if _, err = src.Read(buff); err != nil {
		return "", err
	}

	if format := http.DetectContentType(buff); !strings.HasPrefix(format, "image/") {
		return "", errors.New("Unsupported file format")
	}

	dst, err := os.Create(config.Web.Uploads + file.Filename)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err = src.Seek(0, 0); err != nil {
		return "", err
	}

	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	return file.Filename, nil
}

// TODO: Merge this func with saveFile
func getFile(url string) (string, error) {
	limit := int64(10485760)
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	frontside := path.Base(response.Request.URL.String())

	file, err := os.Create(config.Web.Uploads + frontside)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(&io.LimitedReader{R: response.Body, N: limit})
	if err != nil {
		return "", err
	}

	if format := http.DetectContentType(body); !strings.HasPrefix(format, "image/") {
		return "", errors.New("Unsupported file format")
	}

	n, err := file.Write(body)
	if err != nil {
		return "", err
	}

	if int64(n) == limit {
		return "", errors.New("Image is too big")
	}

	return frontside, file.Close()
}

func api(c echo.Context) error {
	var frontside, backside string

	if id, err := c.FormFile("id"); err == nil {
		if frontside, err = saveFile(id); err != nil {
			return echo.NewHTTPError(http.StatusUnsupportedMediaType, err.Error())
		}
	} else if url := c.FormValue("url"); len(url) != 0 {
		if frontside, err = getFile(url); err != nil {
			return echo.NewHTTPError(http.StatusUnsupportedMediaType, err.Error())
		}
	} else if front, err := c.FormFile("front"); err == nil {
		if frontside, err = saveFile(front); err != nil {
			return echo.NewHTTPError(http.StatusUnsupportedMediaType, err.Error())
		}
		if back, err := c.FormFile("back"); err == nil {
			if backside, err = saveFile(back); err != nil {
				return echo.NewHTTPError(http.StatusUnsupportedMediaType, err.Error())
			}
		}
	} else {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	result := getElectedResult(frontside, backside, c.FormValue("template"))

	response, err := json.Marshal(result)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "")
	}
	fmt.Println(string(response))
	return c.JSONPretty(http.StatusOK, json.RawMessage(string(response)), "   ")
}

func landing(c echo.Context) error {
	list, err := templates.Load("")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.Render(http.StatusOK, "landing", map[string]interface{}{
		"templates": list,
	})
}

func getElectedResult(frontside, backside, templ string) map[string]interface{} {
	var wg sync.WaitGroup
	resChan := make(chan map[string]interface{}, 3)
	results := make(map[string]map[interface{}]int)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, _ := ocr.Recognize(config.Web.Uploads+frontside, config.Web.Uploads+backside, templ, "")
			resChan <- data
		}()
	}
	wg.Wait()
	close(resChan)

	for data := range resChan {
		for field, value := range data {
			if _, ok := results[field]; ok {
				if _, ok := results[field][value]; ok {
					results[field][value]++
				} else {
					results[field][value] = 1
				}
			} else {
				results[field] = map[interface{}]int{
					value: 1,
				}
			}
		}
	}

	result := make(map[string]interface{})

	for field, options := range results {
		max := 1
		for value, weight := range options {
			if weight >= max {
				result[field] = value
				max = weight
			}
		}
	}

	fmt.Printf("Results: %+v\n", results)
	fmt.Printf("Result: %+v\n", result)

	return result
}

func result(c echo.Context) error {
	var facePreview string

	template := c.FormValue("template")

	face, err := c.FormFile("face")
	if face != nil && face.Size != 0 {
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		if _, err = saveFile(face); err != nil {
			return echo.NewHTTPError(http.StatusUnsupportedMediaType, err.Error())
		}
		facePreview = config.Web.Uploads + face.Filename
	}

	id, err := c.FormFile("id")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if _, err = saveFile(id); err != nil {
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, err.Error())
	}

	data, idPreview := ocr.Recognize(config.Web.Uploads+id.Filename, "", template, config.Web.Preview)

	if data == nil || len(data) == 0 {
		data = map[string]interface{}{"error": "Could not recognize document"}
		idPreview = []string{"static/images/empty-contour.png"}
	}

	list, err := templates.Load("")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	fmt.Println(data)

	return c.Render(http.StatusOK, "landing", map[string]interface{}{
		"templates":    list,
		"data":         data,
		"id_preview":   idPreview,
		"face_preview": facePreview,
	})
}

func Service() {
	os.MkdirAll(config.Web.Uploads, os.ModePerm)
	os.MkdirAll(config.Web.Preview, os.ModePerm)

	e := echo.New()
	defer e.Close()

	e.Use(middleware.Logger())
	e.Use(middleware.BodyLimit(config.Web.UploadLimit))
	e.Use(middleware.Recover())
	e.Static("/static", config.Web.Static)
	e.Static("web/uploads", config.Web.Uploads)
	e.Static("web/preview", config.Web.Preview)
	e.Static("swagger", "swagger.yml")

	t := &Template{
		templates: template.Must(template.ParseGlob(config.Web.Templates + "/idmatch_landing.html")),
	}
	e.Renderer = t

	e.GET("/", landing)
	e.POST("/", result)
	e.POST("/ocr", api)

	e.Logger.Fatal(e.Start(":8080"))
}
