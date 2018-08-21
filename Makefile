.PHONY: all
all: build dist

.PHONY:
build:
	go build --race -ldflags "-X github.com/wellsjo/SuperSearch/search.debug=true" -o bin/ss main.go

dist:
	go build -ldflags "-X github.com/wellsjo/SuperSearch/search.debug=false" -o bin/dist main.go

.PHONY: test
test:
	go test -v ./...

.PHONY: install
install: search
	cp bin/ss /usr/local/bin

clean:
	rm -rf bin
