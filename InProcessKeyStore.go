package main

//InProcessKeyStore stores keys and their capabilities in process
type InProcessKeyStore struct {
	keys        map[string][]byte
	delegations map[string][]string
}

//Set takes a key and stores the scope against it in process
func (instance InProcessKeyStore) Set(key string, scope []byte) {
	instance.keys[key] = scope
}

//Delegate finds the key toe delegate and uses the state to create the root of the delegation
func (instance InProcessKeyStore) Delegate(key string, delegatedKey string, policySet PolicySet) error {
	bytes, err := instance.Get(key)
	CheckError(err)
	nextDelegation := PolicySetFromBytes(bytes)
	nextDelegation.SetDelegation(policySet)
	instance.Set(delegatedKey, nextDelegation.Bytes())

	if _, ok := instance.delegations[key]; ok {
		instance.delegations[key] = append(instance.delegations[key], delegatedKey)
	} else {
		instance.delegations[key] = []string{delegatedKey}
	}

	return nil
}

//Revoke removes the specified key from the
func (instance InProcessKeyStore) Revoke(key string) error {
	if keys, ok := instance.delegations[key]; ok {
		for _, key := range keys {
			instance.Revoke(key)
		}
	}

	delete(instance.keys, key)
	return nil
}

//Get returns the scope byte representation of the scope indexed by the key.
//If the key is not present in the map then an error is returned.
func (instance InProcessKeyStore) Get(key string) ([]byte, error) {
	if _, ok := instance.keys[key]; !ok {
		return []byte{}, ErrAPIKeyNotFound
	}
	return instance.keys[key], nil
}

//CreateInProcKeyStore create, initialize and return a new InProcessKeyStore.
func CreateInProcKeyStore() KeyStore {
	keyStore := InProcessKeyStore{
		keys:        map[string][]byte{},
		delegations: map[string][]string{},
	}

	return keyStore
}
