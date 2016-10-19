capnproto:
	(curl -O https://capnproto.org/capnproto-c++-0.5.3.tar.gz && \
	tar zxf capnproto-c++-0.5.3.tar.gz && \
	cd capnproto-c++-0.5.3 && \
	./configure --prefix=$$HOME && \
	make -j3 && \
	make install)
	go get -u -t zombiezen.com/go/capnproto2/...
	(cd capability && \
		capnp compile -I$$GOPATH/src/zombiezen.com/go/capnproto2/std -ogo capapi.capnp)

build: 
	go build

test:
	(cd tests && go test -v ./...)

install: capnproto
	go get -t ./...


.PHONY: capnproto build test install
