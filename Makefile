.PHONY: build test release

build:
	go build -o drift .

test:
	go test ./...

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
