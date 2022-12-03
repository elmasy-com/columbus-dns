LDFLAGS = -s
LDFLAGS += -w
LDFLAGS += -X 'main.Version=$(shell git describe --tags --abbrev=0)'
LDFLAGS += -X 'main.Commit=$(shell git rev-list -1 HEAD)'
BASE = columbus-dns

dev:
	@if [ -f "./$(BASE)" ]; then rm $(BASE); fi  
	go build -o $(BASE) --race -ldflags="$(LDFLAGS)" .

clean:
	@if [ -f "./$(BASE)" ]; then rm $(BASE); fi
	@if [ -f "./$(BASE)-linux-amd64" ]; then rm $(BASE)-linux-amd64; fi
	@if [ -f "./$(BASE)-linux-arm64" ]; then rm $(BASE)-linux-arm64 ; fi
	@if [ -f "./checksums" ]; then rm checksums; fi

build-linux-amd64:
	GOOS=linux   GOARCH=amd64 go build -o $(BASE)-linux-amd64   -ldflags="$(LDFLAGS)" .

build-linux-arm64:
	GOOS=linux   GOARCH=arm64 go build -o $(BASE)-linux-arm64   -ldflags="$(LDFLAGS)" .

build-all: build-linux-amd64 build-linux-arm64

release: clean build-all
	sha512sum $(BASE)-linux-amd64 >> checksums
	sha512sum $(BASE)-linux-arm64 >> checksums
	cat checksums | gpg --clearsign -u daniel@elmasy.com > checksums.signed
	mv checksums.signed checksums