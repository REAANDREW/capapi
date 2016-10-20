project := capapi
package := github.com/reaandrew/$(project)

capnproto:
	(curl -O https://capnproto.org/capnproto-c++-0.5.3.tar.gz && \
	tar zxf capnproto-c++-0.5.3.tar.gz && \
	cd capnproto-c++-0.5.3 && \
	./configure --prefix=$$HOME && \
	make -j3 && \
	make install)
	go get -u -t zombiezen.com/go/capnproto2/...
	capnp compile -I$$GOPATH/src/zombiezen.com/go/capnproto2/std -ogo capapi.capnp

build: 
	go get -t ./...
	go build

test:
	go test -v ./...


.PHONY: release

release:
		mkdir -p release
		GOOS=linux GOARCH=amd64 go build -o release/$(project)-linux-amd64 $(package)
		GOOS=linux GOARCH=386 go build -o release/$(project)-linux-386 $(package)
		GOOS=windows GOARCH=amd64 go build -o release/$(project)-windows-amd64 $(package)
		GOOS=windows GOARCH=386 go build -o release/$(project)-windows-386 $(package)

.PHONY: capnproto build test release
