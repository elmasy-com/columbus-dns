LDFLAGS = -s
LDFLAGS += -w
LDFLAGS += -X 'main.Version=$(shell git describe --tags --abbrev=0)'
LDFLAGS += -X 'main.Commit=$(shell git rev-list -1 HEAD)'

clean:
	@if [ -f "./columbus-dns" ]; then rm columbus-dns; fi
	@if [ -f "./checksum" ]; then rm checksum; fi  

dev: clean
	@if [ -f "./columbus-dns" ]; then rm columbus-dns; fi  
	go build -o columbus-dns --race -ldflags="$(LDFLAGS)" .

build:
	go build -o columbus-dns -ldflags="$(LDFLAGS)" .

release: clean build
	sha512sum columbus-dns | gpg --clearsign -u daniel@elmasy.com > checksum