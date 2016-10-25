package main

import (
	"strings"

	capnp "zombiezen.com/go/capnproto2"
)

//EXTENSION METHODS FOR THE GENERATED CODE!

//TODO: KeyValue type has array of values not a single value! The type KeyValue should have an array of values not a single value.  The use of this as shown below is to take the list of values and join them into a single value.  This is not good and needs to change.

//KeyValueListFromMap creates a new KeyValue_List from a map
//TODO: Instead of checking the errors let them propogate up the chain
func KeyValueListFromMap(input map[string][]string) (KeyValue_List, error) {
	_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	CheckError(err)

	keyValueList, err := NewKeyValue_List(seg, int32(len(input)))
	CheckError(err)

	count := 0
	for key, value := range input {
		keyValuePolicy, err := NewKeyValue(seg)
		CheckError(err)

		keyValuePolicy.SetKey(key)
		keyValuePolicy.SetValue(strings.Join(value, ","))
		keyValueList.Set(count, keyValuePolicy)
		count++
	}

	return keyValueList, nil
}
