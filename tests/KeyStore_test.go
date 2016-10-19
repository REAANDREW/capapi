package tests

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/reaandrew/capapi/core"
	"github.com/reaandrew/capapi/infrastructure/inproc"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInProcessKeyStore(t *testing.T) {
	log.SetLevel(log.ErrorLevel)
	Convey("Returns", t, func() {
		keyStore := inproc.InProcessKeyStore{
			Keys: map[string][]byte{},
		}

		Convey("err when capability key is not found", func() {

			_, err := keyStore.Get("talula")

			So(err, ShouldEqual, core.ErrAPIKeyNotFound)
		})

		Convey("value when key is found", func() {

			var expectedKey = "fubar"
			var expectedValue = []byte{1, 2, 3}

			keyStore.Set(expectedKey, expectedValue)
			value, _ := keyStore.Get(expectedKey)

			So(value, ShouldResemble, expectedValue)
		})
	})
}
