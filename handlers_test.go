package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gopkg.in/mgo.v2"
)

func TestNonJsonBodyReturnsBadRequest(t *testing.T) {
	mgoSession := connectToMongo(t)
	defer mgoSession.DB(mgoDatabaseName).DropDatabase()

	request, _ := http.NewRequest("POST", "/r", bytes.NewBufferString("Hello"))
	response := httptest.NewRecorder()

	ReportHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("Expected bad request status code, received %v", response.Code)
	}
}

func TestDocumentUriMustBeOnGovuk(t *testing.T) {
	mgoSession := connectToMongo(t)
	defer mgoSession.DB(mgoDatabaseName).DropDatabase()

	payload := `{"csp-report": {
		"document-uri": "https://www.example.com/",
		"blocked-uri": "https://evil.example.com/",
		"violated-directive": "directive",
		"original-policy": "policy"
	}}`

	request, _ := http.NewRequest("POST", "/r", bytes.NewBufferString(payload))
	response := httptest.NewRecorder()

	ReportHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("Expected bad request status code, received %v", response.Code)
	}
}

func TestValidReportsAreStored(t *testing.T) {
	mgoSession := connectToMongo(t)
	defer mgoSession.DB(mgoDatabaseName).DropDatabase()

	payload := `{"csp-report": {
		"document-uri": "https://www.gov.uk/page",
		"blocked-uri": "https://evil.example.com/",
		"violated-directive": "directive",
		"original-policy": "policy"
	}}`

	request, _ := http.NewRequest("POST", "/r", bytes.NewBufferString(payload))
	response := httptest.NewRecorder()

	ReportHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Expected OK status code, received %v", response.Code)
	}

	time.Sleep(100 * time.Millisecond) // Allow the goroutine time to store

	collection := mgoSession.DB(mgoDatabaseName).C("reports")
	report := CspReport{}
	err := collection.Find(nil).One(&report)

	if err != nil {
		t.Fatal("Error when retrieving document from MongoDB")
	}

	if report.Details.DocumentUri != "https://www.gov.uk/page" {
		t.Fatalf("Stored report contained unexpected document URI: %v", report.Details.DocumentUri)
	}

}

func connectToMongo(t *testing.T) (session *mgo.Session) {
	session, err := mgo.DialWithTimeout("localhost", 200*time.Millisecond)

	if err != nil {
		t.Fatalf(err.Error())
	}

	// Queries return "read tcp 127.0.0.1:27017: i/o timeout" unless
	// the session socket timeout is increased.
	session.SetSocketTimeout(1 * time.Second)

	return session
}
