BINARY  := inframap-d2
DEMO_CFG := testdata/demo.yml

.PHONY: build test test-v lint demo demo-svg clean help

## build: Compile the binary
build:
	go build -o $(BINARY) .

## test: Run all tests
test:
	go test ./...

## test-v: Run all tests (verbose)
test-v:
	go test -v ./...

## lint: Run golangci-lint
lint:
	golangci-lint run

## demo: Build and generate demo.d2 from bundled testdata
demo: build
	./$(BINARY) generate -c $(DEMO_CFG) -o demo.d2
	@echo "Generated demo.d2 — open it or run 'make demo-svg' to render"

## demo-svg: Generate demo.d2 and render to SVG (requires d2)
demo-svg: demo
	d2 demo.d2 demo.svg
	@echo "Generated demo.svg"

## clean: Remove binary and demo artifacts
clean:
	rm -f $(BINARY) demo.d2 demo.svg

## help: Show available targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
