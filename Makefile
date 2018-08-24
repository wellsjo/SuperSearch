.PHONY: all
all: build dist

.PHONY:
build:
	go build --race -ldflags "-X github.com/wellsjo/SuperSearch/search.debugMode=true" -o bin/ss main.go

dist:
	go build -ldflags "-X github.com/wellsjo/SuperSearch/search.debugMode=false" -o bin/dist main.go

.PHONY: test
test:
	go test -v ./...

.PHONY: install
install: clean dist
	cp bin/dist $$HOME/bin/ss

.PHONY: clean
clean:
	rm -rf bin
