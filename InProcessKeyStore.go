package main

//InProcessKeyStore stores keys and their capabilities in process
type InProcessKeyStore struct {
	Keys map[string][]byte
}

//Set takes a key and stores the scope against it in process
func (instance InProcessKeyStore) Set(key string, scope []byte) {
	instance.Keys[key] = scope
}

//Get returns the scope byte representation of the scope indexed by the key.
//If the key is not present in the map then an error is returned.
func (instance InProcessKeyStore) Get(key string) ([]byte, error) {
	if _, ok := instance.Keys[key]; !ok {
		return []byte{}, ErrAPIKeyNotFound
	}
	return instance.Keys[key], nil
}

//CreateInProcKeyStore create, initialize and return a new InProcessKeyStore.
func CreateInProcKeyStore() KeyStore {
	keyStore := InProcessKeyStore{
		Keys: map[string][]byte{},
	}

	return keyStore
}
