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
	(cd tests && go test -v ./...)

.PHONY: capnproto build test
