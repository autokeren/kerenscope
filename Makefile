.PHONY: build test install clean

build:
	go build -o bin/keren .

test:
	go test ./...

install:
	go install github.com/autokeren/kerenscope

clean:
	rm -rf bin dist
