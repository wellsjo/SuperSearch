.PHONY: all

all: search install

search:
	go build -o bin/ss main.go

install: search
	cp bin/ss /usr/local/bin
