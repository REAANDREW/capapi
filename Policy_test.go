package main

import (
	"net/http"
	"testing"

	capnp "zombiezen.com/go/capnproto2"

	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPolicy(t *testing.T) {

	Convey("Policy", t, func() {

		Convey("test my knowledge of mux", func() {
			r := mux.NewRouter()
			r.Path("/something/{(1|2)}")

			req, _ := http.NewRequest("GET", "http://localhost:1234/something/1", nil)

			var routeMatch mux.RouteMatch

			result := r.Match(req, &routeMatch)

			So(result, ShouldEqual, true)
		})

		Convey("validates", func() {

			Convey("false", func() {
				Convey("when the request path does not match the policy exact path", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policy, _ := NewPolicy(seg)

					policy.SetPath("/some/path")

					request, _ := NewHTTPRequest(seg)

					request.SetPath("/some/other/path")

					So(policy.validate(request), ShouldEqual, false)

				})
				Convey("when the request path does not match the policy templated path", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policy, _ := NewPolicy(seg)

					policy.SetPath("/some/path/{id:(1|2)}")

					request, _ := NewHTTPRequest(seg)

					request.SetPath("/some/path/3")

					result := policy.validate(request)
					So(result, ShouldEqual, false)

				})
				Convey("when the request verb does not match a single policy verb", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policy, _ := NewPolicy(seg)

					verbs, _ := capnp.NewTextList(seg, 1)

					verbs.Set(0, "PUT")

					policy.SetVerbs(verbs)

					request, _ := NewHTTPRequest(seg)

					request.SetVerb("GET")

					So(policy.validate(request), ShouldEqual, false)
				})
				Convey("when the request verb does not match a any policy verbs", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policy, _ := NewPolicy(seg)

					verbs, _ := capnp.NewTextList(seg, 2)

					verbs.Set(0, "PUT")
					verbs.Set(1, "POST")

					policy.SetVerbs(verbs)

					request, _ := NewHTTPRequest(seg)

					request.SetVerb("GET")

					So(policy.validate(request), ShouldEqual, false)
				})
				Convey("when the request has a header key which is not present in the header policy", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policy, _ := NewPolicy(seg)

					headerList, _ := NewKeyValuePolicy_List(seg, 1)
					keyValuePolicy, _ := NewKeyValuePolicy(seg)
					keyValuePolicy.SetKey("A")
					headerList.Set(0, keyValuePolicy)

					policy.SetHeaders(headerList)

					request, _ := NewHTTPRequest(seg)

					headers, _ := NewKeyValue_List(seg, 1)
					keyValue, _ := NewKeyValue(seg)
					keyValue.SetKey("B")
					headers.Set(0, keyValue)

					request.SetHeaders(headers)

					So(policy.validate(request), ShouldEqual, false)
				})
				Convey("when the request has a header value which does not match a header policy value for the specified key", func() {

				})
				Convey("when the request has a query string key which is not present in the query policy", func() {

				})
				Convey("when the request has a query value which does not match a query policy value for the specified key", func() {

				})
			})

			Convey("true", func() {
				Convey("when the request path does match the policy exact path", func() {

				})
				Convey("when the request path does match the policy templated path", func() {

				})
				Convey("when the request verb does match a single policy verb", func() {

				})
				Convey("when the request verb does match one of many policy verbs", func() {

				})
				Convey("when the request has a header key which is present in the header policy", func() {

				})
				Convey("when the request has a header value which does match a header policy value for the specified key", func() {

				})
				Convey("when the request has a query string key which is present in the query policy", func() {

				})
				Convey("when the request has a query value which does match a query policy value for the specified key", func() {

				})
			})

		})

	})

}
