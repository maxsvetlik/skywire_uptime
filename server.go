package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"

	sessions "github.com/ipfans/echo-session"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	//"skywire_uptime/database"
	scrape "skywire_uptime/scrape"
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

type Template struct {
	templates *template.Template
}

// Credentials which stores google ids.
type Credentials struct {
	Cid     string `json:"cid"`
	Csecret string `json:"csecret"`
}

// User is a retrieved and authentiacted user.
type User struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Profile       string `json:"profile"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Gender        string `json:"gender"`
}

var cred Credentials
var conf *oauth2.Config
var state string
var store = sessions.NewCookieStore([]byte("secret"))

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func HomeHandler(c echo.Context) error {
	homePage := BasePageStruct{HomePage, false, ""}
	return c.Render(http.StatusOK, "base-vcenter", homePage)
}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func init_auth() {
	var c Credentials
	file, err := ioutil.ReadFile("./creds.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(file, &c)

	conf = &oauth2.Config{
		ClientID:     c.Cid,
		ClientSecret: c.Csecret,
		RedirectURL:  "http://localhost:9090/auth",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email", // You have to select your own scope from here -> https://developers.google.com/identity/protocols/googlescopes#google_sign-in
		},
		Endpoint: google.Endpoint,
	}
}

func getLoginURL(state string) string {
	// State can be some kind of random generated hash string.
	// See relevant RFC: http://tools.ietf.org/html/rfc6749#section-10.12
	return conf.AuthCodeURL(state)
}

func authHandler(c echo.Context) error {
	// Handle the exchange code to initiate a transport.
	session := sessions.Default(c)
	retrievedState := session.Get("state")
	if retrievedState != c.QueryParam("state") {
		// c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("Invalid session state: %s", retrievedState))
		return nil
	}

	tok, err := conf.Exchange(oauth2.NoContext, c.QueryParam("code"))
	if err != nil {
		// c.AbortWithError(http.StatusBadRequest, err)
		return nil
	}

	client := conf.Client(oauth2.NoContext, tok)
	email, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		// c.AbortWithError(http.StatusBadRequest, err)
		return nil
	}
	defer email.Body.Close()
	data, _ := ioutil.ReadAll(email.Body)
	log.Println("Email body: ", string(data))
	// c.Status(http.StatusOK)
	return nil
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(sessions.Sessions("goquestsession", store))
	e.Static("/", "public")

	t := &Template{
		templates: template.Must(template.New("hello").ParseFiles(
			"public/templates/home.tmpl",
			"public/templates/header.tmpl",
			"public/templates/footer.tmpl",
			"public/templates/base.tmpl")),
	}

	e.Renderer = t
	//ExampleFindElement()
	e.GET("/", HomeHandler)
	e.GET("/login", authHandler)
	//e.Logger.Fatal(e.Start(":8080"))
	fmt.Printf("%v+", scrape.ScrapeSkywireNodes())
}
