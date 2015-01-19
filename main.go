package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/alext/tablecloth"
)

var (
	eventPort = getenvDefault("PORT", "8080")
)

func getenvDefault(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		val = defaultVal
	}

	return val
}

func main() {
	if wd := os.Getenv("GOVUK_APP_ROOT"); wd != "" {
		tablecloth.WorkingDir = wd
	}

	publicMux := http.NewServeMux()
	publicMux.HandleFunc("/e", ReportHandler)

	log.Println("event-store: listening for events on " + eventPort)

	err := tablecloth.ListenAndServe(fmt.Sprintf(":%v", eventPort), publicMux, "reports")

	if err != nil {
		log.Fatal(err)
	}
}
