package main

type InProcessKeyStore struct {
	Keys map[string][]byte
}

func (instance InProcessKeyStore) Set(key string, scope []byte) {
	instance.Keys[key] = scope
}

func (instance InProcessKeyStore) Get(key string) ([]byte, error) {
	if _, ok := instance.Keys[key]; !ok {
		return []byte{}, ErrAPIKeyNotFound
	}
	return instance.Keys[key], nil
}

func CreateInProcKeyStore() KeyStore {
	keyStore := InProcessKeyStore{
		Keys: map[string][]byte{},
	}

	return keyStore
}
