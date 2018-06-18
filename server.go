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
	scrape "skywire_uptime/scrape"
	"strconv"
	"strings"
	"text/template"
	"time"
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
	NetworkNodes   string
	TimeSinceLast  string
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
var dbFile = "./database/testing.db"

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	return fmt.Sprintf("%02d hr %02d min ago", h, m)
}

func HomeHandler(c echo.Context) error {
	homePage := BasePageStruct{HomePage, false, "", "Now", "0", "Online", "0", "unknown", ""}
	r := c.Request()
	URI := r.RequestURI
	search, search_err := dbc.GetLastSearch()
	if search_err == nil {
		homePage.NetworkNodes = strconv.Itoa(search.NumNodesOnline)
		homePage.TimeSinceLast = fmtDuration(time.Now().Sub(search.Timestamp))
	}
	if len(URI) > 1 {
		//trim the /? characters from URI
		URI = URI[2:]

		if len(URI) > 3 {
			homePage.IsSearching = true
			reqNode, err := dbc.GetNodeByKey(strings.ToLower(URI))
			search, search_err := dbc.GetLastSearch()
			if err == db.ErrNodeNotFound {
				homePage.PublicKey = URI
				homePage.FirstTimeSeen = "No data"
				homePage.AvgTotalUptime = "No data"
				homePage.CurrentStatus = "Offline"
			} else {
				homePage.PublicKey = URI
				fts := reqNode.FirstSeen
				homePage.FirstTimeSeen = fts.Format("2006-01-02")
				homePage.CurrentStatus = "Not yet implemented"  //TODO
				homePage.AvgTotalUptime = "Not yet implemented" //TODO

				// Get last search time
				if search_err == nil {
					time_diff := search.Timestamp.Sub(reqNode.LastSeen)
					if time_diff <= 600 {
						homePage.CurrentStatus = "Online"
					} else {
						homePage.CurrentStatus = "Offline"
					}

					totalPulses, err := dbc.GetPingsSinceCreation(reqNode.FirstSeen)
					if err == nil {
						total := 100.0
						if totalPulses > 0 {
							total = (float64(reqNode.TimesSeen) / float64(totalPulses)) * 100.0
						}
						homePage.AvgTotalUptime = strconv.FormatFloat(total, 'f', -1, 64) + "%"
					}
				}
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

func scrapeOnTimer(db *db.DbConn) {
	delay := 10 * time.Minute
	stop := make(chan bool)
	for {
		err := scrape.QueryNetworkToDB(db)
		if err != nil {
			return
		}
		select {
		case <-time.After(delay):
		case <-stop:
			return
		}
	}
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
	//err := scrape.QueryNetworkToDB(dbc)
	//if err != nil {
	//	fmt.Printf("Issue adding to db")
	//}
}
