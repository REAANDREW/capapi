# CAPAPI

[![Build Status](https://travis-ci.org/REAANDREW/capapi.svg?branch=master)](https://travis-ci.org/REAANDREW/capapi)
[![Go Report Card](https://goreportcard.com/badge/github.com/REAANDREW/capapi)](https://goreportcard.com/report/github.com/REAANDREW/capapi)

Capapi is a object capability based HTTP API Security Gateway.  

![capapi](https://github.com/REAANDREW/capapi/blob/master/capapi.png)

## Getting Started

TBC

## Rough Write Up

- A KeyStore has a list of PolicySets indexed by an APIKey
- An APIKey is an unguessable 512 bit base64 encoded string
- A PolicySet is :
    - Serializable
    - Executable
    - Immutable
    - Composable
- A PolicySet can be:
    - Delegated
    - Revoked
- A PolicySet uses HTTP as its langauge to construct a capability e.g.
    - Path
    - Verb
    - Headers
    - QueryString
- A PolicySet can hold one or many policies and each policy has:
    - A permitted path either exact (/some/path) or templated (/some/path/{id:(1|2)}
    - A permitted list of verbs e.g. GET, PUT, PATCH etc...
    - A permitted list of header keys and optional list of permitted values for the specified key
    - A permitted list of querystring keys and optional list of permitted values for the specified key
- When a PolicySet is Delegated, during execution the top level parent capability will be evaluated first meaning that each delegation has to have the same or less capability scope than its parent capability
- When a PolicySet is Revoked all derived PolicySets are also revoked
- When a PolicySet is Delegated, it is not necessary to check a delegation as during exeuction it will always be evalutated after its parent.
- A Delegation or Revocation has to be executed by the Gateway.
- If any of the following conditions are true then the request is not authorized and will not be executed:
    - A path which is not supported by any policy in the set
    - A verbs which is not supported by any policy in the set
    - A header key which is not supported by any policy in the set
    - A header value which is not supported by any policy in the set for the specified key
    - A querystring key which is not supported by any policy in the set
    - A querystring value which is not supported by any policy in the set for the specified key
 

To ensure that the caller only has a reference to a Proxy, this project uses the [Cap'N Proto](https://capnproto.org) library to serialize the PolicySet using a Type System.

> WHAT ARE OBJECT CAPABILITIES?
>
> A capability is:
>
> - an unguessable,
> - communicable,
> - token of authority
> - which references an object
> - and a set of access rights.
>
> DESIGNING SECURE SYSTEMS
>
> WITH OBJECT-CAPABILITIES, PYTHON, AND CAP'N PROTO
>
> Drew Fisher
>
> https://smpfle21zb7r5nnat5uq.oasis.sandstorm.io/index.html#/
>
