create keyspace capapi with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };
create table capapi.capability( api_key varchar PRIMARY KEY, policy_set  blob);
create table capapi.delegation(api_key varchar PRIMARY KEY, delegations  list<varchar>);
