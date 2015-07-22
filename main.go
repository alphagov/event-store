package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alext/tablecloth"
	"gopkg.in/mgo.v2"
)

var (
	eventPort  = getenvDefault("PORT", "3097")
	mgoNodes   = getenvDefault("EVENT_STORE_MONGO_NODES", "localhost")
	mgoSession *mgo.Session
)

func ConnectToMongo(hostname string) (*mgo.Session, error) {
	session, err := mgo.DialWithTimeout(hostname, 200*time.Millisecond)
	if err != nil {
		return nil, err
	}

	// Queries return "read tcp 127.0.0.1:27017: i/o timeout" unless
	// the session socket timeout is increased.
	session.SetSocketTimeout(1 * time.Second)

	session.SetMode(mgo.Strong, true)

	return session, nil
}

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

	mgoSession, err := ConnectToMongo(mgoNodes)
	if err != nil {
		log.Fatal("Error connectiong to mongo : ", err)
	}

	if err != nil {
		log.Fatal(err)
	}

	publicMux := http.NewServeMux()
	publicMux.HandleFunc("/e", ReportHandler(mgoSession))
	publicMux.HandleFunc("/healthcheck", HealthcheckHandler(mgoSession))

	log.Println("event-store: listening for events on " + eventPort)

	err = tablecloth.ListenAndServe(fmt.Sprintf(":%v", eventPort), publicMux, "reports")

	if err != nil {
		log.Fatal(err)
	}
}
