.PHONY: deps test build

BINARY := event-store
ORG_PATH := github.com/alphagov
REPO_PATH := $(ORG_PATH)/$(BINARY)

all: test build

build: vendor
	gom build -o $(BINARY)

run: build
	./$(BINARY)

test: vendor
	gom test

clean:
	rm -rf bin $(BINARY) _vendor

vendor: deps
	rm -rf _vendor/src/$(ORG_PATH)
	mkdir -p _vendor/src/$(ORG_PATH)
	ln -s $(CURDIR) _vendor/src/$(REPO_PATH)

deps:
	gom install
