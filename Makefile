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
	git tag $(VERSION)
	git push origin $(VERSION)
