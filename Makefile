.PHONY: build run test clean

BINARY := event-store
BUILDFILES := main.go handlers.go

build: $(BINARY)

run: _vendor
	gom run $(BUILDFILES)

test: _vendor
	gom test

clean:
	rm -f $(BINARY)

_vendor: Gomfile
	gom install
	touch _vendor

$(BINARY): _vendor $(BUILDFILES)
	gom build -o $(BINARY) $(BUILDFILES)
