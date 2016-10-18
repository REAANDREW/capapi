capnp:
	capnp compile -I$GOPATH/src/zombiezen.com/go/capnproto2/std -ogo capapi.capnp

build: capnp
	go build

test:
	go test -v ./...

install:
	go get -u -t zombiezen.com/go/capnproto2/...
	go get -t ./...


.PHONY: capnp build test install
