package main

import (
	"encoding/json"
	"testing"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

type Something struct {
	Props []string `json:"props"`
}

func TestSpikes(t *testing.T) {

	log.SetLevel(log.ErrorLevel)

	Convey("Spikes", t, func() {
		Convey("Testing my knowledge of Json Serialization in GO", func() {
			jsonBytes := []byte(`[{"props" : ["a","b"]}]`)
			var somethings []Something

			json.Unmarshal(jsonBytes, &somethings)

			So(len(somethings), ShouldEqual, 1)
			So(len(somethings[0].Props), ShouldEqual, 2)
			So(somethings[0].Props[0], ShouldEqual, "a")
			So(somethings[0].Props[1], ShouldEqual, "b")
		})
	})

}
