package main

import (
	"net/http"
	"testing"

	capnp "zombiezen.com/go/capnproto2"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPolicy(t *testing.T) {

	log.SetLevel(log.ErrorLevel)

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

					So(policy.Validate(request), ShouldEqual, false)
				})
				Convey("when the request path does not match the policy templated path", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policy, _ := NewPolicy(seg)

					policy.SetPath("/some/path/{id:(1|2)}")

					request, _ := NewHTTPRequest(seg)

					request.SetPath("/some/path/3")

					result := policy.Validate(request)
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

					So(policy.Validate(request), ShouldEqual, false)
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

					So(policy.Validate(request), ShouldEqual, false)
				})

				Convey("when the request has a header value which does not match a header policy value for the specified key", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policy, _ := NewPolicy(seg)

					headerList, _ := NewKeyValuePolicy_List(seg, 1)
					keyValuePolicy, _ := NewKeyValuePolicy(seg)
					keyValuePolicy.SetKey("X-Client")

					valuesList, _ := capnp.NewTextList(seg, 3)
					valuesList.Set(0, "1")
					valuesList.Set(0, "2")
					valuesList.Set(0, "3")

					keyValuePolicy.SetValues(valuesList)

					headerList.Set(0, keyValuePolicy)

					policy.SetHeaders(headerList)

					request, _ := NewHTTPRequest(seg)

					headers, _ := NewKeyValue_List(seg, 1)
					keyValue, _ := NewKeyValue(seg)
					keyValue.SetKey("X-Client")
					keyValue.SetValue("4")
					headers.Set(0, keyValue)

					request.SetHeaders(headers)

					So(policy.Validate(request), ShouldEqual, false)
				})
				Convey("when the request has a query string key which is not present in the query policy", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policy, _ := NewPolicy(seg)

					queryList, _ := NewKeyValuePolicy_List(seg, 1)
					keyValuePolicy, _ := NewKeyValuePolicy(seg)
					keyValuePolicy.SetKey("A")
					queryList.Set(0, keyValuePolicy)

					policy.SetQuery(queryList)

					request, _ := NewHTTPRequest(seg)

					query, _ := NewKeyValue_List(seg, 1)
					keyValue, _ := NewKeyValue(seg)
					keyValue.SetKey("B")
					keyValue.SetValue("1")
					query.Set(0, keyValue)

					request.SetQuery(query)

					So(policy.Validate(request), ShouldEqual, false)
				})
				Convey("when the request has a query value which does not match a query policy value for the specified key", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policy, _ := NewPolicy(seg)

					queryList, _ := NewKeyValuePolicy_List(seg, 1)
					keyValuePolicy, _ := NewKeyValuePolicy(seg)
					keyValuePolicy.SetKey("X-Client")

					valuesList, _ := capnp.NewTextList(seg, 3)
					valuesList.Set(0, "1")
					valuesList.Set(0, "2")
					valuesList.Set(0, "3")

					keyValuePolicy.SetValues(valuesList)

					queryList.Set(0, keyValuePolicy)

					policy.SetQuery(queryList)

					request, _ := NewHTTPRequest(seg)

					query, _ := NewKeyValue_List(seg, 1)
					keyValue, _ := NewKeyValue(seg)
					keyValue.SetKey("X-Client")
					keyValue.SetValue("4")
					query.Set(0, keyValue)

					request.SetQuery(query)

					So(policy.Validate(request), ShouldEqual, false)
				})
			})

			Convey("true", func() {
				Convey("when the request path does match the policy exact path", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policy, _ := NewPolicy(seg)

					policy.SetPath("/some/path")

					request, _ := NewHTTPRequest(seg)

					request.SetPath("/some/path")

					So(policy.Validate(request), ShouldEqual, true)
				})
				Convey("when the request path does match the policy templated path", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policy, _ := NewPolicy(seg)

					policy.SetPath("/some/path/{id:(1|2)}")

					request, _ := NewHTTPRequest(seg)

					request.SetPath("/some/path/1")

					result := policy.Validate(request)
					So(result, ShouldEqual, true)
				})
				Convey("when the request verb does match a single policy verb", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policy, _ := NewPolicy(seg)

					verbs, _ := capnp.NewTextList(seg, 1)

					verbs.Set(0, "PUT")

					policy.SetVerbs(verbs)

					request, _ := NewHTTPRequest(seg)

					request.SetVerb("PUT")

					So(policy.Validate(request), ShouldEqual, true)
				})
				Convey("when the request verb does match one of many policy verbs", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policy, _ := NewPolicy(seg)

					verbs, _ := capnp.NewTextList(seg, 2)

					verbs.Set(0, "PUT")
					verbs.Set(1, "POST")

					policy.SetVerbs(verbs)

					request, _ := NewHTTPRequest(seg)

					request.SetVerb("PUT")

					So(policy.Validate(request), ShouldEqual, true)
				})

				Convey("when the request verb is the same as the policy verb but is not the same case", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policy, _ := NewPolicy(seg)

					verbs, _ := capnp.NewTextList(seg, 1)

					verbs.Set(0, "GET")

					policy.SetVerbs(verbs)

					request, _ := NewHTTPRequest(seg)

					request.SetVerb("get")

					So(policy.Validate(request), ShouldEqual, true)
				})
				Convey("when the request has a header key which is present in the header policy", func() {
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
					keyValue.SetKey("A")
					headers.Set(0, keyValue)

					request.SetHeaders(headers)

					So(policy.Validate(request), ShouldEqual, true)
				})
				Convey("when the request has a header value which does match a header policy value for the specified key", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policy, _ := NewPolicy(seg)

					headerList, _ := NewKeyValuePolicy_List(seg, 1)
					keyValuePolicy, _ := NewKeyValuePolicy(seg)
					keyValuePolicy.SetKey("X-Client")

					valuesList, _ := capnp.NewTextList(seg, 3)
					valuesList.Set(0, "1")
					valuesList.Set(0, "2")
					valuesList.Set(0, "3")

					keyValuePolicy.SetValues(valuesList)

					headerList.Set(0, keyValuePolicy)

					policy.SetHeaders(headerList)

					request, _ := NewHTTPRequest(seg)

					headers, _ := NewKeyValue_List(seg, 1)
					keyValue, _ := NewKeyValue(seg)
					keyValue.SetKey("X-Client")
					keyValue.SetValue("3")
					headers.Set(0, keyValue)

					request.SetHeaders(headers)

					So(policy.Validate(request), ShouldEqual, true)
				})
				Convey("when the request has a query string key which is present in the query policy", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policy, _ := NewPolicy(seg)

					queryList, _ := NewKeyValuePolicy_List(seg, 1)
					keyValuePolicy, _ := NewKeyValuePolicy(seg)
					keyValuePolicy.SetKey("A")
					queryList.Set(0, keyValuePolicy)

					policy.SetQuery(queryList)

					request, _ := NewHTTPRequest(seg)

					query, _ := NewKeyValue_List(seg, 1)
					keyValue, _ := NewKeyValue(seg)
					keyValue.SetKey("A")
					keyValue.SetValue("1")
					query.Set(0, keyValue)

					request.SetQuery(query)

					So(policy.Validate(request), ShouldEqual, true)
				})
				Convey("when the request has a query value which does match a query policy value for the specified key", func() {
					_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

					policy, _ := NewPolicy(seg)

					queryList, _ := NewKeyValuePolicy_List(seg, 1)
					keyValuePolicy, _ := NewKeyValuePolicy(seg)
					keyValuePolicy.SetKey("X-Client")

					valuesList, _ := capnp.NewTextList(seg, 3)
					valuesList.Set(0, "1")
					valuesList.Set(0, "2")
					valuesList.Set(0, "3")

					keyValuePolicy.SetValues(valuesList)

					queryList.Set(0, keyValuePolicy)

					policy.SetQuery(queryList)

					request, _ := NewHTTPRequest(seg)

					query, _ := NewKeyValue_List(seg, 1)
					keyValue, _ := NewKeyValue(seg)
					keyValue.SetKey("X-Client")
					keyValue.SetValue("3")
					query.Set(0, keyValue)

					request.SetQuery(query)

					So(policy.Validate(request), ShouldEqual, true)
				})
			})
		})
	})
}
