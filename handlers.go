package main

import (
	"encoding/json"
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
// - ViolatedDirective, OriginalPolicy: one or more CSP policies which can
//   at their most complex contain these characters:
//     default-src: 'self' https://0.example.com *.gov.uk;
type CspDetails struct {
	DocumentUri       string `json:"document-uri" bson:"document_uri" validate:"min=1,max=200,regexp=^https://www(\\.preview\\.alphagov\\.co|-origin\\.production\\.alphagov\\.co|\\.gov)\\.uk/[^\\s]*$"`
	Referrer          string `json:"referrer" bson:"referrer" validate:"max=200"`
	BlockedUri        string `json:"blocked-uri" bson:"blocked_uri" validate:"max=200"`
	ViolatedDirective string `json:"violated-directive" bson:"violated_directive" validate:"min=1,max=200,regexp=^[a-z0-9 '/\\*\\.:;-]+$"`
	OriginalPolicy    string `json:"original-policy" bson:"original_policy" validate:"min=1,max=200"`
}

func getMgoSession() *mgo.Session {
	mgoSessionOnce.Do(func() {
		var err error

		mgoSession, err = mgo.Dial(mgoURL)
		if err != nil {
			panic(err)
		}
	})

	return mgoSession.Clone()
}

func storeCspReport(report CspReport) {
	session := getMgoSession()
	defer session.Close()
	session.SetMode(mgo.Strong, true)

	collection := session.DB(mgoDatabaseName).C("reports")

	err := collection.Insert(report)

	if err != nil {
		panic(err)
	}
}

// ReportHandler receives JSON from a request body
func ReportHandler(w http.ResponseWriter, req *http.Request) {
	var err error
	var newCspReport CspReport

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

	err = json.Unmarshal(body, &newCspReport)

	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	newCspReport.ReportTime = time.Now().UTC()

	if validationError := validator.Validate(newCspReport); validationError != nil {
		log.Println("Request failed validation:", validationError)
		log.Println("Failed with report:", newCspReport)
		http.Error(w, "Unable to validate JSON", http.StatusBadRequest)
		return
	}

	go storeCspReport(newCspReport)

	w.Write([]byte("JSON received"))
}
