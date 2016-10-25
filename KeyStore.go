package main

//KeyStore is the interface for storing and retrieving a PolicySet by a key.
type KeyStore interface {
	Set(key string, scope []byte)
	Get(key string) ([]byte, error)
	Delegate(key string, delegatedKey string, policySet PolicySet) error
	Revoke(key string) error
}
