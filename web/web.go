package web

import (
	"html/template"
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/tzununbekov/go-idmatch/ocr"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func landing(c echo.Context) error {
	return c.Render(http.StatusOK, "landing", "")
}

func result(c echo.Context) error {
	file, err := c.FormFile("id")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create("web/uploads/" + file.Filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	data, preview := ocr.Recognize("web/uploads/"+file.Filename, "KG idcard old", "web/preview")

	return c.Render(http.StatusOK, "landing", map[string]interface{}{
		"data":    data,
		"preview": preview,
	})
}

func Service() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/static", "web/static")
	e.Static("web/preview", "web/preview")

	t := &Template{
		templates: template.Must(template.ParseGlob("web/templates/idmatch_landing.html")),
	}
	e.Renderer = t

	e.GET("/", landing)
	e.POST("/", result)

	e.Logger.Fatal(e.Start(":8080"))
}
