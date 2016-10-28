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

//Delegate finds the key toe delegate and uses the state to create the root of the delegation
func (instance *CassandraKeyStore) Delegate(key string, delegatedKey string, policySet PolicySet) error {
	return nil
}

//Revoke removes the specified key from the
func (instance *CassandraKeyStore) Revoke(key string) error {
	return nil
}

//Get returns the scope byte representation of the scope indexed by the key.
//If the key is not present in the map then an error is returned.
func (instance *CassandraKeyStore) Get(key string) ([]byte, error) {
	return []byte{}, nil
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
