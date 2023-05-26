all: build test check

build:

test:
	go test -v ./...

check-lint:
	revive ./...

check: check-lint

clean:

.PHONY: build check test
