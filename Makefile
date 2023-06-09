LDFLAGS = -s
LDFLAGS += -w
LDFLAGS += -X 'main.Version=$(shell git describe --tags --abbrev=0)'
LDFLAGS += -X 'main.Commit=$(shell git rev-list -1 HEAD)'
LDFLAGS += -extldflags "-static"'

clean:
	@if [ -f "./columbus-dns" ]; then rm columbus-dns; fi
	@if [ -f "./columbus-dns-linux-amd64" ]; then rm columbus-dns-linux-amd64; fi
	@if [ -f "./columbus-dns-linux-arm64" ]; then rm columbus-dns-linux-arm64 ; fi
	@if [ -f "./checksums" ]; then rm checksums; fi

dev:
	go build -o columbus-dns --race -ldflags="$(LDFLAGS)" .

build-linux-amd64:
	GOOS=linux   GOARCH=amd64 go build -o columbus-dns-linux-amd64  -tags netgo -ldflags="$(LDFLAGS)" .

build-linux-arm64:
	GOOS=linux   GOARCH=arm64 go build -o columbus-dns-linux-arm64  -tags netgo -ldflags="$(LDFLAGS)" .

build-all: build-linux-amd64 build-linux-arm64

release: clean build-all
	sha512sum columbus-dns-linux-amd64 >> checksums
	sha512sum columbus-dns-linux-arm64 >> checksums
	cat checksums | gpg --clearsign -u daniel@elmasy.com > checksums.signed
	mv checksums.signed checksums