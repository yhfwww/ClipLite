.PHONY: all build clean

all: build

build:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-H windowsgui -s -w" -o ClipLite.exe .

build-linux:
	go build -o ClipLite .

clean:
	rm -f ClipLite ClipLite.exe
