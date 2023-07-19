LDFLAGS = -s
LDFLAGS += -w
LDFLAGS += -X 'main.Version=$(shell git describe --tags --abbrev=0)'
LDFLAGS += -X 'main.Commit=$(shell git rev-list -1 HEAD)'
LDFLAGS += -extldflags "-static"'

clean:
	@if [ -d "./release" ]; then rm -rf ./release; fi


dev:
	go build -o columbus-dns --race -ldflags="$(LDFLAGS)" .

build-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./release/columbus-dns-linux-amd64 -tags netgo -ldflags="$(LDFLAGS)" .
	
build-linux-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o ./release/columbus-dns-linux-arm64 -tags netgo -ldflags="$(LDFLAGS)" .

build-all: build-linux-amd64 build-linux-arm64

release: clean build-all
	mkdir -p release
	sha512sum ./release/columbus-dns-linux-amd64 >> ./release/checksums
	sha512sum ./release/columbus-dns-linux-arm64 >> ./release/checksums
	cat ./release/checksums | gpg --clearsign -u daniel@elmasy.com > ./release/checksums.signed
	mv ./release/checksums.signed ./release/checksums
	cp columbus-dns.conf ./release/columbus-dns.conf
	cp columbus-dns.service ./release/columbus-dns.service