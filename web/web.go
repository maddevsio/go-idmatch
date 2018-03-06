package web

import (
	"errors"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/maddevsio/go-idmatch/config"
	"github.com/maddevsio/go-idmatch/ocr"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func saveFile(file *multipart.FileHeader) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	buff := make([]byte, 512)
	if _, err := src.Read(buff); err != nil {
		return err
	}

	if format := http.DetectContentType(buff); !strings.HasPrefix(format, "image/") {
		return errors.New("Unsupported file format")
	}

	dst, err := os.Create(config.Web.Uploads + file.Filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := src.Seek(0, 0); err != nil {
		return err
	}

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}

func landing(c echo.Context) error {
	return c.Render(http.StatusOK, "landing", "")
}

func result(c echo.Context) error {
	var facePreview string

	face, err := c.FormFile("face")
	if face.Size != 0 {
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		if err := saveFile(face); err != nil {
			return echo.NewHTTPError(http.StatusUnsupportedMediaType, err.Error())
		}
		facePreview = config.Web.Uploads + face.Filename
	}

	id, err := c.FormFile("id")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if err := saveFile(id); err != nil {
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, err.Error())
	}

	data, idPreview := ocr.Recognize(config.Web.Uploads+id.Filename, "KG idcard old", config.Web.Preview)

	if data == nil || len(data) == 0 {
		data = map[string]interface{}{"error": "Could not recognize document"}
		idPreview = "static/images/empty-contour.png"
	}

	return c.Render(http.StatusOK, "landing", map[string]interface{}{
		"data":         data,
		"id_preview":   idPreview,
		"face_preview": facePreview,
	})
}

func Service() {
	os.MkdirAll(config.Web.Uploads, os.ModePerm)
	os.MkdirAll(config.Web.Preview, os.ModePerm)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.BodyLimit(config.Web.UploadLimit))
	e.Use(middleware.Recover())
	e.Static("/static", config.Web.Static)
	e.Static("web/uploads", config.Web.Uploads)
	e.Static("web/preview", config.Web.Preview)

	t := &Template{
		templates: template.Must(template.ParseGlob(config.Web.Templates + "/idmatch_landing.html")),
	}
	e.Renderer = t

	e.GET("/", landing)
	e.POST("/", result)

	e.Logger.Fatal(e.Start(":8080"))
}
