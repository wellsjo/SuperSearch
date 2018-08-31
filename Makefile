.PHONY: all
all: build dist install

.PHONY:
build:
	go build --race -o bin/ss main.go

dist:
	go build -o bin/dist main.go

.PHONY: test
test:
	go test -v ./...

bench:
	go test -bench "Search" ./...

.PHONY: install
install: dist
	cp bin/dist $$HOME/bin/ss

.PHONY: clean
clean:
	rm -rf bin
