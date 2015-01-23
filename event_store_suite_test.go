package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestEventStore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EventStore Suite")
}
