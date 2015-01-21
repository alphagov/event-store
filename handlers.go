package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/validator.v2"
)

var (
	mgoSession      *mgo.Session
	mgoSessionOnce  sync.Once
	mgoDatabaseName = getenvDefault("EVENT_STORE_MONGO_DB", "event_store")
	mgoURL          = getenvDefault("EVENT_STORE_MONGO_URL", "localhost")
)

type CspReport struct {
	Details    CspDetails `json:"csp-report" bson:"csp_report"`
	ReportTime time.Time  `bson:"date_time"`
}

// - DocumentUri: a GOV.UK preview, staging or production URL
// - Referrer, BlockedUri: may be blank
// - ViolatedDirective: a content security policy which can
//   at its most complex contain these characters:
//     default-src: 'self' https://0.example.com *.gov.uk;
type CspDetails struct {
	DocumentUri       string `json:"document-uri" bson:"document_uri" validate:"min=1,max=200,regexp=^https://www(\\.preview\\.alphagov\\.co|-origin\\.production\\.alphagov\\.co|\\.gov)\\.uk/[^\\s]*$"`
	Referrer          string `json:"referrer" bson:"referrer" validate:"max=200"`
	BlockedUri        string `json:"blocked-uri" bson:"blocked_uri" validate:"max=200"`
	ViolatedDirective string `json:"violated-directive" bson:"violated_directive" validate:"min=1,max=200,regexp=^[a-z0-9 '/\\*\\.:;-]+$"`
}

// ReportHandler receives JSON from a request body
func ReportHandler(session *mgo.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var report CspReport

		if req.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			w.Header().Set("Allow", "POST")
			return
		}

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(body, &report); err != nil {
			http.Error(w, "Error parsing JSON", http.StatusBadRequest)
			return
		}

		report.ReportTime = time.Now().UTC()

		if validationError := validator.Validate(report); validationError != nil {
			log.Println("Request failed validation:", validationError)
			log.Println("Failed with report:", report)
			http.Error(w, "Unable to validate JSON", http.StatusBadRequest)
			return
		}

		collection := session.DB(mgoDatabaseName).C("reports")

		if err = collection.Insert(report); err != nil {
			panic(err)
		}

		w.Write([]byte("JSON received"))
	}
}

type Status struct {
	MongoDB bool `json:"mongo"`
}

func HealthcheckHandler(session *mgo.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			w.Header().Set("Allow", "GET")
			return
		}

		var status Status
		var responseCode int

		if session.Ping() == nil {
			status.MongoDB = true
			responseCode = http.StatusOK
		} else {
			status.MongoDB = false
			responseCode = http.StatusServiceUnavailable
		}

		w.WriteHeader(responseCode)

		encoder := json.NewEncoder(w)
		err := encoder.Encode(&status)

		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot encode healthcheck response: %v", err), http.StatusInternalServerError)
		}
	}
}
