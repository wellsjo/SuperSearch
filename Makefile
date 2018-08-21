.PHONY: all
all: search test

.PHONY: search
search:
	go build -o bin/ss main.go

.PHONY: test
test:
	go test -v ./...

.PHONY: install
install: search
	cp bin/ss /usr/local/bin
