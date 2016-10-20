package main

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthorizationParser(t *testing.T) {
	Convey("Parsing authorization header", t, func() {

		Convey("Returns err when not present in the header", func() {
			req, _ := http.NewRequest("GET", "http://localhost:8080", nil)
			_, err := ParseAuthorization(req)
			So(err, ShouldEqual, ErrNoAuthorizationHeader)
		})

		Convey("Returns err when does not contain Bearer", func() {
			req, _ := http.NewRequest("GET", "http://localhost:8080", nil)
			req.Header.Set("Authorization", "something")
			_, err := ParseAuthorization(req)
			So(err, ShouldEqual, ErrMalformedAuthorizationHeader)
		})

		Convey("Returns err when does not contain value", func() {
			req, _ := http.NewRequest("GET", "http://localhost:8080", nil)
			req.Header.Set("Authorization", "Bearer")
			_, err := ParseAuthorization(req)
			So(err, ShouldEqual, ErrNoAPIKey)
		})

		Convey("Returns API Key", func() {
			req, _ := http.NewRequest("GET", "http://localhost:8080", nil)
			req.Header.Set("Authorization", "Bearer 1234")
			apiKey, _ := ParseAuthorization(req)
			So(apiKey, ShouldEqual, "1234")
		})
	})
}
