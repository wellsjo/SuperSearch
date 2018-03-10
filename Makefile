.PHONY: all

all: search

search:
	go build -o bin/ss main.go

test:
	go test src/gitignore

install: search
	cp bin/ss /usr/local/bin
