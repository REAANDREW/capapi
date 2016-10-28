package main

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/gocql/gocql"
)

func createCassandraSession(clusterMembers []string) *gocql.Session {
	cluster := gocql.NewCluster(clusterMembers...)
	cluster.ProtoVersion = 2
	cluster.Keyspace = "capapi"
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	return session
}

func TestCassandraKeyStore(t *testing.T) {

	log.SetLevel(log.ErrorLevel)
	var clusterMembers = []string{"0.0.0.0:9042"}

	var cassandraSession = createCassandraSession(clusterMembers)
	defer cassandraSession.Close()

	cassandraSession.Query("create keyspace capapi with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };")
	cassandraSession.Query("create table capapi.capability( api_key varchar PRIMARY KEY, policy_set blob);")
	cassandraSession.Query("create table capapi.delegation(api_key varchar PRIMARY KEY, delegations list<varchar>);")

	Convey("CassandraKeyStore", t, func() {

		var cassandraKeyStore = NewCassandraKeyStore(clusterMembers)
		cassandraKeyStore.Start()
		defer cassandraKeyStore.Stop()

		Convey("Set", func() {
			key, _ := CreateKey()
			policySet := NewPolicySetBuilder().
				WithPolicy(NewPolicyBuilder().WithVerbs([]string{"GET", "POST", "PUT"})).BuildPolicySet()
			policyBytes := policySet.Bytes()

			cassandraKeyStore.Set(key, policyBytes)

			var id string
			var policySetValue []byte

			if err := cassandraSession.Query(`SELECT api_key, policy_set FROM capability WHERE api_key = ? LIMIT 1`, key).Consistency(gocql.One).Scan(&id, &policySetValue); err != nil {
				log.Fatal(err)
			}

			So(id, ShouldEqual, key)
		})

		Convey("Get", func() {
			key, _ := CreateKey()
			policySet := NewPolicySetBuilder().
				WithPolicy(NewPolicyBuilder().WithVerbs([]string{"GET", "POST", "PUT"})).BuildPolicySet()
			policyBytes := policySet.Bytes()

			if err := cassandraSession.Query(`INSERT INTO capability (api_key,policy_set) VALUES (?, ?)`, key, policyBytes).Exec(); err != nil {
				log.Fatal(err)
			}

			bytes, err := cassandraKeyStore.Get(key)

			So(err, ShouldBeNil)
			So(bytes, ShouldResemble, policyBytes)
		})
	})

	Convey("Testing my knowledge of the GOCQL driver for Cassandra", t, func() {

		cluster := gocql.NewCluster("0.0.0.0:9042")
		cluster.ProtoVersion = 2
		cluster.Keyspace = "capapi"
		cluster.Consistency = gocql.Quorum
		session, err := cluster.CreateSession()
		if err != nil {
			log.Fatal(err)
		}
		defer session.Close()

		Convey("Inserting a Policy Set", func() {
			key, _ := CreateKey()
			policySet := NewPolicySetBuilder().
				WithPolicy(NewPolicyBuilder().WithVerbs([]string{"GET", "POST", "PUT"})).BuildPolicySet()
			policyBytes := policySet.Bytes()

			if err := session.Query(`INSERT INTO capability (api_key,policy_set) VALUES (?, ?)`, key, policyBytes).Exec(); err != nil {
				log.Fatal(err)
			}

			var id string
			var policySetValue []byte

			if err := session.Query(`SELECT api_key, policy_set FROM capability WHERE api_key = ? LIMIT 1`, key).Consistency(gocql.One).Scan(&id, &policySetValue); err != nil {
				log.Fatal(err)
			}

			So(id, ShouldEqual, key)
		})

		Convey("Inserting and updating a list of delegations", func() {
			key1, _ := CreateKey()
			delegatedKey1, _ := CreateKey()
			delegatedKey2, _ := CreateKey()

			if err := session.Query(`UPDATE delegation SET delegations = delegations + ? WHERE api_key = ?`, []string{delegatedKey1}, key1).Exec(); err != nil {
				log.Fatal(err)
			}

			if err := session.Query(`UPDATE delegation SET delegations = delegations + ? WHERE api_key = ?`, []string{delegatedKey2}, key1).Exec(); err != nil {
				log.Fatal(err)
			}

			var apiKey string
			var delegations []string

			if err := session.Query(`SELECT api_key, delegations FROM delegation WHERE api_key = ? LIMIT 1`, key1).Consistency(gocql.One).Scan(&apiKey, &delegations); err != nil {
				log.Fatal(err)
			}
		})

	})

}
