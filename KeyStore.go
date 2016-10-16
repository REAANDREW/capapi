package main

type keyStore interface {
	Set(key string, scope []byte)
	Get(key string) ([]byte, error)
}

type inProcessKeyStore struct {
	keys map[string][]byte
}

func (instance inProcessKeyStore) Set(key string, scope []byte) {
	instance.keys[key] = scope
}

func (instance inProcessKeyStore) Get(key string) ([]byte, error) {
	if _, ok := instance.keys[key]; !ok {
		return []byte{}, errAPIKeyNotFound
	}
	return instance.keys[key], nil
}
