package scrape

import (
	"fmt"
	//"io/ioutil"
	"database/sql"
	"github.com/tebeka/selenium"
	db "skywire_uptime/database"
	"time"
)

func ScrapeSkywireNodes() []string {
	fmt.Printf("Beginning Skywire scrape")
	const (
		seleniumPath    = "/home/maxwell/go/src/github.com/tebeka/selenium/vendor/selenium-server-standalone-3.8.1.jar"
		geckoDriverPath = "/home/maxwell/go/src/github.com/tebeka/selenium/vendor/geckodriver-v0.19.1-linux64"
		port            = 8080
	)
	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
		selenium.GeckoDriver(geckoDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
		//selenium.Output(ioutil.Discard),       // Output debug information to STDERR.
	}
	selenium.SetDebug(false)
	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}
	defer service.Stop()

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "firefox"}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		panic(err)
	}
	defer wd.Quit()

	// Navigate to the simple playground interface.
	if err := wd.Get("http://discovery.skycoin.net:8001"); err != nil {
		panic(err)
	}

	// Expand keys on button press
	btn, err := wd.FindElement(selenium.ByCSSSelector, "a")
	if err != nil {
		panic(err)
	}
	if err := btn.Click(); err != nil {
		panic(err)
	}

	//TODO add checks for if page loading successful.
	// if unsuccessful, do not add search to db

	// TODO handle delay in node listing better than waiting 5 seconds
	time.Sleep(5000 * time.Millisecond)

	tbl, err := wd.FindElements(selenium.ByXPATH, "//table/tbody/tr/td[2]")

	if err != nil {
		panic(err)
	}

	var nodeList []string
	fmt.Printf("Starting key collection")
	for _, element := range tbl {
		e_string, err := element.Text()
		if err != nil {
			panic(err)
		}
		// Get the public key
		nodeList = append(nodeList, e_string)
	}
	return nodeList
}

// Scrapes network and adds to database(s)
func QueryNetworkToDB(dbc *db.DbConn) error {
	nodeList := ScrapeSkywireNodes()

	// If search was succesful, add to search database
	_, err := dbc.InsertSearch(len(nodeList), time.Now())

	if err != nil {
		fmt.Printf("Error inserting search into db")
		return err
	}

	now := time.Now()
	for _, nodelet := range nodeList {
		//if node doesn't exist
		_, err := dbc.GetNodeByKey(nodelet)
		if err == sql.ErrNoRows {
			_, err := dbc.InsertNode(nodelet, now, now, 1)
			if err != nil {
				fmt.Printf("Error inserting scraped public key into db: %v\n", err)
			}
		} else {
			//if node exists
			err := dbc.UpdateNode(nodelet, time.Now())
			if err != nil {
				fmt.Printf("Error updating scraped public key into db.\n")
			}
		}
	}

	return nil
}
