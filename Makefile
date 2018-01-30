.PHONY: all

all: search

search:
	go build -o bin/ss main.go

install: search
	cp bin/ss /usr/local/bin
