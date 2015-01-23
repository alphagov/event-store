package main_test

import (
	. "github.com/alphagov/event-store"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"net/http"
	"net/http/httptest"

	"gopkg.in/mgo.v2"
)

var _ = Describe("Handlers", func() {
	var session *mgo.Session
	var err error
	var mgoDatabaseName = "event_store"

	BeforeEach(func() {
		session, err = ConnectToMongo("localhost")
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		session.DB(mgoDatabaseName).DropDatabase()
		session.Close()
	})

	Describe("ReportHandler", func() {
		It("should return bad request for non-JSON bodies", func() {
			request, _ := http.NewRequest("POST", "/r", bytes.NewBufferString("Hello"))
			response := httptest.NewRecorder()

			ReportHandler(session)(response, request)

			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should only accept documents from GOV.UK", func() {
			payload := `{"csp-report": {
				"document-uri": "https://www.example.com/",
				"blocked-uri": "https://evil.example.com/",
				"violated-directive": "directive",
				"original-policy": "policy"
			}}`

			request, _ := http.NewRequest("POST", "/r", bytes.NewBufferString(payload))
			response := httptest.NewRecorder()

			ReportHandler(session)(response, request)

			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("should store valid reports in the database", func() {
			payload := `{"csp-report": {
				"document-uri": "https://www.gov.uk/page",
				"blocked-uri": "https://evil.example.com/",
				"violated-directive": "directive",
				"original-policy": "policy"
			}}`

			request, _ := http.NewRequest("POST", "/r", bytes.NewBufferString(payload))
			response := httptest.NewRecorder()

			ReportHandler(session)(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))

			collection := session.DB(mgoDatabaseName).C("reports")
			report := CSPReport{}

			Expect(collection.Find(nil).One(&report)).To(BeNil())

			Expect(report.Details.DocumentUri).To(Equal("https://www.gov.uk/page"))
		})
	})

	Describe("HealthcheckHandler", func() {
		It("returns a 200 OK when it pings Mongo", func() {
			request, _ := http.NewRequest("GET", "/healthcheck", nil)
			response := httptest.NewRecorder()

			HealthcheckHandler(session)(response, request)

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body).To(MatchJSON(`{"mongo": true}`))
		})

		It("returns Method Not Allowed when not using a GET", func() {
			request, _ := http.NewRequest("DELETE", "/healthcheck", nil)
			response := httptest.NewRecorder()

			HealthcheckHandler(session)(response, request)

			Expect(response.Code).To(Equal(http.StatusMethodNotAllowed))
		})
	})
})
