package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"io"
	"net/http"
	db "skywire_uptime/database"
	"text/template"
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
	Page           WhichPage
	IsSearching    bool
	PublicKey      string
	FirstTimeSeen  string
	AvgTotalUptime string
	CurrentStatus  string
	Message        string
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

var dbc *db.DbConn
var dbFile = "testing"

func HomeHandler(c echo.Context) error {
	homePage := BasePageStruct{HomePage, false, "", "Now", "0", "Online", ""}
	r := c.Request()
	URI := r.RequestURI
	if len(URI) > 1 {
		//trim the /? characters from URI
		URI = URI[2:]

		if len(URI) > 3 {
			homePage.IsSearching = true
			//homePage.PublicKey = URI
			//homePage.FirstTimeSeen = time.Now()
			//homePage.AvgTotalUptime = "100%"
			//homePage.CurrentStatus = "Online"
			reqNode, err := dbc.GetNodeByKey(URI)
			if err == db.ErrNodeNotFound {
				homePage.PublicKey = URI
				homePage.FirstTimeSeen = "N/A"
				homePage.AvgTotalUptime = "N/A"
				homePage.CurrentStatus = "Node not found"
				fmt.Printf("No node found \n")
			} else {
				homePage.PublicKey = URI
				fts := reqNode.FirstSeen
				year, month, day := fts.Date()
				homePage.FirstTimeSeen = month.String() + "/" + string(day) + "/" + string(year)
				homePage.CurrentStatus = "Not yet implemented"  //TODO
				homePage.AvgTotalUptime = "Not yet implemented" //TODO
			}
		} else {
			homePage.PublicKey = URI
			homePage.FirstTimeSeen = "N/A"
			homePage.AvgTotalUptime = "N/A"
			homePage.CurrentStatus = "Public key misformed"
		}
	}

	return c.Render(http.StatusOK, "base-vcenter", homePage)
}
func NodeRequestHandler(c echo.Context) error {
	r := c.Request()
	r.ParseForm()
	n := new(NodeRequest)

	if err := c.Bind(n); err != nil {
		return err
	}
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
	dbc = db.ConnectDB(dbFile)
	defer dbc.Close()

	e.Renderer = t
	e.POST("/search", NodeRequestHandler)
	e.GET("/", HomeHandler)
	e.Logger.Fatal(e.Start(":8080"))
	//fmt.Printf("%v+", scrape.ScrapeSkywireNodes())
}
