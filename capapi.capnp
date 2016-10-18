using Go = import "/go.capnp";
@0x9041cfd31a197b3f;
$Go.package("main");
$Go.import("github.com/reaandrew/cap-based-subscription");

struct HTTPRequest {
    path @0 :Text;
    # The path of the http request

    verb @1 :Text;
    # The verb of the http request

    headers @2 :List(KeyValue);

    query @3 :List(KeyValue);
}

struct KeyValue{
    key @0 :Text;

    value @1 :Text;
}

struct HTTPResponse {
    body @0 :Text;
    status @1 :UInt32;
}

struct KeyValuePolicy{
    key @0 :Text;
    values @1 :List(Text);
}

struct Policy {
    path @0 :Text;

    verbs @1 :List(Text);

    headers @2 :List(KeyValuePolicy);

    query @3 :List(KeyValuePolicy);
}

struct PolicySet{
    policies @0 :List(Policy);
}

struct APIKey {
    value @0 :Text;
}

interface HTTPProxyFactory {
    getHTTPProxy @0 (key :APIKey) -> (proxy :HTTPProxy);
}

interface HTTPProxy {
    request @0 (requestObj :HTTPRequest) -> (response :HTTPResponse);
    delegate @1 (scope :PolicySet) -> (key :APIKey);
    revoke @2 () -> (result :Bool);
}
