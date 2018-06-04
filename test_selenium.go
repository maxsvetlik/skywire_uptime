package main

import (
	"fmt"
	"os"
	"time"

	"github.com/tebeka/selenium"
)

// This example shows how to navigate to a http://play.golang.org page, input a
// short program, run it, and inspect its output.
//
// If you want to actually run this example:
//
//   1. Ensure the file paths at the top of the function are correct.
//   2. Remove the word "Example" from the comment at the bottom of the
//      function.
//   3. Run:
//      go test -test.run=Example$ github.com/tebeka/selenium
func Example() {
	// Start a Selenium WebDriver server instance (if one is not already
	// running).
	const (
		// These paths will be different on your system.
		seleniumPath    = "/home/maxwell/go/src/github.com/tebeka/selenium/vendor/selenium-server-standalone-3.8.1.jar"
		geckoDriverPath = "/home/maxwell/go/src/github.com/tebeka/selenium/vendor/geckodriver-v0.19.1-linux64"
		port            = 8080
	)
	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
		selenium.GeckoDriver(geckoDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
		selenium.Output(os.Stderr),            // Output debug information to STDERR.
	}
	selenium.SetDebug(true)
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

	time.Sleep(2000 * time.Millisecond)

	tbl, err := wd.FindElements(selenium.ByXPATH, "//table/tbody/tr/td[2]")

	if err != nil {
		panic(err)
	}

	for _, element := range tbl {
		e_string, err := element.Text()
		if err != nil {
			panic(err)
		}
		// Get the public key
		fmt.Printf("%s\n", e_string)
	}
}

func main() {
	Example()
}
