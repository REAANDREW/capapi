package main

type keyStore interface {
	Set(key string, scope []byte)
	Get(key string) []byte
}

type inProcessKeyStore struct {
	keys map[string][]byte
}

func (instance inProcessKeyStore) Set(key string, scope []byte) {
	instance.keys[key] = scope
}

func (instance inProcessKeyStore) Get(key string) []byte {
	return instance.keys[key]
}
