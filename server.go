package main

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"io"
	"net/http"
	"text/template"
	//db "skywire_uptime/database"
	//scrape "skywire_uptime/scrape"
)

// WhichPage type represents which page the templating engine is rendering
type WhichPage int

// The different pages which we render
const (
	HomePage        WhichPage = iota // Home page
	PreferencesPage                  // Preferences page
	ErrorPage                        // Error page
)

// BasePageStruct is passed to the home page rendering engine
type BasePageStruct struct {
	Page     WhichPage
	LoggedIn bool
	Message  string
}
type NodeRequest struct {
	PublicKey string `form:"publicKey"`
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func HomeHandler(c echo.Context) error {
	homePage := BasePageStruct{HomePage, false, ""}
	r := c.Request()
	URI := r.RequestURI
	//trim the /? characters from URI
	URI = URI[2:]

	//TODO remove first two characters
	//TODO query db
	//TODO print stats

	return c.Render(http.StatusOK, "base-vcenter", homePage)
}
func NodeRequestHandler(c echo.Context) error {
	r := c.Request()
	r.ParseForm()
	n := new(NodeRequest)

	if err := c.Bind(n); err != nil {
		return err
	}

	//publicKey := c.FormValue("publicKey")

	//homePage := BasePageStruct{HomePage, false, ""}
	//return c.Render(http.StatusOK, "base-vcenter", homePage)
	return c.JSON(http.StatusOK, n)
}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/", "public")

	t := &Template{
		templates: template.Must(template.New("hello").ParseFiles(
			"public/templates/home.tmpl",
			"public/templates/header.tmpl",
			"public/templates/footer.tmpl",
			"public/templates/base.tmpl")),
	}

	// Open database
	//dbc = db.ConnectDB(dbFile)
	//defer dbc.Close()

	e.Renderer = t
	e.POST("/search", NodeRequestHandler)
	e.GET("/", HomeHandler)
	e.Logger.Fatal(e.Start(":8080"))
	//fmt.Printf("%v+", scrape.ScrapeSkywireNodes())
}
