.PHONY: all
all: build dist install

.PHONY:
build:
	go build --race -ldflags "-X github.com/wellsjo/SuperSearch/src/log/log.debugMode=true" -o bin/ss main.go

dist:
	go build -ldflags "-X github.com/wellsjo/SuperSearch/src/log/log.debugMode=false" -o bin/dist main.go

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
