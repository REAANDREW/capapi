package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	. "github.com/onsi/gomega"
	capnp "zombiezen.com/go/capnproto2"
)

/*

Due to composition it simply means that given an API Key, when it is delegated, the parent scope is always evaluated first
therefor when the new scope is evaluated it must be further defined than the parent otherwise it would not get evaluated

WIN, WIN, WIN, WIN!!

*/

const key = "unsecure_key_number_1"

func CreateKeyStore() keyStore {
	msg, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

	policySet, _ := NewRootPolicySet(seg)

	policyList, _ := NewPolicy_List(seg, 1)

	policy, _ := NewPolicy(seg)

	textList, _ := capnp.NewTextList(seg, 0)

	policy.SetVerbs(textList)

	policyList.Set(0, policy)

	policySet.SetPolicies(policyList)

	byteValue, _ := msg.Marshal()

	keyStore := inProcessKeyStore{
		keys: map[string][]byte{
			key: byteValue,
		},
	}

	return keyStore
}

func CreateFakeEndpoint() *httptest.Server {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprintln(w, "You Made It!")
	}))
	return server
}

type SystemUnderTest struct {
	APIGateway      apiSecurityGateway
	APIGatewayProxy *httptest.Server
	FakeEndpoint    *httptest.Server
	KeyStore        keyStore
	ResponseBody    string
	ResponseCode    int
	ServerListener  net.Listener
}

func CreateSystemUnderTest(keyStore keyStore) *SystemUnderTest {
	instance := &SystemUnderTest{}

	instance.KeyStore = keyStore

	instance.FakeEndpoint = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//var expectedResponseBody = "You Made It Baby, Yeh!"
		var expectedResponseBody = instance.ResponseBody
		w.WriteHeader(instance.ResponseCode)
		fmt.Fprintln(w, expectedResponseBody)
	}))

	var gatewayProxy = apiSecurityGatewayProxy{
		upStream: ":12345",
	}

	instance.APIGatewayProxy = httptest.NewUnstartedServer(gatewayProxy.handler())

	return instance
}
func (instance *SystemUnderTest) setResponseBody(value string) {
	instance.ResponseBody = value
}

func (instance *SystemUnderTest) setResponseCode(value int) {
	instance.ResponseCode = value
}

func (instance *SystemUnderTest) start() {
	instance.FakeEndpoint.Start()
	instance.APIGatewayProxy.Start()

	serverListener, err := net.Listen("tcp", ":12345")
	instance.ServerListener = serverListener

	checkError(err)

	upStreamURL, _ := url.Parse(instance.FakeEndpoint.URL)
	var gateway = apiSecurityGateway{
		upStream: *upStreamURL,
		keyStore: instance.KeyStore,
	}
	go gateway.start(serverListener)
}

func (instance *SystemUnderTest) stop() {
	instance.FakeEndpoint.Close()
	instance.APIGatewayProxy.Close()
	instance.ServerListener.Close()
	time.Sleep(1 * time.Millisecond)
}

type PolicyBuilder struct {
	Path    string
	Verbs   []string
	Headers map[string][]string
	Query   map[string][]string
}

func newPolicyBuilder() PolicyBuilder {
	return PolicyBuilder{
		Verbs:   []string{},
		Headers: map[string][]string{},
		Query:   map[string][]string{},
	}
}

func (instance PolicyBuilder) withVerb(verb string) PolicyBuilder {
	return PolicyBuilder{
		Path:    instance.Path,
		Verbs:   append(instance.Verbs, verb),
		Headers: instance.Headers,
		Query:   instance.Query,
	}
}

func (instance PolicyBuilder) build(seg *capnp.Segment) Policy {
	policy, _ := NewPolicy(seg)

	verbList, _ := capnp.NewTextList(seg, int32(len(instance.Verbs)))
	for i := 0; i < len(instance.Verbs); i++ {
		verbList.Set(i, instance.Verbs[i])
	}
	policy.SetVerbs(verbList)

	headerList, _ := NewKeyValuePolicy_List(seg, int32(len(instance.Headers)))
	count := 0
	for key, value := range instance.Headers {
		headerPolicy, _ := NewKeyValuePolicy(seg)
		headerPolicy.SetKey(key)

		headerValueList, _ := capnp.NewTextList(seg, int32(len(value)))
		for i := 0; i < len(value); i++ {
			headerValueList.Set(i, value[i])
		}
		headerPolicy.SetValues(headerValueList)

		headerList.Set(count, headerPolicy)
		count++
	}
	policy.SetHeaders(headerList)

	queryList, _ := NewKeyValuePolicy_List(seg, int32(len(instance.Query)))
	count = 0
	for key, value := range instance.Query {
		queryPolicy, _ := NewKeyValuePolicy(seg)
		queryPolicy.SetKey(key)

		queryValueList, _ := capnp.NewTextList(seg, int32(len(value)))
		for i := 0; i < len(value); i++ {
			queryValueList.Set(i, value[i])
		}
		queryPolicy.SetValues(queryValueList)

		queryList.Set(count, queryPolicy)
		count++
	}
	policy.SetQuery(queryList)

	return policy
}

type PolicySetBuilder struct {
	PolicyBuilders []PolicyBuilder
}

func (instance PolicySetBuilder) withPolicy(builder PolicyBuilder) PolicySetBuilder {
	return PolicySetBuilder{
		PolicyBuilders: append(instance.PolicyBuilders, builder),
	}
}

func (instance PolicySetBuilder) build() (string, []byte) {
	msg, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))

	policySet, _ := NewRootPolicySet(seg)
	policyList, _ := NewPolicy_List(seg, int32(len(instance.PolicyBuilders)))

	for i := 0; i < len(instance.PolicyBuilders); i++ {
		policy := instance.PolicyBuilders[i].build(seg)
		policyList.Set(i, policy)
	}

	policySet.SetPolicies(policyList)

	byteValue, _ := msg.Marshal()
	keyBytes := make([]byte, 64)
	_, err := rand.Read(keyBytes)
	checkError(err)

	key := base64.StdEncoding.EncodeToString(keyBytes)
	log.WithFields(log.Fields{
		"key": key,
	}).Info("Key Generated")
	return key, byteValue
}

func newPolicySetBuilder() PolicySetBuilder {
	return PolicySetBuilder{
		PolicyBuilders: []PolicyBuilder{},
	}
}

func TestWithUnRestrictedAccess(t *testing.T) {

	RegisterTestingT(t)

	var keystore = CreateKeyStore()
	var sut = CreateSystemUnderTest(keystore)
	var expectedResponseBody = "You Made It Baby, Yeh!"
	var expectedResponseCode = 200

	sut.setResponseBody(expectedResponseBody)
	sut.setResponseCode(expectedResponseCode)
	defer sut.stop()
	sut.start()
	client := &http.Client{}
	req, _ := http.NewRequest("GET", sut.APIGatewayProxy.URL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	Expect(resp.StatusCode).To(Equal(expectedResponseCode))
	Expect(strings.Trim(string(body), "\n")).To(Equal(expectedResponseBody))
}

func TestWithRestrictedAccessToSingleVerb(t *testing.T) {

	log.SetLevel(log.DebugLevel)
	RegisterTestingT(t)

	var keystore = CreateKeyStore()
	var sut = CreateSystemUnderTest(keystore)
	var expectedResponseBody = "You Made It Baby, Yeh!"
	var expectedResponseCode = 200

	sut.setResponseBody(expectedResponseBody)
	sut.setResponseCode(expectedResponseCode)
	defer sut.stop()
	sut.start()

	key, bytes := newPolicySetBuilder().
		withPolicy(newPolicyBuilder().withVerb("PUT")).
		build()

	keystore.Set(key, bytes)

	client := &http.Client{}
	log.WithFields(log.Fields{
		"url": sut.APIGatewayProxy.URL,
	}).Info("API Gateway Proxy URL")
	req, _ := http.NewRequest("PUT", sut.APIGatewayProxy.URL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	Expect(resp.StatusCode).To(Equal(expectedResponseCode))
	Expect(strings.Trim(string(body), "\n")).To(Equal(expectedResponseBody))
}
