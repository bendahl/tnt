.Phony: all clean test

all: clean test tnt

clean:
	rm -f tnt

tnt:
	go build -ldflags "-X main.version=$$(git describe --abbrev=0 --always) -X main.commit=$$(git rev-parse HEAD)"

test:
	go test ./...
