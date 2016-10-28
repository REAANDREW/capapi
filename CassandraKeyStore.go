package main

import (
	"github.com/gocql/gocql"
)

//NewCassandraKeyStore creates and returns a new CassandraKeyStore
func NewCassandraKeyStore(clusterMembers []string) *CassandraKeyStore {
	return &CassandraKeyStore{
		clusterMembers: clusterMembers,
	}
}

//CassandraKeyStore is a type which enables the use of a Cassandra as a KeyStore
type CassandraKeyStore struct {
	clusterMembers []string
	session        *gocql.Session
}

//Set takes a key and stores the scope against it in process
func (instance *CassandraKeyStore) Set(key string, policySetBytes []byte) {
	if err := instance.session.Query(`INSERT INTO capability (api_key,policy_set) VALUES (?, ?)`, key, policySetBytes).Exec(); err != nil {
		CheckError(err)
	}
}

//AddDelegation adds the delegated key to the list of the parent
func (instance *CassandraKeyStore) AddDelegation(parent string, delegationKey string) error {
	if err := instance.session.Query(`UPDATE delegation SET delegations = delegations + ? WHERE api_key = ?`, []string{delegationKey}, parent).Exec(); err != nil {
		return err
	}
	return nil
}

//GetDelegations returns the keys which are delegations of the parent
func (instance *CassandraKeyStore) GetDelegations(parent string) ([]string, error) {
	var delegations []string

	if err := instance.session.Query(`SELECT delegations FROM delegation WHERE api_key = ? LIMIT 1`, parent).Consistency(gocql.One).Scan(&delegations); err != nil {
		if err == gocql.ErrNotFound {
			return []string{}, nil
		}
		return []string{}, err
	}

	return delegations, nil
}

//Delegate finds the key toe delegate and uses the state to create the root of the delegation
func (instance *CassandraKeyStore) Delegate(key string, delegatedKey string, policySet PolicySet) error {
	bytes, err := instance.Get(key)
	if err != nil {
		return err
	}

	nextDelegation := PolicySetFromBytes(bytes)
	nextDelegation.SetDelegation(policySet)
	instance.Set(delegatedKey, nextDelegation.Bytes())

	return instance.AddDelegation(key, delegatedKey)
}

//Revoke removes the specified key from the
func (instance *CassandraKeyStore) Revoke(key string) error {

	delegations, err := instance.GetDelegations(key)
	if err != nil {
		return err
	}

	for _, key := range delegations {
		instance.Revoke(key)
	}

	if err := instance.session.Query(`DELETE FROM capability WHERE api_key = ?`, key).Exec(); err != nil {
		return err
	}
	if err := instance.session.Query(`DELETE FROM delegation WHERE api_key = ?`, key).Exec(); err != nil {
		return err
	}

	return nil
}

//Get returns the scope byte representation of the scope indexed by the key.
//If the key is not present in the map then an error is returned.
func (instance *CassandraKeyStore) Get(key string) ([]byte, error) {
	var id string
	var policySetValue []byte

	if err := instance.session.Query(`SELECT api_key, policy_set FROM capability WHERE api_key = ? LIMIT 1`, key).Consistency(gocql.One).Scan(&id, &policySetValue); err != nil {
		if err == gocql.ErrNotFound {
			return []byte{}, ErrAPIKeyNotFound
		}
		return []byte{}, err
	}

	return policySetValue, nil
}

//Start connects to the Cassandra DB and creates a session which the CassandraKeyStore can use
func (instance *CassandraKeyStore) Start() error {
	cluster := gocql.NewCluster(instance.clusterMembers...)
	cluster.ProtoVersion = 2
	cluster.Keyspace = "capapi"
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	instance.session = session
	return err
}

//Stop closes the session to the Cassandra DB
func (instance *CassandraKeyStore) Stop() {
	instance.session.Close()
}
