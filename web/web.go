package web

import (
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/maddevsio/go-idmatch/config"
	"github.com/maddevsio/go-idmatch/ocr"
	"github.com/maddevsio/go-idmatch/templates"
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
	filename := path.Base(response.Request.URL.String())

	file, err := os.Create(config.Web.Uploads + filename)
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

	return filename, file.Close()
}

func api(c echo.Context) error {
	var filename, url string
	if id, err := c.FormFile("id"); err == nil {
		if filename, err = saveFile(id); err != nil {
			return echo.NewHTTPError(http.StatusUnsupportedMediaType, err.Error())
		}
	} else if url = c.FormValue("url"); len(url) != 0 {
		if filename, err = getFile(url); err != nil {
			return echo.NewHTTPError(http.StatusUnsupportedMediaType, err.Error())
		}
	} else {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data, _ := ocr.Recognize(config.Web.Uploads+filename, "", "")
	response, err := json.Marshal(data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "")
	}
	return c.JSON(http.StatusOK, string(response))
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

	data, idPreview := ocr.Recognize(config.Web.Uploads+id.Filename, template, config.Web.Preview)

	if data == nil || len(data) == 0 {
		data = map[string]interface{}{"error": "Could not recognize document"}
		idPreview = "static/images/empty-contour.png"
	}

	list, err := templates.Load("")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

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
