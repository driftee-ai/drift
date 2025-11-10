.PHONY: build test release dist clean-dist lint integration-test

build:
	go build -o drift .

test:
	go test -count=1 ./...

integration-test:
	go test -v -tags=integration -count=1 ./...

release: test
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make release VERSION=vX.Y.Z"; \
		exit 1; \
	fi
	@if [ "$(shell git rev-parse --abbrev-ref HEAD)" != "main" ]; then \
		echo "Error: You must be on the main branch to make a release."; \
		exit 1; \
	fi
	@git fetch origin
	@if [ "$(shell git rev-parse HEAD)" != "$(shell git rev-parse @{u})" ]; then \
		echo "Error: Your local main branch is not up-to-date with the remote."; \
		exit 1; \
	fi
	git tag $(VERSION)
	git push origin $(VERSION)

DIST_DIR := dist
APP_NAME := drift

dist: clean-dist
	@echo "Building binaries for multiple platforms..."
	@mkdir -p $(DIST_DIR)

	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 .
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build -o $(DIST_DIR)/$(APP_NAME)-linux-arm64 .

	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build -o $(DIST_DIR)/$(APP_NAME)-darwin-amd64 .
	# macOS ARM64
	GOOS=darwin GOARCH=arm64 go build -o $(DIST_DIR)/$(APP_NAME)-darwin-arm64 .

	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build -o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe .
	# Windows ARM64
	GOOS=windows GOARCH=arm64 go build -o $(DIST_DIR)/$(APP_NAME)-windows-arm64.exe .

	@echo "Binaries built in $(DIST_DIR)/"

clean-dist:
	@echo "Cleaning $(DIST_DIR)/ directory..."
	@rm -rf $(DIST_DIR)

lint:
	$(shell go env GOPATH)/bin/golangci-lint run ./...